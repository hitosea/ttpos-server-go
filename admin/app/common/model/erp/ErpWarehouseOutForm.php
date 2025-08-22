<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use app\common\model\product\Material;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\ProductBom;
use app\shop\model\product\RelatedMaterial;
use think\facade\Db;

/**
 * 库存记录模型
 */
class ErpWarehouseOutForm extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_out_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;

    /**
     * inventory_type 类型 1-入库 2-出库
     */
    const INVENTORY_TYPE_IN = 1;
    const INVENTORY_TYPE_OUT = 2;

    /**
     * type 操作类型 30-销售出库 40-调整出库 41-删除出库
     */
    const TYPE_SALE_OUT = 30;
    const TYPE_ADJUST_OUT = 40;
    const TYPE_ADJUST_OUT_DEL = 41;

    /**
     * status 状态 10-已入库 20-已出库 30-已撤销
     */

    const STATUS_IN = 10;
    const STATUS_OUT = 20;
    const STATUS_REVOKED = 30;

    /**
     * 库存数量
     */
    public function getNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_uuid', 'uuid');
    }

    /**
     * 关联出库记录详情
     */
    public function erpWarehouseOutFormItem()
    {
        return $this->hasOne(ErpWarehouseOutFormItem::class, 'warehouse_out_form_uuid', 'uuid');
    }

    /**
     * 获取列表
     *
     * @param [type] $params
     * @return object
     */
    public function getList($params)
    {
        $startTime = isset($params['date'][0]) ? strtotime($params['date'][0]) : 0;
        $endTime = isset($params['date'][1]) ? strtotime($params['date'][1] . ' 23:59:59') : 0;

        $list = [];
        $model = (new self())->hasWhere('erpWarehouseOutFormItem', ['status' => 1])->with([
            'erpWarehouseOutFormItem' => function ($q) {
                return $q->with([
                    'productBom' => function ($q) {
                        return $q->withTrashed()->with([
                            'product' => function ($q) {
                                return $q->withTrashed();
                            },
                            'relatedMaterial'
                        ]);
                    },
                    'material' => function ($q) {
                        return $q->withTrashed()->with(['unit']);
                    }
                ]);
            },
            'operator',
        ]);

        // 操作类型 30-销售出库 40-调整出库 41-删除出库
        if (isset($params['type']) && $params['type']) {
            $model = $model->where('ErpWarehouseOutForm.scene', [
                30 => 0,
                40 => 1,
                41 => 4,
            ][$params['type']]);
        }

        // 起始时间
        if ($startTime && $endTime) {
            $model = $model->where('ErpWarehouseOutForm.update_time', 'between', [$startTime, $endTime]);
        }

        $paginate = $model->order('ErpWarehouseOutForm.update_time desc')->paginate($params);

        foreach ($paginate->items() as $item) {
            $product = [
                'product_name_text' => '',
                'product_unit_text' => '',
            ];
            $formItem = $item['erpWarehouseOutFormItem'];
            // 商品规格
            if ($formItem->productBom) {
                if ($formItem->productBom->product) {
                    $product['product_name_text'] = $formItem->productBom->product->product_name_text;
                }
                $product['product_unit_text'] = $formItem->productBom->spec_name_text;
            }

            // 材料
            if ($formItem->material) {
                $product['product_name_text'] = $formItem->material->product_name_text;
                if ($formItem->material->unit) {
                    $product['product_unit_text'] = $formItem->material->unit->unit_name_text;
                }
            }
            // 是否显示撤销按钮
            $isShowOutCancel = 1;
            if (($formItem->productBom && $formItem->productBom->relatedMaterial->count() > 0) || in_array($item['scene'], [0, 4])) {
                $isShowOutCancel = 0;
            }
            $list[] = [
                'id' => $item['uuid'],
                'number' => $item['form_no'],
                'type' => [
                    0 => self::TYPE_SALE_OUT,
                    1 => self::TYPE_ADJUST_OUT,
                    4 => self::TYPE_ADJUST_OUT_DEL,
                ][$item['scene']],
                'product' => $product,
                'num' => floatval($formItem['num']),
                'remark' => $item['remark'],
                'status' => [
                    0 => self::STATUS_OUT,
                    1 => self::STATUS_REVOKED,
                ][$item['status']],
                'operator' => [
                    'shop_user_id' => $item['operator']['shop_user_id'] ?? 0,
                    'user_name' => $item['operator']['username'] ?? '',
                    'real_name' => $item['operator']['real_name'] ?? '',
                ],
                'out_time' => strtotime($item['create_time']),
                'revoke_time' => $item['revoke_time'],
                'is_show_out_cancel' => $isShowOutCancel,
            ];
        }

        return [
            'current_page' => $paginate->currentPage(),
            'last_page' => $paginate->lastPage(),
            'per_page' => $paginate->listRows(),
            'total' => $paginate->total(),
            'data' => $list,
        ];
    }

    /**
     * 详情
     */
    public function detail($id)
    {
        $model = new self;
        $info = $model->with([
            'erpWarehouseOutFormItem' => [
                'productBom' => function ($q) {
                    return $q->withTrashed()->with([
                        'product' => function ($q) {
                            return $q->withTrashed();
                        },
                        'relatedMaterial'
                    ]);
                },
                'material' => function ($q) {
                    return $q->withTrashed()->with(['unit']);
                }
            ],
            'operator',
        ])->where('uuid', $id)->find();

        return $info;
    }

    /**
     * 撤销
     */
    public function cancel()
    {
        // 不是撤销状态的才能撤销
        if ($this->status == 1) {
            $this->error = '记录已撤销';
            return false;
        }
        $this->startTrans();
        try {
            // 如果是销售出库，不允许撤销
            if ($this->scene == 0) {
                $this->error = '销售出库不允许撤销';
                return false;
            }
            // 如果是删除出库，不允许撤销
            if ($this->scene == 4) {
                $this->error = '删除出库不允许撤销';
                return false;
            }
            $formItem = $this->erpWarehouseOutFormItem;
            if ($formItem->productBom) {
                if ($formItem->productBom->delete_time > 0) {
                    $this->error = '规格不存在，无法进行撤销操作';
                    return false;
                }
                if ($formItem->productBom->relatedMaterial->count() > 0) {
                    $this->error = '关联材料商品无法进行撤销操作';
                    return false;
                }
                // 回滚规格库存
                ProductBom::where('uuid', $formItem->productBom->uuid)->inc('stock_num', $formItem->num)->update();
            }

            // 回滚材料库存
            if ($formItem->material) {
                Material::where('uuid', $formItem->material->uuid)->inc('stock_num', $formItem->num)->update();
                // 更新材料关联的规格库存
                $relatedMaterialUuidList = RelatedMaterial::where('material_uuid', $formItem->material->uuid)->column('uuid') ?: [];
                RelatedMaterial::updateStock($relatedMaterialUuidList);
            }
            
            // 更新出库记录表
            $this->status = 1;
            $this->revoke_time = time();
            $this->save();
            // 更新出库明细表
            $formItem->revoke_time = time();
            $formItem->save();
            $this->commit();
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
        return $this->id;
    }

    /**
     * 删除
     */
    public function del()
    {
        // 撤销状态的才能删除
        if ($this->status != 1) {
            $this->error = '记录不能删除';
            return false;
        }
        return Db::transaction(function () {
            $this->erpWarehouseOutFormItem->delete();
            $this->delete();

            return true;
        });
    }

    /**
     * 出库编号：18位纯数字（前2位OO，2-10位是年月日，中间位是0000，后4位随机生成）
     *
     * @return string
     */
    public function generateOutCode()
    {
        $date = date('Ymd');
        $rand = rand(1000, 9999);
        $code = 'OO' . $date . '0000' . $rand;
        return $code;
    }

    /**
     * 添加出库记录
     */
    public function addOutForm($scene, $shopUserUuid, $data)
    {
        $res = $this->save([
            'form_no' => $this->generateOutCode(),
            'scene' => $scene,
            'operator_uuid' => $shopUserUuid,
            'remark' => $data['remark'] ?? '',
            'status' => $data['status'] ?? 0,
            'associated_order_uuid' => $data['associated_order_uuid'] ?? 0,
        ]);
        if (!$res) {
            return false;
        }
        $outFormItem = new ErpWarehouseOutFormItem();
        return $outFormItem->save([
            'warehouse_out_form_uuid' => $this->uuid,
            'product_bom_uuid' => $data['product_bom_uuid'] ?? 0,
            'material_uuid' => $data['material_uuid'] ?? 0,
            'num' => $data['num'] ?? 0,
            'scene' => $scene,
            'status' => 1,
            'reduce_stock' => 1,
        ]);
    }
}
