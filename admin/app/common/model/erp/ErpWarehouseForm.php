<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\ProductBom;
use app\common\model\product\Material;
use app\shop\model\product\RelatedMaterial;

/**
 * 库存记录模型
 */
class ErpWarehouseForm extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    protected $pk = 'id';
    protected $autoWriteTimestamp = true;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['number', 'in_time'];

    /**
     * type 操作类型 10-采购入库 20-调整入库 21-添加入库
     */
    const TYPE_PURCHASE_IN = 10;
    const TYPE_ADJUST_IN = 20;
    const TYPE_ADJUST_IN_ADD = 21;

    const OLD_TYPE = [
        10 => self::SCENE_PURCHASE_IN,
        20 => self::SCENE_ADJUST_IN,
        21 => self::SCENE_ADD_IN,
    ];

    const OLD_TYPE_MAP = [
        0 => self::TYPE_PURCHASE_IN,
        1 => self::TYPE_ADJUST_IN_ADD,
        2 => self::TYPE_ADJUST_IN,
    ];

    /**
     * scene 场景, 0-purchase采购入库 1-add添加入库 2-adjust调整入库
     */
    const SCENE_PURCHASE_IN = 0;
    const SCENE_ADD_IN = 1;
    const SCENE_ADJUST_IN = 2;

    /**
     * status 状态 0-success已入库 1-canceled已撤销
     */
    const STATUS_SUCCESS = 0;
    const STATUS_CANCELED = 1;

    const STATUS_MAP = [
        0 => 10,
        1 => 30,
    ];

    /**
     * 兼容字段
     */
    public function getNumberAttr($value, $data)
    {
        return $this->form_no;
    }
    public function getInTimeAttr($value, $data)
    {
        return $this->getData('create_time');
    }

    /**
     * 库存数量
     */
    public function getNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 关联采购单
     */
    public function purchaseOrder()
    {
        return $this->belongsTo(ErpPurchaseOrder::class, 'purchase_order_uuid', 'uuid');
    }

    /**
     * 关联产品SKU
     */
    public function productSku()
    {
        return $this->belongsTo(ProductBom::class, 'product_bom_uuid', 'uuid');
    }

    /**
     * 关联原料
     */
    public function material()
    {
        return $this->belongsTo(Material::class, 'material_uuid', 'uuid');
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_uuid', 'uuid')->field(['uuid', 'uuid as shop_user_id', 'username as user_name', 'real_name']);
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

        $model = (new self())->where('scene', '<>', 3);
        if (isset($params['name']) && $params['name']) {
            $model = $model->like('name', $params['name']);
        }

        // 操作类型 10-采购入库 20-调整入库 21-添加入库
        if (isset($params['type']) && $params['type']) {
            $model = $model->where('scene', self::OLD_TYPE[$params['type']]);
        }

        // 起始时间
        if ($startTime && $endTime) {
            $model = $model->where('create_time', 'between', [$startTime, $endTime]);
        }

        $paginate = $model->with([
            'purchaseOrder',
            'productSku' => [ 
                'product' => function($query) {
                    $query->WithTrashed();
                }
            ],
            'material',
            'operator',
        ])->order('create_time desc')->paginate($params);
        //
        $list = [];
        foreach ($paginate->items() as $item) {
            // 采购订单
            $purchaseOrder = [];
            if ($item['purchaseOrder']) {
                $purchaseOrder = [
                    'id' => $item['purchaseOrder']['uuid'],
                    'name' => $item['purchaseOrder']['name'],
                    'number' => $item['purchaseOrder']['form_no'],
                ];
            }
            // 商品
            $product = [];
            if ($item['productSku']) {
                $product = [
                    'product_name_text' => $item['productSku']['product']['product_name_text'],
                ];
            }
            if ($item['material']) {
                $product = [
                    'product_name_text' => $item['material']['product_name_text'],
                ];
            }
            // 规格
            $productSkuNameText = '';
            if ($item['productSku']) {
                $productSkuNameText = $item['productSku']['spec_name_text'];
            }
            // 操作人
            $operator = [];
            if ($item['operator']) {
                $operator = [
                    'shop_user_id' => $item['operator']['uuid'],
                    'user_name' => $item['operator']['username'],
                    'real_name' => $item['operator']['real_name'],
                ];
            }
            $list[] = [
                'id' => $item['uuid'],
                'number' => $item['form_no'],
                'type' => self::OLD_TYPE_MAP[$item['scene']],
                'purchaseOrder' => $purchaseOrder,
                'product' => $product,
                'product_sku_name_text' => $productSkuNameText,
                'num' => $item['num'],
                'remark' => $item['remark'],
                'status' => self::STATUS_MAP[$item['status']],
                'operator' => $operator,
                'create_time' => strtotime($item['create_time']),
                'in_time' => strtotime($item['create_time']),
                'revoke_time' => $item['revoke_time'],
                'is_show_in_cancel' => $this->isShowInCancel($item),
            ];
        }
        return [
            'current_page' => $paginate->currentPage(),
            'last_page' => $paginate->lastPage(),
            'total' => $paginate->total(),
            'per_page' => $paginate->listRows(),
            'data' => $list,
        ];
    }

    /**
     * 检查库存
     *
     * @param [type] $item
     * @param [type] $num
     * @return int
     */
    private function isShowInCancel($item)
    {
        if (!$item['productSku'] && !$item['material']) {
            return 0;
        }

        $stockNum = 0;
        // 产品规格库存
        if ($item['productSku']) {
            $stockNum = $item['productSku']['stock_num'];
        }
        // 材料库存
        if ($item['material']) {
            $stockNum = $item['material']['stock_num'];
        }

        return floatval($stockNum) >= floatval($item['num']) ? 1 : 0;
    }

    /**
     * 详情
     */
    public function detail($id)
    {
        $model = new self;
        $info = $model->with([
                'purchaseOrder',
                'productSku' => [ 'product' ],
                'material',
                'operator',
            ])
            ->where('uuid', $id)
            ->find();
        return $info;
    }

    /**
     * 撤销
     */
    public function cancel()
    {
        // 不是撤销状态的才能撤销
        if ($this->status == self::STATUS_CANCELED) {
            $this->error = '记录已撤销';
            return false;
        }
        $this->startTrans();
        try {
            // 撤销商品规格库存
            if ($this->product_bom_uuid > 0) {
                if (!$this->productSku) {
                    $this->error = '规格不存在，无法进行撤销操作';
                    return false;
                }
                if ($this->productSku->stock_num < $this->num) {
                    $this->error = '商品规格库存不足';
                    return false;
                }
                ProductBom::where(['uuid' => $this->product_bom_uuid])->dec('stock_num', $this->num)->update();
            }
            
            // 撤销材料库存
            if ($this->material_uuid > 0) {
                if (!$this->material) {
                    $this->error = '规格不存在，无法进行撤销操作';
                    return false;
                }
                if ($this->material->stock_num < $this->num) {
                    $this->error = '原料库存不足';
                    return false;
                }
                Material::where(['uuid' => $this->material_uuid])->dec('stock_num', $this->num)->update();
                $relatedMaterialUuidList = RelatedMaterial::where('material_uuid', $this->material_uuid)->column('uuid') ?: [];
                RelatedMaterial::updateStock($relatedMaterialUuidList);
            }
            
            $this->status = self::STATUS_CANCELED;
            $this->revoke_time = time();
            $this->save();
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
        if ($this->status != self::STATUS_CANCELED) {
            $this->error = '记录不能删除';
            return false;
        }
        return $this->delete();
    }

    /**
     * 入库编号：18位纯数字（前2位WR，2-10位是年月日，中间位是0000，后4位随机生成）
     *
     * @return string
     */
    public function generateInCode()
    {
        $date = date('Ymd');
        $rand = rand(1000, 9999);
        $code = 'WR' . $date . '0000' . $rand;
        return $code;
    }
}
