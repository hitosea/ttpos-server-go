<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\ProductBom;
use app\common\model\product\Material;
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
    protected $append = ['product_sku_name_text', 'number', 'type', 'in_time'];

    /**
     * inventory_type 类型 1-入库 2-出库
     */
    const INVENTORY_TYPE_IN = 1;
    const INVENTORY_TYPE_OUT = 2;

    /**
     * type 操作类型 10-采购入库 20-调整入库 21-添加入库 30-销售出库 40-调整出库 41-删除出库
     */
    const TYPE_PURCHASE_IN = 10;
    const TYPE_ADJUST_IN = 20;
    const TYPE_ADJUST_IN_ADD = 21;
    const TYPE_SALE_OUT = 30;
    const TYPE_ADJUST_OUT = 40;
    const TYPE_ADJUST_OUT_DEL = 41;
    const OLD_TYPE = [
        10 => self::SCENE_PURCHASE_IN,
        20 => self::SCENE_ADJUST_IN,
        21 => self::SCENE_ADD_IN,
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

    /**
     * 兼容字段
     */
    public function getNumberAttr($value, $data)
    {
        return $this->form_no;
    }
    public function getTypeAttr($value, $data)
    {
        return $this->scene;
    }
    public function getInTimeAttr($value, $data)
    {
        return $this->getData('create_time');
    }

    /**
     * 商品规格名称
     */
    public static function getProductSkuNameTextAttr($value, $data = [])
    {
        if (isset($data['product_sku_name']) && $data['product_sku_name']) {
            return extractLanguage($value ?: $data['product_sku_name'] ?? '');
        } else {
            // 兼容旧数据
            $sku = ProductBom::where('uuid', $data['material_uuid'])->find();
            return extractLanguage($value ?: $sku['spec_name'] ?? '');
        }
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
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_uuid', 'uuid');
    }

    /**
     * 关联产品SKU
     */
    public function productSku()
    {
        return $this->belongsTo(ProductBom::class, 'material_uuid', 'uuid');
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

        $model = new self;
        if (isset($params['name']) && $params['name']) {
            $model = $model->like('name', $params['name']);
        }

        // 操作类型 10-采购入库 20-调整入库 21-添加入库 30-销售出库 40-调整出库 41-删除出库
        if (isset($params['type']) && $params['type']) {
            $model = $model->where('scene', self::OLD_TYPE[$params['type']]);
        }

        // 起始时间
        if ($startTime && $endTime) {
            $model = $model->where('create_time', 'between', [$startTime, $endTime]);
        }

        $list = $model->with([
            'purchaseOrder' => function ($q) {
                $q->field('id, name, num as number');
            },
            // 'product' => function ($q) {
            //     $q->field('product_id, product_name, type, product_unit');
            // },
            // 'productSku' => function ($q) {
            //     $q->field('product_sku_id, product_id, spec_name, stock_num, material_stock')->with('material');
            // },
            'operator' => function ($q) {
                $q->field('uuid, uuid as shop_user_id, username as user_name, real_name');
            }
        ])->order('create_time desc')->paginate($params);
        //
        foreach ($list as $item) {
            // 是否显示入库撤销按钮 1-显示 0-不显示
            if ($item->purchase_order_uuid == 0) {
                $item['is_show_in_cancel'] = $this->checkStock($item);
            } else {
                $purchaseDetailList = (new ErpPurchaseDetail)->getListAll($item->purchase_order_uuid);
                foreach ($purchaseDetailList as $detail) {
                    if (!$detail->sku) {
                        $item['is_show_in_cancel'] = 0;
                    }
                    if ($detail->sku?->product_sku_id == $detail->product_sku_id) {
                        $item['is_show_in_cancel'] = $this->checkStock($detail, $detail->actual_purchase_num);
                    }
                }
            }
        }
        return $list;
    }

    /**
     * 检查库存
     *
     * @param [type] $item
     * @param [type] $num
     * @return int
     */
    private function checkStock($item, $num = null)
    {
        if (!$item->product) {
            return 0;
        }
        $num = $num ?: $item->num;
        $stockNum = 0;
        switch ($item->product['type']) {
            case Product::TYPE_PRODUCT:
                if ($item?->purchase_order_uuid == 0) {
                    if ($item->productSku?->material?->count() > 0) {
                        return 0;
                    }
                    $stockNum = $item->productSku?->stock_num;
                } else {
                    if ($item->sku?->material?->count() > 0) {
                        return 0;
                    }
                    $stockNum = $item->sku?->stock_num;
                }
                break;
            case Product::TYPE_MATERIAL:
                $stockNum = $item->sku?->material_stock;
                break;
        }
        return $stockNum >= $num ? 1 : 0;
    }

    /**
     * 详情
     */
    public function detail($id)
    {
        $model = new self;
        $info = $model->with(['purchaseOrder', 'product', 'productSku', 'productSku.material', 'operator'])
            ->where('id', $id)
            ->find();
        return $info;
    }

    /**
     * 新增（出/入库）
     */
    public function add($params, $isSave = true)
    {
        $data['form_no'] = self::generateInCode();
        $data['scene'] = $params['scene'] ?? 0;
        $data['purchase_order_uuid'] = $params['purchase_order_uuid'] ?? 0;
        $data['material_uuid'] = $params['material_uuid'] ?? 0;
        $data['num'] = $params['num'] ?? 0;
        $data['remark'] = $params['remark'] ?? '';
        $data['operator_uuid'] = $params['operator_uuid'];
        $data['status'] = $params['status'] ?? self::STATUS_SUCCESS;
        //
        if ($isSave) {
            $model = new self;
            $model->save($data);
            return $model->uuid;
        }
        return $data;
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
            // 判断是否库存足够
            if ($this->material_uuid > 0) {
                // 入库撤销操作，回滚减少库存
                if ($this->material_uuid > 0) {
                    $product = $this->productSku;
                    $material = $this->material;
                    if ($product) {
                        if ($product->stock_num < $this->num) {
                            $this->error = '商品规格库存不足';
                            return false;
                        }
                        ProductBom::where(['uuid' => $this->material_uuid])->dec('stock_num', $this->num)->update();
                    } elseif ($material) {
                        if ($material->stock_num < $this->num) {
                            $this->error = '原料库存不足';
                            return false;
                        }
                        Material::where(['uuid' => $this->material_uuid])->dec('stock_num', $this->num)->update();
                    }
                }
            } else {
                $purchaseDetailList = (new ErpPurchaseDetail)->getListAll($this->purchase_order_uuid);
                // 遍历回滚采购单数量
                foreach ($purchaseDetailList as $detail) {
                    if (!$detail->sku) {
                        $this->error = '规格不存在，无法进行撤销操作';
                        return false;
                    }
                    if ($detail->sku?->uuid == $detail->material_uuid) {
                        // 入库撤销操作，回滚减少库存 物料类型 0-商品 1-原料
                        switch ($detail->material_type) {
                            case ErpPurchaseDetail::MATERIAL_TYPE_PRODUCT:
                                if ($detail->sku?->stock_num < $this->num) {
                                    $this->error = '库存不足';
                                    return false;
                                }
                                ProductBom::where(['uuid' => $detail->material_uuid])->dec('stock_num', $detail->num)->update();
                                break;
                            case ErpPurchaseDetail::MATERIAL_TYPE_MATERIAL:
                                if ($detail->material?->stock_num < $this->num) {
                                    $this->error = '库存不足';
                                    return false;
                                }
                                Material::where(['uuid' => $detail->material_uuid])->dec('stock_num', $detail->num)->update();
                                break;
                        }
                    }
                }
            }
            //
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
        return $this->destroy(['id' => $this['id']]);
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
}
