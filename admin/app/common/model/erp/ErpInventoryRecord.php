<?php

namespace app\common\model\erp;

use help\StringHelp;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\ProductSku;

/**
 * 库存记录模型
 */
class ErpInventoryRecord extends BaseModel
{
    use SoftDelete;
    protected $name = 'warehouse_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['product_sku_name_text'];

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

    /**
     * status 状态 10-已入库 20-已出库 30-已撤销
     */

    const STATUS_IN = 10;
    const STATUS_OUT = 20;
    const STATUS_REVOKED = 30;

    //
    public static function onBeforeInsert($model)
    {
        if (!isset($model['id'])) {
            $model['id'] = StringHelp::uuid();
        }
        return $model;
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
            $sku = ProductSku::where('product_sku_id', $data['product_sku_id'])->find();
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
        return $this->belongsTo(ErpPurchaseOrder::class, 'purchase_order_id', 'id');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_id', 'product_id');
    }

    /**
     * 关联产品SKU
     */
    public function productSku()
    {
        return $this->belongsTo(ProductSku::class, 'product_sku_id', 'product_sku_id');
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_id', 'shop_user_id')->field(['shop_user_id', 'user_name', 'real_name']);
    }

    /**
     * 获取列表
     *
     * @param [type] $params
     * @return object
     */
    public function getList($params, $type = self::INVENTORY_TYPE_IN)
    {
        $startTime = isset($params['date'][0]) ? strtotime($params['date'][0]) : 0;
        $endTime = isset($params['date'][1]) ? strtotime($params['date'][1] . ' 23:59:59') : 0;

        $model = new self;
        if (isset($params['name']) && $params['name']) {
            $model = $model->like('name', $params['name']);
        }

        // 操作类型 10-采购入库 20-调整入库 21-添加入库 30-销售出库 40-调整出库 41-删除出库
        if (isset($params['type']) && $params['type']) {
            $model = $model->where('type', $params['type']);
        }

        // 起始时间
        if ($startTime && $endTime) {
            $model = $model->where('create_time', 'between', [$startTime, $endTime]);
        }

        $list = $model->with([
            'purchaseOrder' => function ($q) {
                $q->field('id, name, number');
            },
            'product' => function ($q) {
                $q->field('product_id, product_name, type, product_unit');
            },
            'productSku' => function ($q) {
                $q->field('product_sku_id, product_id, spec_name, stock_num, material_stock')->with('material');
            },
            'operator' => function ($q) {
                $q->field('uuid, uuid as shop_user_id, username as user_name, real_name');
            }
        ])->where('inventory_type', $type)->order('create_time desc')->paginate($params);
        //
        foreach ($list as $item) {
            if ($type == self::INVENTORY_TYPE_IN) {
                // 是否显示入库撤销按钮 1-显示 0-不显示
                if ($item->purchase_order_id == 0) {
                    $item['is_show_in_cancel'] = $this->checkStock($item);
                } else {
                    $purchaseDetailList = (new ErpPurchaseDetail)->getListAll($item->purchase_order_id);
                    foreach ($purchaseDetailList as $detail) {
                        if (!$detail->sku) {
                            $item['is_show_in_cancel'] = 0;
                        }
                        if ($detail->sku?->product_sku_id == $detail->product_sku_id) {
                            $item['is_show_in_cancel'] = $this->checkStock($detail, $detail->actual_purchase_num);
                        }
                    }
                }
            } else {
                // 是否显示出库撤销按钮 1-显示 0-不显示
                $item['is_show_out_cancel'] = 1;
                if (($item->product && $item->product['type'] == Product::TYPE_PRODUCT && $item->productSku?->material?->count() > 0) || self::TYPE_ADJUST_OUT_DEL == $item->type) {
                    $item['is_show_out_cancel'] = 0;
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
                if ($item?->purchase_order_id == 0) {
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
    public function addNew($inventory_type, $params, $isSave = true)
    {
        if ($inventory_type == self::INVENTORY_TYPE_IN) {
            $data['form_no'] = self::generateInCode();
            $data['in_time'] = time();
        } else {
            $data['form_no'] = self::generateOutCode();
            $data['out_time'] = time();
        }
        //
        $data['inventory_type'] = $inventory_type;
        $data['purchase_order_id'] = $params['purchase_order_id'] ?? 0;
        $data['product_id'] = $params['product_id'] ?? 0;
        $data['product_sku_id'] = $params['product_sku_id'] ?? 0;
        $data['product_sku_name'] = $params['product_sku_name'] ?? '';
        $data['order_id'] = $params['order_id'] ?? 0;
        $data['type'] = $params['type'] ?? 0;
        $data['num'] = $params['num'] ?? 0;
        $data['remark'] = $params['remark'] ?? '';
        $data['operator_id'] = $params['operator_id'];
        $data['status'] = $params['status'] ?? 10;
        $data['shop_supplier_id'] = $params['shop_supplier_id'] ?? 0;
        $data['app_id'] = self::$app_id;
        $data['id'] = StringHelp::uuid();
        $data['create_time'] = time();
        $data['update_time'] = time();
        //
        if ($isSave) {
            $model = new self;
            $model->save($data);
            return $model->id;
        }
        return $data;
    }

    /**
     * 撤销
     */
    public function cancel()
    {
        // 不是撤销状态的才能撤销
        if ($this->status == self::STATUS_REVOKED) {
            $this->error = '记录已撤销';
            return false;
        }
        $this->startTrans();
        try {
            // 如果是入库，需要回滚库存
            if ($this->inventory_type == self::INVENTORY_TYPE_IN) {
                // 判断是否库存足够
                if ($this->purchase_order_id == 0) {
                    // 入库撤销操作，回滚减少库存
                    switch ($this->product['type']) {
                        case Product::TYPE_PRODUCT:
                            if ($this->productSku?->stock_num < $this->num) {
                                $this->error = '库存不足';
                                return false;
                            }
                            Product::where(['product_id' => $this->product_id])->dec('product_stock', $this->num)->update();
                            ProductSku::where(['product_sku_id' => $this->product_sku_id])->dec('stock_num', $this->num)->update();
                            break;
                        case Product::TYPE_MATERIAL:
                            if ($this->productSku?->material_stock < $this->num) {
                                $this->error = '库存不足';
                                return false;
                            }
                            Product::where(['product_id' => $this->product_id])->dec('product_material_stock', $this->num)->update();
                            ProductSku::where(['product_sku_id' => $this->product_sku_id])->dec('material_stock', $this->num)->update();
                            break;
                    }
                } else {
                    $purchaseDetailList = (new ErpPurchaseDetail)->getListAll($this->purchase_order_id);
                    // 遍历回滚采购单数量
                    foreach ($purchaseDetailList as $detail) {
                        if (!$detail->sku) {
                            $this->error = '规格不存在，无法进行撤销操作';
                            return false;
                        }
                        if ($detail->sku?->product_sku_id == $detail->product_sku_id) {
                            // 入库撤销操作，回滚减少库存
                            switch ($detail->product['type']) {
                                case Product::TYPE_PRODUCT:
                                    if ($detail->sku?->stock_num < $this->actual_purchase_num) {
                                        $this->error = '库存不足';
                                        return false;
                                    }
                                    Product::where(['product_id' => $detail->product_id])->dec('product_stock', $detail->actual_purchase_num)->update();
                                    ProductSku::where(['product_sku_id' => $detail->product_sku_id])->dec('stock_num', $detail->actual_purchase_num)->update();
                                    break;
                                case Product::TYPE_MATERIAL:
                                    if ($detail->sku?->material_stock < $this->actual_purchase_num) {
                                        $this->error = '库存不足';
                                        return false;
                                    }
                                    Product::where(['product_id' => $detail->product_id])->dec('product_material_stock', $detail->actual_purchase_num)->update();
                                    ProductSku::where(['product_sku_id' => $detail->product_sku_id])->dec('material_stock', $detail->actual_purchase_num)->update();
                                    break;
                            }
                        }
                    }
                }
            }
            // 如果是出库，需要回滚库存
            if ($this->inventory_type == self::INVENTORY_TYPE_OUT) {
                // 如果是销售出库，不允许撤销
                if ($this->type == self::TYPE_SALE_OUT) {
                    $this->error = '销售出库不允许撤销';
                    return false;
                }
                // 如果是删除出库，不允许撤销
                if ($this->type == self::TYPE_ADJUST_OUT_DEL) {
                    $this->error = '删除出库不允许撤销';
                    return false;
                }
                if (!$this->productSku) {
                    $this->error = '规格不存在，无法进行撤销操作';
                    return false;
                }
                if ($this->product['type'] == Product::TYPE_PRODUCT && $this->productSku->material?->count() > 0) {
                    $this->error = '关联材料商品无法进行撤销操作';
                    return false;
                }
                // 出库撤销操作，回滚增加库存
                switch ($this->product['type']) {
                    case Product::TYPE_PRODUCT:
                        Product::where(['product_id' => $this->product_id])->inc('product_stock', $this->num)->update();
                        ProductSku::where(['product_sku_id' => $this->product_sku_id])->inc('stock_num', $this->num)->update();
                        break;
                    case Product::TYPE_MATERIAL:
                        Product::where(['product_id' => $this->product_id])->inc('product_material_stock', $this->num)->update();
                        ProductSku::where(['product_sku_id' => $this->product_sku_id])->inc('material_stock', $this->num)->update();
                        break;
                }
            }
            $this->status = self::STATUS_REVOKED;
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
        if ($this->status != self::STATUS_REVOKED) {
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
