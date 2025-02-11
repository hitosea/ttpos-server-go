<?php

namespace app\common\model\erp;

use think\facade\Env;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\ProductSku;

/**
 * 采购订单模型
 */
class ErpPurchaseOrder extends BaseModel
{
    use SoftDelete;
    protected $name = 'purchase_form';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * 采购方式 10-总部采购 20-自行采购
     */
    const TYPE_ALL = 10;
    const TYPE_SELF = 20;

    /**
     * 状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库
     */
    const STATUS_WAIT = 10;
    const STATUS_REJECTED = 20;
    const STATUS_PURCHASING = 30;
    const STATUS_PURCHASED = 40;
    const STATUS_STORED = 50;

    /**
     * 商品总数
     */
    public function getTotalNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 采购员
     */
    public function purchaser()
    {
        return $this->belongsTo(User::class, 'purchaser_id', 'shop_user_id')->field('shop_user_id, real_name');
    }

    /**
     * 申请人
     */
    public function applicant()
    {
        return $this->belongsTo(User::class, 'applicant_id', 'shop_user_id')->field('shop_user_id, real_name');
    }

    /**
     * 采购单明细
     */
    public function details()
    {
        return $this->hasMany(ErpPurchaseDetail::class, 'purchase_order_id', 'id');
    }

    /**
     * 采购单操作日志
     */

    public function logs()
    {
        return $this->hasMany(ErpPurchaseOperationLog::class, 'purchase_order_id', 'id')->with('operator');
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
        $erpSupplierId = trim($params['erp_supplier_id'] ?? "");
        $purchaserId = trim($params['purchaser_id'] ?? "");

        //
        $where = '';
        if (!empty($erpSupplierId)) {
            $where .= "where es.id = $erpSupplierId";
        }
        if (!empty($purchaserId)) {
            if (empty($erpSupplierId)) {
                $where .= "where es.purchaser_id = $purchaserId";
            } else {
                $where .= " and es.purchaser_id = $purchaserId";
            }
        }

        //
        $prefix = Env::get('DB_PREFIX');
        return self::alias('eo')->with(['purchaser'])
            ->field('eo.*, de.erp_supplier_names, de.erp_purchase_names, COALESCE(NULLIF(suser.real_name, ""), suser.user_name) as applicant_name')
            ->leftJoin("shop_user suser", "eo.applicant_id = suser.shop_user_id")
            ->leftJoin(
                "
                (
                    SELECT
                        de.purchase_order_id, es.purchaser_id, p.erp_supplier_id,
                        group_concat(DISTINCT es.name) as erp_supplier_names,
                        group_concat(DISTINCT user.real_name) as erp_purchase_names
                    FROM {$prefix}erp_purchase_detail as de
                    LEFT JOIN {$prefix}product_sku as sku ON de.product_sku_id = sku.product_sku_id
                    LEFT JOIN {$prefix}product as p ON sku.product_id = p.product_id
                    LEFT JOIN {$prefix}erp_supplier as es ON es.id = p.erp_supplier_id
                    LEFT JOIN {$prefix}shop_user as user ON es.purchaser_id = user.shop_user_id
                    $where
                    GROUP BY de.purchase_order_id
                ) as de",
                "eo.id = de.purchase_order_id"
            )
            ->when(!empty($erpSupplierId), function ($q) use ($erpSupplierId) {
                $q->where('de.erp_supplier_id', $erpSupplierId);
            })
            ->when(!empty($purchaserId), function ($q) use ($purchaserId) {
                $q->where('de.purchaser_id', $purchaserId);
            })
            ->when(isset($params['keyword']) && $params['keyword'], function ($q) use ($params) {
                $q->where(function ($q) use ($params) {
                    $q->like('eo.name', $params['keyword']);
                    $q->orLike('suser.real_name', $params['keyword']);
                });
            })
            ->when(isset($params['status']) && $params['status'], function ($q) use ($params) {
                $q->where('eo.status', $params['status']);
            })
            ->when($startTime && $endTime, function ($q) use ($startTime, $endTime) {
                $q->where('eo.create_time', 'between', [$startTime, $endTime]);
            })
            ->order('eo.create_time desc')
            ->paginate($params);
    }

    /**
     * 详情
     *
     * @param [type] $id
     * @return self
     */
    public function detail($id): self
    {
        return self::alias('eo')
            ->with(['purchaser', 'details' => function ($q) {
                $q->alias('de')->with(['productImage' => function ($qi) {
                    $qi->field('file_id, save_name, storage');
                }])
                    ->field('de.*, p.type as product_type, p.product_name, es.name as erp_supplier_name, sku.spec_name as product_sku_name, pm.image_id as product_image_id')
                    ->field('IF(p.type=10, sku.stock_num, sku.material_stock) as product_stock, sku.product_price')
                    ->leftJoin("product_sku sku", "de.product_sku_id = sku.product_sku_id")
                    ->leftJoin("product p", "sku.product_id = p.product_id")
                    ->leftJoin("erp_supplier es", "es.id = p.erp_supplier_id")
                    ->leftJoin("product_image pm", "pm.product_id = p.product_id")
                    ->group('sku.product_sku_id');
            }, 'logs'])
            ->field('eo.*, suser.real_name as applicant_name, supplier.name as supplier_name')
            ->leftJoin("shop_user suser", "eo.applicant_id = suser.shop_user_id")
            ->leftJoin("supplier supplier", "supplier.shop_supplier_id = eo.shop_supplier_id")
            ->find($id);
    }

    /**
     * 新增
     */
    public function add($params)
    {
        $shopUserId = $params['shop_user_id'] ?? 0;
        $shopSupplierId = $params['shop_supplier_id'] ?? 0;

        if (!isset($params['purchase_detail']) || empty($params['purchase_detail'])) {
            $this->error = '请选择采购明细';
            return false;
        }
        $this->startTrans();
        try {
            $model = new self;
            $total_num = 0;
            $total_amount = 0;
            foreach ($params['purchase_detail'] as &$value) {
                $total_num += floatval($value['estimate_purchase_num']);
                $total_amount += floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
            }
            $data = [
                'number' => $this->getPurchaseNumber(),
                'name' => $params['name'],
                'type' => $params['type'],
                'applicant_id' => $params['applicant_id'],
                'total_num' => $total_num,
                'total_amount' => $total_amount,
                'arrival_time' => strtotime($params['arrival_time']),
                'remark' => $params['remark'],
                'shop_supplier_id' => $shopSupplierId,
                'app_id' => self::$app_id,
            ];
            $model->save($data);
            //
            foreach ($params['purchase_detail'] as &$value) {
                // 判断价格格式为 小数点后2位，范围：0-1000000
                if (!preg_match('/^(0|[1-9]\d{0,6})(\.\d{1,2})?$/', $value['estimate_purchase_price']) || floatval($value['estimate_purchase_price']) > 1000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-1000000';
                    return false;
                }
                // 判断数量格式为 小数点后4位，范围：0-99999999
                if (!preg_match('/^(0|[1-9]\d{0,7})(\.\d{1,4})?$/', $value['estimate_purchase_num']) || floatval($value['estimate_purchase_num']) > 99999999) {
                    $this->error = '数量格式错误，请输入小数点后4位，范围：0-99999999';
                    return false;
                }
                $value['purchase_order_id'] = $model->id;
                $value['estimate_total_amount'] = floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
                $value['shop_supplier_id'] = $shopSupplierId;
                $value['app_id'] = self::$app_id;
            }
            $purchaseDetailModel = new ErpPurchaseDetail;
            $purchaseDetailModel->saveAll($params['purchase_detail']);
            // 操作日志
            $shopUser = User::where('shop_user_id', $shopUserId)->field('shop_user_id, real_name, user_name')->find();
            $data = [
                'purchase_order_id' => $model->id,
                'operator_id' => $shopUser->shop_user_id ?: 0,
                'username' => $shopUser->real_name ?: $shopUser->user_name,
                'status' => 10, // 待审核
                'operation' => '添加',
                'shop_supplier_id' => $shopSupplierId,
                'app_id' => self::$app_id,
            ];
            (new ErpPurchaseOperationLog)->save($data);
            $this->commit();
            return $model->id;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 编辑
     */
    public function edit($params)
    {
        $shopUserId = $params['shop_user_id'] ?? 0;
        $shopSupplierId = $this->shop_supplier_id ?? 0;

        if (!isset($params['purchase_detail']) || empty($params['purchase_detail'])) {
            $this->error = '请选择采购明细';
            return false;
        }

        $this->startTrans();
        try {
            $total_num = 0;
            $total_amount = 0;
            foreach ($params['purchase_detail'] as &$value) {
                // 判断价格格式为 小数点后2位，范围：0-1000000
                if (!preg_match('/^(0|[1-9]\d{0,6})(\.\d{1,2})?$/', $value['estimate_purchase_price']) || floatval($value['estimate_purchase_price']) > 1000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-1000000';
                    return false;
                }
                // 判断数量格式为 小数点后4位，范围：0-99999999
                if (!preg_match('/^(0|[1-9]\d{0,7})(\.\d{1,4})?$/', $value['estimate_purchase_num']) || floatval($value['estimate_purchase_num']) > 99999999) {
                    $this->error = '数量格式错误，请输入小数点后4位，范围：0-99999999';
                    return false;
                }
                $total_num += floatval($value['estimate_purchase_num']);
                $total_amount += floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
                //
                $value['purchase_order_id'] = $this->id;
                $value['estimate_total_amount'] = floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
                $value['shop_supplier_id'] = $shopSupplierId;
                $value['app_id'] = self::$app_id;
            }

            $data = [
                'name' => $params['name'],
                'type' => $params['type'],
                'applicant_id' => $params['applicant_id'],
                'total_num' => $total_num,
                'total_amount' => $total_amount,
                'arrival_time' => strtotime($params['arrival_time']),
                'remark' => $params['remark'],
                'app_id' => $this->app_id,
            ];
            $this->save($data);

            $purchaseDetailModel = new ErpPurchaseDetail;
            $purchaseDetailModel->where('purchase_order_id', $this->id)->delete(); // 先删后加
            $purchaseDetailModel->saveAll($params['purchase_detail']);
            // 操作日志
            $shopUser = User::where('shop_user_id', $shopUserId)->field('shop_user_id, real_name, user_name')->find();
            $data = [
                'purchase_order_id' => $this->id,
                'operator_id' => $shopUser->shop_user_id ?: 0,
                'username' => $shopUser->real_name ?: $shopUser->user_name,
                'status' => $this->status,
                'operation' => '编辑',
                'shop_supplier_id' => $shopSupplierId,
                'app_id' => $this->app_id,
            ];
            (new ErpPurchaseOperationLog)->save($data);
            //
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 调整数据
     */
    public function adjust($params)
    {
        $shopUserId = $params['shop_user_id'] ?? 0;
        $shopSupplierId = $this->shop_supplier_id ?? 0;

        if (!isset($params['purchase_detail']) || empty($params['purchase_detail'])) {
            $this->error = '请选择采购明细';
            return false;
        }
        $purchaseDetailModel = new ErpPurchaseDetail;
        $total_num = 0;
        $total_amount = 0;
        //
        $this->startTrans();
        try {
            foreach ($params['purchase_detail'] as &$value) {
                // 判断价格格式为 小数点后2位，范围：0-1000000
                if (!preg_match('/^(0|[1-9]\d{0,6})(\.\d{1,2})?$/', $value['actual_purchase_price']) || floatval($value['actual_purchase_price']) > 1000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-1000000';
                    return false;
                }
                // 判断数量格式为 小数点后4位，范围：0-99999999
                if (!preg_match('/^(0|[1-9]\d{0,7})(\.\d{1,4})?$/', $value['actual_purchase_num']) || floatval($value['actual_purchase_num']) > 99999999) {
                    $this->error = '数量格式错误，请输入小数点后4位，范围：0-99999999';
                    return false;
                }
                $total_num += floatval($value['actual_purchase_num']);
                $total_amount += floatval($value['actual_purchase_price']) * floatval($value['actual_purchase_num']);
                //
                $purchaseDetailModel->where('id', $value['purchase_detail_id'])->save([
                    'id' => $value['purchase_detail_id'],
                    'actual_purchase_num' => $value['actual_purchase_num'],
                    'actual_purchase_price' => $value['actual_purchase_price'],
                    'actual_total_amount' => floatval($value['actual_purchase_price']) * floatval($value['actual_purchase_num']),
                ]);
            }
            $data = [
                'total_num' => $total_num,
                'total_amount' => $total_amount,
            ];
            $this->save($data);
            // 操作日志
            $shopUser = User::where('shop_user_id', $shopUserId)->field('shop_user_id, real_name, user_name')->find();
            $operationLog = new ErpPurchaseOperationLog;
            $operationLogData = [
                'purchase_order_id' => $this->id,
                'operator_id' => $params['shop_user_id'] ?? $shopUser?->shop_user_id ?: 0,
                'username' => $params['username'] ?? $shopUser?->real_name ?: $shopUser?->user_name,
                'status' => $this->status,
                'operation' => '调整数据',
                'shop_supplier_id' => $shopSupplierId,
                'app_id' => $this->app_id,
            ];
            //
            if ($params['operation_log_id'] ?? '') {
                $operationLogData['id'] = $params['operation_log_id'];
            }
            //
            $operationLog->save($operationLogData);
            //
            $this->commit();
            //
            return $operationLog->getKey();
            //
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 操作
     */
    public function operate($params)
    {
        $shopUserId = $params['shop_user_id'] ?? 0;
        $shopSupplierId = $this->shop_supplier_id ?? 0;

        // 通过和驳回操作，需要状态为待审核
        if (in_array($params['status'], [self::STATUS_PURCHASING, self::STATUS_REJECTED], true)) {
            if ($this['status'] !== self::STATUS_WAIT) {
                $this->error = '订单状态不能操作';
                return false;
            }
        }
        // 已采购操作，需要状态为通过
        if ($params['status'] == self::STATUS_PURCHASED) {
            if ($this['status'] != self::STATUS_PURCHASING) {
                $this->error = '订单状态不能操作';
                return false;
            }
        }
        // 已入库操作，需要状态为已采购
        if ($params['status'] == self::STATUS_STORED) {
            if ($this['status'] != self::STATUS_PURCHASED && $this['status'] != self::STATUS_PURCHASING) {
                $this->error = '订单状态不能操作';
                return false;
            }
        }

        $this->startTrans();
        try {
            $this->status = $params['status'];
            $this->save();
            // 采购单明细实际采购数量和实际价格
            $purchaseDetailModel = new ErpPurchaseDetail;
            $detailList = $purchaseDetailModel->where('purchase_order_id', $this->id)->select();
            foreach ($detailList as $detail) {
                $detail->actual_purchase_num = $detail['actual_purchase_num'] > 0 ? $detail['actual_purchase_num'] : $detail['estimate_purchase_num'];
                $detail->actual_purchase_price = $detail['actual_purchase_price'] > 0 ? $detail['actual_purchase_price'] : $detail['estimate_purchase_price'];
                $detail->actual_total_amount = $detail->actual_purchase_price * $detail->actual_purchase_num;
                $detail->save();
            }
            // 操作日志
            $shopUser = User::where('shop_user_id', $shopUserId)->field('shop_user_id, real_name, user_name')->find();
            $operationLog = new ErpPurchaseOperationLog;
            $operationLogData = [
                'purchase_order_id' => $this->id,
                'operator_id' => $params['shop_user_id'] ?? $shopUser?->shop_user_id ?: 0,
                'username' => $params['username'] ?? $shopUser?->real_name ?: $shopUser?->user_name,
                'status' => $this->status,
                'operation' => ErpPurchaseOperationLog::statusOperation[$this->status],
                'remark' => $params['remark'] ?? '',
                'shop_supplier_id' => $shopSupplierId,
                'app_id' => $this->app_id,
            ];
            //
            if ($params['operation_log_id'] ?? '') {
                $operationLogData['id'] = $params['operation_log_id'];
            }
            $operationLog->save($operationLogData);
            // 采购入库操作
            $inventoryRecordId = 0;
            if ($params['status'] == self::STATUS_STORED) {
                // 商品库存增加
                $materialIds = [];
                foreach ($detailList as $detail) {
                    $product = $detail->product;
                    if (!$product) continue;
                    switch ($product['type']) {
                        case Product::TYPE_PRODUCT:
                            Product::where(['product_id' => $detail->product_id])->inc('product_stock', $detail->actual_purchase_num)->update();
                            // 规格库存增加
                            ProductSku::where(['product_sku_id' => $detail->product_sku_id])->inc('stock_num', $detail->actual_purchase_num)->update();
                            break;
                        case Product::TYPE_MATERIAL:
                            $materialIds = array_merge($materialIds, [$detail->product_id]);
                            Product::where(['product_id' => $detail->product_id])->inc('product_material_stock', $detail->actual_purchase_num)->update();
                            // 规格库存增加
                            ProductSku::where(['product_sku_id' => $detail->product_sku_id])->inc('material_stock', $detail->actual_purchase_num)->update();
                            break;
                    }
                }
                // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
                (new Product)->reCalProductStock(array_unique($materialIds));
                // 入库记录
                $inventoryRecord = new ErpInventoryRecord;
                $inventoryRecordData = [
                    'purchase_order_id' => $this->id,
                    'type' => ErpInventoryRecord::TYPE_PURCHASE_IN,
                    'num' => $this->total_num,
                    'operator_id' => $operationLogData['operator_id'],
                    'username' => $operationLogData['username'],
                    'remark' => $this->remark ?: '',
                    'shop_supplier_id' => $shopSupplierId,
                ];
                //
                if ($params['erp_inventory_record_id'] ?? '') {
                    $inventoryRecordData['id'] = $params['erp_inventory_record_id'];
                }
                //
                $inventoryRecord->addNew(ErpInventoryRecord::INVENTORY_TYPE_IN, $inventoryRecordData);
            }
            $this->commit();
            //
            return [$operationLog->getKey(), $inventoryRecordId];
            //
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 删除
     */
    public function del()
    {
        // 待审核和驳回的订单才能删除
        if (!in_array($this['status'], [self::STATUS_WAIT, self::STATUS_REJECTED], true)) {
            $this->error = '订单状态不能删除';
            return false;
        }
        return $this->destroy(['id' => $this['id']]);
    }

    /**
     * 采购编号：18位纯数字（前2位PU，2-10位是年月日，中间位是0000，后4位随机生成）
     */
    public function getPurchaseNumber()
    {
        $prefix = 'PU';
        $date = date('Ymd');
        $random = rand(1000, 9999);
        $number = $prefix . $date . '0000' . $random;
        return $number;
    }

    /**
     * 获取采购单数量
     */
    public function getPurchaseOrderCount($status, $shop_supplier_id = 0)
    {
        $model = new self;
        if ($status) {
            $model = $model->where('status', $status);
        }
        return $model->count();
    }
}
