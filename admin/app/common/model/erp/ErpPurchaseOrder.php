<?php

namespace app\common\model\erp;

use think\facade\Env;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\product\ProductSku;
use app\common\model\erp\ErpWarehouseForm;

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
    protected $append = ['number', 'applicant_id', 'total_num', 'total_amount', 'status'];

    /**
     * 采购方式 10-总部采购 20-自行采购
     */
    const TYPE_ALL = 10;
    const TYPE_SELF = 20;

    /**
     * 状态 10-待审核 20-已驳回 30-采购中 40-已采购 50-已入库
     * 状态 0-待审核 1-已驳回 2-采购中 3-已采购 4-已入库
     */
    const STATUS_WAIT = 10;
    const STATUS_REJECTED = 20;
    const STATUS_PURCHASING = 30;
    const STATUS_PURCHASED = 40;
    const STATUS_STORED = 50;

    /**
     * 原状态映射
     */
    const ORIGINAL_STATUS = [
        0 => 10, // 待审核
        1 => 20, // 已驳回
        2 => 30, // 采购中
        3 => 40, // 已采购
        4 => 50  // 已入库
    ];

    /**
     * 新的状态映射
     */
    const NEW_STATUS = [
        10 => 0, // 待审核
        20 => 1, // 已驳回
        30 => 2, // 采购中
        40 => 3, // 已采购
        50 => 4  // 已入库
    ];

    /**
     * 兼容字段
     */
    public function getNumberAttr($value)
    {
        return $this->form_no ?: '';
    }
    public function getTotalNumAttr($value)
    {
        return floatval($this->num ?: 0);
    }
    public function getApplicantIdAttr($value)
    {
        return $this->applicant_uuid ?: '';
    }
    public function getTotalAmountAttr($value)
    {
        return floatval($this->amount ?: 0);
    }
    public function getStatusAttr($value)
    {
        return self::ORIGINAL_STATUS[$this->getData('status')] ?? 10;
    }

    /**
     * 采购员
     */
    public function purchaser()
    {
        return $this->belongsTo(User::class, 'purchaser_uuid', 'uuid')->field('uuid, real_name');
    }

    /**
     * 申请人
     */
    public function applicant()
    {
        return $this->belongsTo(User::class, 'applicant_uuid', 'uuid')->field('uuid, real_name');
    }

    /**
     * 采购单明细
     */
    public function details()
    {
        return $this->hasMany(ErpPurchaseDetail::class, 'purchase_form_uuid', 'uuid');
    }

    /**
     * 采购单操作日志
     */

    public function logs()
    {
        return $this->hasMany(ErpPurchaseOperationLog::class, 'purchase_form_uuid', 'uuid')->with('operator');
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
            ->field('eo.*, COALESCE(NULLIF(suser.real_name, ""), suser.username) as applicant_name')
            ->leftJoin("staff suser", "eo.applicant_uuid = suser.uuid")

            // todo 兼容
            // ->leftJoin(
            //     "
            //     (
            //         SELECT
            //             de.purchase_order_id, es.purchaser_id, p.erp_supplier_id,
            //             group_concat(DISTINCT es.name) as erp_supplier_names,
            //             group_concat(DISTINCT user.real_name) as erp_purchase_names
            //         FROM {$prefix}erp_purchase_detail as de
            //         LEFT JOIN {$prefix}product_sku as sku ON de.product_sku_id = sku.product_sku_id
            //         LEFT JOIN {$prefix}product as p ON sku.product_id = p.product_id
            //         LEFT JOIN {$prefix}erp_supplier as es ON es.id = p.erp_supplier_id
            //         LEFT JOIN {$prefix}shop_user as user ON es.purchaser_id = user.shop_user_id
            //         $where
            //         GROUP BY de.purchase_order_id
            //     ) as de",
            //     "eo.id = de.purchase_order_id"
            // )

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
        $appId = request()->appId;
        $licenses = request()->licenses;
        return self::alias('eo')
            ->with([
                'purchaser',
                'details',
                // todo 兼容
                // 'details' => function ($q) {
                //     $q->alias('de')->with(['productImage' => function ($qi) {
                //         $qi->field('file_id, save_name, storage');
                //     }])
                //         ->field('de.*, p.type as product_type, p.product_name, es.name as erp_supplier_name, sku.spec_name as product_sku_name, pm.image_id as product_image_id')
                //         ->field('IF(p.type=10, sku.stock_num, sku.material_stock) as product_stock, sku.product_price')
                //         ->leftJoin("product_sku sku", "de.product_sku_id = sku.product_sku_id")
                //         ->leftJoin("product p", "sku.product_id = p.product_id")
                //         ->leftJoin("erp_supplier es", "es.id = p.erp_supplier_id")
                //         ->leftJoin("product_image pm", "pm.product_id = p.product_id")
                //         ->group('sku.product_sku_id');
                // },
                'logs'
            ])
            ->field("eo.*, suser.real_name as applicant_name, {$appId} as supplier_name")
            ->leftJoin("staff suser", "eo.applicant_uuid = suser.uuid")
            ->where('eo.id', $id)
            ->find();
    }

    /**
     * 简单详情
     */
    public function simpleDetail($id)
    {
        return self::where('id', $id)->find();
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
                'form_no' => $this->getPurchaseNumber(),
                'name' => $params['name'],
                'type' => $params['type'],
                'applicant_uuid' => $params['applicant_id'],
                'num' => $total_num,
                'amount' => $total_amount,
                'arrival_time' => strtotime($params['arrival_time']),
                'remark' => $params['remark'],
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
                $value['purchase_form_uuid'] = $model->uuid;
                $value['estimate_amount'] = floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
                $value['estimate_num'] = $value['estimate_purchase_num'];
                $value['estimate_price'] = $value['estimate_purchase_price'];
            }
            $purchaseDetailModel = new ErpPurchaseDetail;
            $purchaseDetailModel->saveAll($params['purchase_detail']);
            // 操作日志
            $shopUser = User::where('uuid', $shopUserId)->field('uuid, real_name, username')->find();
            $data = [
                'purchase_form_uuid' => $model->uuid,
                'operator_uuid' => $shopUser->uuid ?: 0,
                'username' => $shopUser->real_name ?: $shopUser->username,
                'status' => 10, // 待审核
                'operation' => '添加',
            ];
            (new ErpPurchaseOperationLog)->save($data);
            $this->commit();
            return $model->uuid;
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
                $value['purchase_form_uuid'] = $this->uuid;
                $value['estimate_num'] = $value['estimate_purchase_num'];
                $value['estimate_price'] = $value['estimate_purchase_price'];
                $value['estimate_amount'] = floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
            }

            $data = [
                'name' => $params['name'],
                'type' => $params['type'],
                'applicant_uuid' => $params['applicant_id'],
                'num' => $total_num,
                'amount' => $total_amount,
                'arrival_time' => strtotime($params['arrival_time']),
                'remark' => $params['remark'],
            ];
            $this->save($data);

            $purchaseDetailModel = new ErpPurchaseDetail;
            $purchaseDetailModel->where('purchase_form_uuid', $this->uuid)->delete(); // 先删后加
            $purchaseDetailModel->saveAll($params['purchase_detail']);
            // 操作日志
            $shopUser = User::where('uuid', $shopUserId)->field('uuid, real_name, username')->find();
            $data = [
                'purchase_form_uuid' => $this->uuid,
                'operator_uuid' => $shopUser->uuid ?: 0,
                'username' => $shopUser->real_name ?: $shopUser->username,
                'status' => $this->status,
                'operation' => '编辑',
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
                    'num' => $value['actual_purchase_num'],
                    'price' => $value['actual_purchase_price'],
                    'amount' => floatval($value['actual_purchase_price']) * floatval($value['actual_purchase_num']),
                ]);
            }
            $data = [
                'total_num' => $total_num,
                'total_amount' => $total_amount,
            ];
            $this->save($data);
            // 操作日志
            $shopUser = User::where('uuid', $shopUserId)->field('uuid, real_name, username')->find();
            $operationLog = new ErpPurchaseOperationLog;
            $operationLogData = [
                'purchase_form_uuid' => $this->uuid,
                'operator_uuid' => $params['shop_user_id'] ?? $shopUser?->uuid ?: 0,
                'username' => $params['username'] ?? $shopUser?->real_name ?: $shopUser?->username,
                'status' => $this->status,
                'operation' => '调整数据',
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
            $this->status = self::NEW_STATUS[$params['status']] ?? 0;
            $this->save();
            // 采购单明细实际采购数量和实际价格
            $purchaseDetailModel = new ErpPurchaseDetail;
            $detailList = $purchaseDetailModel->where('purchase_form_uuid', $this->uuid)->select();
            foreach ($detailList as $detail) {
                $detail->num = $detail['num'] > 0 ? $detail['num'] : $detail['estimate_num'];
                $detail->price = $detail['price'] > 0 ? $detail['price'] : $detail['estimate_price'];
                $detail->amount = $detail->price * $detail->num;
                $detail->save();
            }
            // 操作日志
            $shopUser = User::where('uuid', $shopUserId)->field('uuid, real_name, username')->find();
            $operationLog = new ErpPurchaseOperationLog;
            $operationLogData = [
                'purchase_form_uuid' => $this->uuid,
                'operator_uuid' => $params['shop_user_id'] ?? $shopUser?->uuid ?: 0,
                'username' => $params['username'] ?? $shopUser?->real_name ?: $shopUser?->username,
                'status' => $this->status,
                'operation' => ErpPurchaseOperationLog::statusOperation[$this->status],
                'remark' => $params['remark'] ?? '',
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
                    if ($detail->material_type == ErpPurchaseDetail::MATERIAL_TYPE_PRODUCT) {
                        $product = $detail->product;
                    } else {
                        $product = $detail->material;
                    }
                    if (!$product) continue;
                    switch ($product['type']) {
                        case Product::TYPE_PRODUCT:
                            Product::where(['product_id' => $detail->product_id])->inc('product_stock', $detail->num)->update();
                            // 规格库存增加
                            ProductSku::where(['product_sku_id' => $detail->product_sku_id])->inc('stock_num', $detail->num)->update();
                            break;
                        case Product::TYPE_MATERIAL:
                            $materialIds = array_merge($materialIds, [$detail->product_id]);
                            Product::where(['product_id' => $detail->product_id])->inc('product_material_stock', $detail->num)->update();
                            // 规格库存增加
                            ProductSku::where(['product_sku_id' => $detail->product_sku_id])->inc('material_stock', $detail->num)->update();
                            break;
                    }
                }
                // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
                (new Product)->reCalProductStock(array_unique($materialIds));
                // 入库记录
                $inventoryRecord = new ErpWarehouseForm;
                $inventoryRecordData = [
                    'purchase_form_uuid' => $this->uuid,
                    'type' => ErpWarehouseForm::TYPE_PURCHASE_IN,
                    'num' => $this->total_num,
                    'operator_uuid' => $operationLogData['operator_id'],
                    'username' => $operationLogData['username'],
                    'remark' => $this->remark ?: '',
                ];
                //
                if ($params['erp_inventory_record_id'] ?? '') {
                    $inventoryRecordData['id'] = $params['erp_inventory_record_id'];
                }
                //
                /** @var ErpWarehouseForm $inventoryRecord */
                $inventoryRecord->add(ErpWarehouseForm::INVENTORY_TYPE_IN, $inventoryRecordData);
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
