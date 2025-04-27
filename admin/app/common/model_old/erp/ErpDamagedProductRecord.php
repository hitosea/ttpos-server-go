<?php

namespace app\common\model_old\erp;

use help\StringHelp;
use app\common\library\helper;
use app\common\model_old\BaseModel;
use app\common\model_old\shop\User;
use think\model\concern\SoftDelete;
use app\common\model_old\product\Product;
use app\common\model_old\product\Category;
use app\common\model_old\product\ProductSku;
use app\common\service\sync\SyncService;

/**
 * 报损记录模型
 */
class ErpDamagedProductRecord extends BaseModel
{
    use SoftDelete;
    protected $name = 'erp_damaged_product_record';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = [];

    /**
     * type 损坏类型 1-丢失 2-损坏
     */
    const TYPE_LOST = 1;
    const TYPE_DAMAGED = 2;

    /**
     * 审核类型 0-待审核 1-通过 2-拒绝
     */
    const REVIEW_PENDING = 0;
    const REVIEW_APPROVED = 1;
    const REVIEW_REJECTED = 2;

    //
    public static function onBeforeInsert($model)
    {
        if (!isset($model['id'])) {
            $model['id'] = StringHelp::uuid();
        }
        return $model;
    }

    public static function getApprovedTimeAttr($value, $data = [])
    {
        return date("Y-m-d H:i:s", $data['approved_time'] ?? 0);
    }

    public static function getRejectedTimeAttr($value, $data = [])
    {
        return date("Y-m-d H:i:s", $data['rejected_time'] ?? 0);
    }

    /**
     * 报损数量
     */
    public function getNumAttr($value)
    {
        return floatval($value);
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->belongsTo(Product::class, 'product_id', 'product_id');
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_id', 'shop_user_id')->field(['shop_user_id', 'user_name', 'real_name']);
    }

    /**
     * 关联商品规格表
     */
    public function sku()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductSku', 'product_sku_id', 'product_sku_id');
    }


    /**
     * 获取列表
     *
     * @param [type] $params
     * @return object
     */
    public function getList($params)
    {
        $model = new self;
        $type = null;
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }

        if (isset($params['type']) && $params['type']) {
            $type = $params['type'];
        }
        return $model->with(['sku.product', 'operator'])
            ->when($type && $type > 0, function ($q) use ($type) {
                $q->where('type', $type);
            })
            ->when($start_time && $end_time, function ($q) use ($start_time, $end_time) {
                $q->where('create_time', 'between', [strtotime($start_time), strtotime($end_time)]);
            })
            ->order('create_time desc')->paginate($params);
    }

    /**
     * 报损记录柱状图
     * @param $params
     * @return mixed
     */
    public function getChartList($params)
    {
        $model = new self;
        $type = null;
        if ($params['date']) {
            $timestamp = strtotime($params['date']);
            $year = date("Y", $timestamp);
            $month = date("m", $timestamp);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        } else {
            $now = time();
            $year = date("Y", $now);
            $month = date("m", $now);
            $start_time = date($year . "-" . $month . "-01");
            $end_time = date($year . "-" . $month . "-t 23:59:59");
        }
        if (isset($params['type']) && $params['type']) {
            $type = $params['type'];
        }
        $list = $model->alias('dr')
            ->join('product p', 'dr.product_id = p.product_id')
            ->join('category c', 'p.category_id = c.category_id')
            ->where('dr.delete_time', 0)
            ->where('dr.review_status', 1)
            ->when($type && $type > 0, function ($q) use ($type) {
                $q->where('dr.type', $type);
            })
            ->when($start_time && $end_time, function ($q) use ($start_time, $end_time) {
                $q->where('dr.create_time', 'between', [strtotime($start_time), strtotime($end_time)]);
            })
            ->field([
                'IF(c.parent_id > 0, c.parent_id, p.category_id) AS category_id',
                'c.parent_id',
                'c.name',
                'SUM(num) as damage_count'
            ])
            ->group('p.category_id')
            ->order('p.category_id')
            ->select()->toArray();
        // 对结果进行处理
        foreach ($list as &$damageCount) {
            $category_id = $damageCount['category_id'];
            $damageCount['name'] = Category::getNameTextAttr($damageCount['name']);
            // 入库数
            $entry_num = (new ErpInventoryRecord())->alias('eir')
                ->join('product p', 'eir.product_id = p.product_id')
                ->join('category c', 'p.category_id = c.category_id')
                ->where('eir.delete_time', 0)
                ->where(function ($q) use ($category_id) {
                    $q->where('c.category_id', $category_id)->whereOr('c.parent_id', $category_id);
                })
                ->where('eir.create_time', '<=', strtotime($end_time))
                ->where('eir.status', '=', 10)  // 10-已入库 20-已出库
                ->sum('num');
            // 出库数
            $exit_num = 0;
            $category_stock = helper::bcsub($entry_num, $exit_num);
            $damageCount['damage_ratio'] = $category_stock > 0 ? helper::bcdiv($damageCount['damage_count'], $category_stock, 5) : 0;
            $damageCount['damage_ratio'] = floatval($damageCount['damage_ratio'] * 100) . '%';
            $damageCount['damage_count'] = floatval($damageCount['damage_count']);
        }
        return $list;
    }

    /**
     * 详情
     */
    public function detail($id): self
    {
        $model = new self;
        return $model->with(['product', 'operator'])
            ->where('id', $id)
            ->find();
    }

    /**
     * 新增
     */
    public function add($params)
    {
        $model = new self;
        $num = $params['num'] ?? 0;
        if (!$num) {
            $this->error = '数量不能为0';
            return false;
        }

        $product = Product::where('product_id', $params['product_id'])->find();
        $stock_num = 0;
        if ($product['type'] == Product::TYPE_PRODUCT) {
            $stock_num = ProductSku::where('product_sku_id', $params['product_sku_id'])->value('stock_num');
        }
        if ($product['type'] == Product::TYPE_MATERIAL) {
            $stock_num = ProductSku::where('product_sku_id', $params['product_sku_id'])->value('material_stock');
        }

        if ($num > $stock_num) {
            $this->error = '不能大于库存数量';
            return false;
        }
        //
        $data['number'] = $model->generateNumber();
        $data['type'] = $params['type'] ?? 1;;
        $data['product_id'] = $params['product_id'] ?? 0;
        $data['product_sku_id'] = $params['product_sku_id'] ?? 0;
        $data['num'] = $num;
        $data['remark'] = $params['remark'] ?? 0;
        $data['review_status'] = 0;
        $data['operator_id'] = $params['operator_id'];
        $data['shop_supplier_id'] = $params['shop_supplier_id'] ?? 0;
        $data['app_id'] = self::$app_id;
        $model->save($data);
        return $model->id;
    }

    /**
     * 编辑更改
     */
    public function edit($params)
    {
        $detail = (new self)->detail($params['id'] ?? 0);
        if (!$detail) {
            $this->error = '记录不存在';
            return false;
        }
        //
        $updateArr['type'] = $params['type'] ?? 1;;
        $updateArr['product_id'] = $params['product_id'] ?? 0;
        $updateArr['num'] = $params['num'] ?? 0;
        $updateArr['remark'] = $params['remark'] ?? 0;
        $updateArr['operator_id'] = $params['operator_id'];

        return $detail->save($updateArr);
    }

    /**
     * 云端同步操作
     *
     */
    public function syncReview($params)
    {
        $id = $this->id;
        $reviewStatus = $params['review_status'] ?? 0;
        $refused = $params['refused'] ?? '';
        if (!$id) {
            $this->error = '报损单不存在';
            return false;
        }
        //
        $syncService = new SyncService();
        //
        $syncDetailData = $syncService->syncInventoryDetail($id);
        if (!$syncDetailData) {
            $this->error = $syncService->getError() ?: '网络连接异常，请重试';
            return false;
        } else if (($syncDetailData['code'] ?? 0) == 502) {
            $syncService->syncTables();
        } else {
            $syncDetail = $syncDetailData['data'] ?? [];
            if (!$syncDetail) {
                $this->error = '云端数据异常，请重试';
                return false;
            }
            // 状态对比，如果状态不同，则更新本地状态
            if ($this['review_status'] != $syncDetail['review_status']) {
                $this->review([
                    'review_status' => $syncDetail['review_status'],
                    'refused' => $syncDetail['refused'],
                ]);
            }
            //
            if ($syncDetail['review_status'] == $reviewStatus) {
                return true;
            }
        }
        // 状态对比，如果状态相同，先请求云上操作，再更新本地操作方法
        $syncOperateData = $syncService->syncInventoryReview($id, $reviewStatus, $refused);
        if (!$syncOperateData) {
            $this->error = $syncService->getError() ?: '网络连接异常，请重试';
            return false;
        }
        if ($syncOperateData['code'] != 1) {
            $this->error = $syncOperateData['msg'] ?? '操作失败';
            return false;
        }
        return $this->review($params);
    }

    /**
     * 审核
     */
    public function review($params)
    {
        if ($this->review_status != 0) {
            $this->error = '当前状态不可操作';
            return false;
        }
        $updateArr = [];
        $this->startTrans();
        try {
            if (($params['review_status'] ?? 0) == 1) {
                $updateArr['review_status'] = 1;
                $updateArr['approved_time'] = time();
                // 减少库存
                $product = Product::where('product_id', $this->product_id)->find();
                // 成品
                if ($product['type'] == Product::TYPE_PRODUCT) {
                    $productSku = ProductSku::where('product_sku_id', $this->product_sku_id)->find();
                    if ($this->num > $productSku->stock_num) {
                        $this->error = '报损数量大于剩余库存数量';
                        return false;
                    }
                    $skuStock = helper::bcsub($productSku->stock_num, $this->num);
                    $productSku->save(['stock_num' => $skuStock]);
                    $productStock = helper::bcsub($product->product_stock, $this->num);
                    $product->save(['product_stock' => $productStock]);
                }
                // 材料
                else if ($product['type'] == Product::TYPE_MATERIAL) {
                    $productSku = ProductSku::where('product_sku_id', $this['product_sku_id'])->find();
                    if ($this['num'] > $productSku->material_stock) {
                        $this->error = '报损数量大于剩余库存数量';
                        return false;
                    }
                    $skuStock = helper::bcsub($productSku->material_stock, $this->num, 4);
                    $productSku->save(['material_stock' => $skuStock]);
                    $productMaterialStock = helper::bcsub($product->product_material_stock, $this->num, 4);
                    $product->save(['product_material_stock' => $productMaterialStock]);
                } else {
                    $this->error = '记录不存在';
                    return false;
                }
            }
            if (($params['review_status'] ?? 0) == 2) {
                $updateArr['review_status'] = 2;
                $updateArr['rejected_time'] = time();
                $updateArr['refused'] = $params['refused'] ?? '';
            }
            //
            $this->save($updateArr);
            $this->commit();

            // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
            if (($params['review_status'] ?? 0) == 1 && $product['type'] == Product::TYPE_MATERIAL) {
                $product->reCalProductStock([$this->product_id]);
            }
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 根据规格汇总报损数量
     */
    public function sumDamagedProductNum($product_sku_id)
    {
        return (new self())->where('product_sku_id', $product_sku_id)->sum('num');;
    }

    /**
     * 入库编号：18位纯数字（前2位WT，2-10位是年月日，中间位是0000，后4位随机生成）
     *
     * @return string
     */
    public function generateNumber()
    {
        $date = date('Ymd');
        $rand = rand(1000, 9999);
        $code = 'WT' . $date . '0000' . $rand;
        return $code;
    }
}
