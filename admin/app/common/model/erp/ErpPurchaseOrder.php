<?php

namespace app\common\model\erp;

use app\common\library\helper;
use think\facade\Env;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Material;
use app\shop\model\product\ProductBom;
use app\common\model\file\UploadFile;
use app\common\model\product\Product;
use app\shop\model\product\RelatedMaterial;
use think\facade\Db;

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
    protected $autoWriteTimestamp = true;

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
        $startTime = isset($params['date'][0]) ? strtotime($params['date'][0]) : 0;  // 搜索供应商
        $endTime = isset($params['date'][1]) ? strtotime($params['date'][1] . ' 23:59:59') : 0; // 搜索采购申请时间
        $erpSupplierId = $params['erp_supplier_id'] ?: 0;// 搜索状态
        $purchaserId = $params['purchaser_id'] ?: 0; // 搜索采购员
        $keyword = $params['keyword'] ?: ''; // 搜索采购名称/申请人
        $status = $params['status'] ?: 'all'; // 搜索状态

        //
        $prefix = Env::get('DB_PREFIX');
        $list = self::alias('o')
            ->field("
                o.*,
                COALESCE(NULLIF(u.real_name, ''), u.username) as applicant_name,
                GROUP_CONCAT(DISTINCT i.supplier_name SEPARATOR ', ') AS erp_supplier_names,
                GROUP_CONCAT(DISTINCT i.real_name SEPARATOR ', ') AS erp_purchase_names
            ")
            ->leftJoin('staff u', 'o.applicant_uuid = u.uuid')
            ->leftJoin(
                "
                (
                    SELECT
                        item.uuid,
                        item.purchase_form_uuid,
                        s.uuid AS s_uuid,
                        s1.uuid AS s1_uuid,
                        ss.uuid AS ss_uuid,
                        ss1.uuid AS ss1_uuid,
                        COALESCE (s.NAME, s1.NAME, '') AS supplier_name,
                        COALESCE (ss.real_name, ss1.real_name, '') AS real_name
                    FROM
                        {$prefix}purchase_form_item item
                    LEFT JOIN {$prefix}product_bom bom ON item.material_uuid = bom.uuid
                    LEFT JOIN {$prefix}product_package p ON bom.product_package_uuid = p.uuid
                    LEFT JOIN {$prefix}material m ON item.material_uuid = m.uuid
                    LEFT JOIN {$prefix}supplier s ON p.supplier_uuid = s.uuid
                    LEFT JOIN {$prefix}supplier s1 ON m.supplier_uuid = s1.uuid
                    LEFT JOIN {$prefix}staff ss ON s.staff_uuid = ss.uuid
                    LEFT JOIN {$prefix}staff ss1 ON s1.staff_uuid = ss1.uuid
                ) i
                ",
                "o.uuid = i.purchase_form_uuid"
            )
            ->withAttr('id', function($value, $data) {
                return $data['uuid'];
            })
            ->when($erpSupplierId, function($q) use ($erpSupplierId) {
                $q->where('i.s_uuid', $erpSupplierId)->whereOr('i.s1_uuid', $erpSupplierId);
            })
            ->when($purchaserId, function($q) use ($purchaserId) {
                $q->where('i.ss_uuid', $purchaserId)->whereOr('i.ss1_uuid', $purchaserId);
            })
            ->when($status != 'all', function($q) use ($status) {
                $q->where('o.status', self::NEW_STATUS[$status]);
            })
            ->when($keyword, function($q) use ($keyword) {
                $q->where('o.name', 'like', "%{$keyword}%")
                    ->whereOr('u.username', 'like', "%{$keyword}%")
                    ->whereOr('u.real_name', 'like', "%{$keyword}%");
            })
            ->when($startTime && $endTime, function($q) use ($startTime, $endTime) {
                $q->where('o.create_time', 'between', [$startTime, $endTime]);
            })
            ->group('o.uuid')
            ->order('o.create_time desc')
            ->paginate($params);

            foreach ($list->items() as $data) {
                $data['erp_supplier_names'] = ltrim($data['erp_supplier_names'], ', ');
                $data['erp_purchase_names'] = ltrim($data['erp_purchase_names'], ', ');
            }

            return $list;
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
        $form = self::alias('o')
            ->with([
                'details' => function($q) {
                    return $q->alias('i')
                        ->field("
                            i.*,
                            IF(i.material_type > 0, m.uuid, bom.uuid) as product_sku_id,
                            IF(i.material_type > 0, 20, 10) as product_type,
                            IF(i.material_type > 0, m.stock_num, bom.stock_num) as product_stock,
                            IF(i.material_type > 0, f2.file_url, f1.file_url) as file_url,
                            IF(i.material_type > 0, f2.file_name, f1.file_name) as file_name,
                            IF(i.material_type > 0, f2.storage, f1.storage) as storage,
                            IF(i.material_type > 0, f2.save_name, f1.save_name) as save_name,
                            IF(i.material_type > 0, f2.url_param, f1.url_param) as url_param,
                            IF(i.material_type > 0, m.name, p.name) as product_name,
                            IF(i.material_type > 0, '', f.name) as product_sku_name,
                            IF(i.material_type > 0, IF(s2.name IS NULL, '', s2.name), IF(s1.name IS NULL, '', s1.name)) as erp_supplier_name
                        ")
                        ->leftJoin('product_bom bom', 'i.material_uuid = bom.uuid')
                        ->leftJoin('product_package p', 'bom.product_package_uuid = p.uuid')
                        ->leftJoin('product_flavor f', 'bom.product_flavor_uuid = f.uuid')
                        ->leftJoin('file f1', 'p.image_file_uuid = f1.uuid')
                        ->leftJoin('supplier s1', 'p.supplier_uuid = s1.uuid')
                        ->leftJoin('material m', 'i.material_uuid = m.uuid')
                        ->leftJoin('file f2', 'm.image_uuid = f2.uuid')
                        ->leftJoin('supplier s2', 'm.supplier_uuid = s2.uuid');
                },
                'logs'
            ])
            ->withAttr('id', function($value, $data) {
                return $data['uuid'];
            })
            ->field("o.*, suser.real_name as applicant_name, {$appId} as supplier_name")
            ->leftJoin("staff suser", "o.applicant_uuid = suser.uuid")
            ->where('o.uuid', $id)
            ->find();
        
        if ($form) {
            $file = new UploadFile();
            foreach ($form->details as $item) {
                $filePath = $file->getFilePathAttr(null, [
                    'file_type' => $item['file_type'],
                    'file_url' => $item['file_url'],
                    'file_name' => $item['file_name'],
                    'storage' => $item['storage'],
                    'save_name' => $item['save_name'],
                    'url_param' => $item['url_param'],
                ]);
                $item->productImage = ['file_path' => $filePath];
                $item->product_name = extractLanguage($item->product_name);
                $item->product_sku_name = $item->product_sku_name ? extractLanguage($item->product_sku_name) : '';
                $item->product_stock = floatval($item->product_stock);
            }
        }
            
        return $form;
    }

    /**
     * 简单详情
     */
    public function simpleDetail($id)
    {
        return self::where('uuid', $id)->find();
    }

    /**
     * 新增
     */
    public function add($params)
    {
        $shopUserId = $params['shop_user_id'] ?? 0;

        if (!isset($params['purchase_detail']) || empty($params['purchase_detail'])) {
            $this->error = '请选择采购明细';
            return false;
        }
        $this->startTrans();
        try {
            $model = new self;
            $total_num = 0;
            $total_amount = 0;
            foreach ($params['purchase_detail'] as $value) {
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
            $dataList = [];
            $time = time();
            foreach ($params['purchase_detail'] as $value) {
                // 判断价格格式为 小数点后2位，范围：0-100000000
                if (!preg_match('/^(0|[1-9]\d{0,8})(\.\d{1,2})?$/', $value['estimate_purchase_price']) || floatval($value['estimate_purchase_price']) > 100000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-100000000';
                    return false;
                }
                // 判断数量格式为 小数点后4位，范围：0-99999999
                if (!preg_match('/^(0|[1-9]\d{0,7})(\.\d{1,4})?$/', $value['estimate_purchase_num']) || floatval($value['estimate_purchase_num']) > 99999999) {
                    $this->error = '数量格式错误，请输入小数点后4位，范围：0-99999999';
                    return false;
                }
                // 判断是商品还是材料
                $productBon = ProductBom::with(['product'])->where('uuid', $value['product_sku_id'])->find();
                if ($productBon) {
                    $materialType = 0;
                    $stockNum = $productBon->stock_num;
                    if (floatval(helper::bcadd($stockNum, $value['estimate_purchase_num'], 4)) > 99999999) {
                        $productName = extractLanguage($productBon->product->name);
                        $skuName = extractLanguage($productBon->name);
                        $this->error = "{$productName}({$skuName}) " . __('库存已超99999999');
                        return false;
                    }
                }
                $material = Material::where('uuid', $value['product_sku_id'])->find();
                if ($material) {
                    $materialType = 1;
                    $stockNum = $material->stock_num;
                    if (floatval(helper::bcadd($stockNum, $value['estimate_purchase_num'], 4)) > 99999999) {
                        $materialName = extractLanguage($material->name);
                        $this->error = "{$materialName} " . __('库存已超99999999');
                        return false;
                    }
                }
                $dataList[] = [
                    'purchase_form_uuid' => $model->uuid,
                    'material_type' => $materialType,
                    'material_uuid' => $value['product_sku_id'],
                    'estimate_num' => $value['estimate_purchase_num'],
                    'estimate_price' => $value['estimate_purchase_price'],
                    'estimate_amount' => floatval(helper::bcmul($value['estimate_purchase_price'], $value['estimate_purchase_num'], 2)),
                    'create_time' => $time,
                    'update_time' => $time,
                ];
            }

            $purchaseDetailModel = new ErpPurchaseDetail;
            $purchaseDetailModel->saveAll($dataList);
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
            foreach ($params['purchase_detail'] as $value) {
                // 判断价格格式为 小数点后2位，范围：0-100000000
                if (!preg_match('/^(0|[1-9]\d{0,8})(\.\d{1,2})?$/', $value['estimate_purchase_price']) || floatval($value['estimate_purchase_price']) > 100000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-100000000';
                    return false;
                }
                // 判断数量格式为 小数点后4位，范围：0-99999999
                if (!preg_match('/^(0|[1-9]\d{0,7})(\.\d{1,4})?$/', $value['estimate_purchase_num']) || floatval($value['estimate_purchase_num']) > 99999999) {
                    $this->error = '数量格式错误，请输入小数点后4位，范围：0-99999999';
                    return false;
                }
                $total_num += floatval($value['estimate_purchase_num']);
                $total_amount += floatval($value['estimate_purchase_price']) * floatval($value['estimate_purchase_num']);
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
            self::update($data, ['uuid' => $this['uuid']]);

            $dataList = [];
            $time = time();
            foreach ($params['purchase_detail'] as $value) {
                // 判断价格格式为 小数点后2位，范围：0-100000000
                if (!preg_match('/^(0|[1-9]\d{0,8})(\.\d{1,2})?$/', $value['estimate_purchase_price']) || floatval($value['estimate_purchase_price']) > 100000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-100000000';
                    return false;
                }
                // 判断数量格式为 小数点后4位，范围：0-99999999
                if (!preg_match('/^(0|[1-9]\d{0,7})(\.\d{1,4})?$/', $value['estimate_purchase_num']) || floatval($value['estimate_purchase_num']) > 99999999) {
                    $this->error = '数量格式错误，请输入小数点后4位，范围：0-99999999';
                    return false;
                }
                // 判断是商品还是材料
                $productBon = ProductBom::with(['product'])->where('uuid', $value['product_sku_id'])->find();
                if ($productBon) {
                    $materialType = 0;
                    $stockNum = $productBon->stock_num;
                    if (floatval(helper::bcadd($stockNum, $value['estimate_purchase_num'], 4)) > 99999999) {
                        $productName = extractLanguage($productBon->product->name);
                        $skuName = extractLanguage($productBon->name);
                        $this->error = "{$productName}({$skuName}) " . __('库存已超99999999');
                        return false;
                    }
                }
                $material = Material::where('uuid', $value['product_sku_id'])->find();
                if ($material) {
                    $materialType = 1;
                    $stockNum = $material->stock_num;
                    if (floatval(helper::bcadd($stockNum, $value['estimate_purchase_num'], 4)) > 99999999) {
                        $materialName = extractLanguage($material->name);
                        $this->error = "{$materialName} " . __('库存已超99999999');
                        return false;
                    }
                }
                $dataList[] = [
                    'purchase_form_uuid' => $this->uuid,
                    'material_type' => $materialType,
                    'material_uuid' => $value['product_sku_id'],
                    'estimate_num' => $value['estimate_purchase_num'],
                    'estimate_price' => $value['estimate_purchase_price'],
                    'estimate_amount' => floatval(helper::bcmul($value['estimate_purchase_price'], $value['estimate_purchase_num'], 2)),
                    'create_time' => $time,
                    'update_time' => $time,
                ];
            }

            $purchaseDetailModel = new ErpPurchaseDetail;
            foreach ($this->details as $detial) {
                $detial->delete();
            }
            $purchaseDetailModel->saveAll($dataList);
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
                // 判断价格格式为 小数点后2位，范围：0-100000000
                if (!preg_match('/^(0|[1-9]\d{0,8})(\.\d{1,2})?$/', $value['actual_purchase_price']) || floatval($value['actual_purchase_price']) > 100000000) {
                    $this->error = '价格格式错误，请输入小数点后2位，范围：0-100000000';
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
            $this->id = $this->getData('id');
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
                        $product = $detail->sku;
                    } else {
                        $product = $detail->material;
                    }
                    if (!$product) continue;

                    $hasInventoryAuth = (new Product())->hasInventoryAuth();
                    switch ($detail->material_type ) {
                        case ErpPurchaseDetail::MATERIAL_TYPE_PRODUCT:
                            ProductBom::where(['uuid' => $detail->material_uuid])->inc('stock_num', $detail->num)->update();
                            if ($hasInventoryAuth) {
                                /** @var ProductBom $productBom */
                                $productBom = ProductBom::where(['uuid' => $detail->material_uuid])->find();
                                ProductBom::addWarehouseInForm($productBom, 0, $operationLogData['operator_uuid'], $detail->num, $this->remark, $this->uuid);
                            }
                            break;
                        case ErpPurchaseDetail::MATERIAL_TYPE_MATERIAL:
                            $materialIds = array_merge($materialIds, [$detail->material_uuid]);
                            Material::where(['uuid' => $detail->material_uuid])->inc('stock_num', $detail->num)->update();
                            // 如果开启了进销存, 则创建"采购入库"记录
                            if ($hasInventoryAuth) {
                                /** @var Material $material */
                                $material = Material::where(['uuid' => $detail->material_uuid])->find();
                                Material::addWarehouseInForm($material, 0, $operationLogData['operator_uuid'], $detail->num, $this->remark, $this->uuid);
                            }
                            break;
                    }
                }
                // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
                if (!empty($materialIds)) {
                    $relatedMaterialUuidList = RelatedMaterial::whereIn('material_uuid', $materialIds)->column('uuid') ?: [];
                    RelatedMaterial::updateStock($relatedMaterialUuidList);
                }
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

        return Db::transaction(function () {
            $details = ErpPurchaseDetail::where('purchase_form_uuid', $this->uuid)->select();
            foreach ($details as $detail) {
                $detail->delete();
            }
            $logs = ErpPurchaseOperationLog::where('purchase_form_uuid', $this->uuid)->select();
            foreach ($logs as $log) {
                $log->delete();
            }
            return $this->destroy(['id' => $this->id]);
        });
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
