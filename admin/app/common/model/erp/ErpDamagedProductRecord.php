<?php

namespace app\common\model\erp;

use app\common\library\helper;
use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;
use app\common\model\product\Category;
use app\common\model\product\Material;
use app\common\model\product\ProductBom;
use app\common\model\erp\ErpWarehouseForm;
/**
 * 报损记录模型
 */
class ErpDamagedProductRecord extends BaseModel
{
    use SoftDelete;
    protected $name = 'loss_report_form';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['number', 'type', 'product_id', 'product_sku_id', 'review_status', 'operator_id', 'refused', 'approved_time', 'rejected_time'];

    /**
     * 报损类型, 0-损耗 1-丢失
     */
    const SCENE_LOSS = 0;
    const SCENE_LOST = 1;
    const OLD_SCENE_LOSS = [
        1 => self::SCENE_LOSS,
        2 => self::SCENE_LOST
    ];

    /**
     * 状态,0-pending待审核 1-approved已通过 2-rejected已驳回
     */
    const STATUS_PENDING = 0;
    const STATUS_APPROVED = 1;
    const STATUS_REJECTED = 2;

    /**
     * 兼容字段
     */
    public function getNumberAttr($value, $data = [])
    {
        return $this?->form_no ?: '';
    }
    public function getTypeAttr($value, $data = [])
    {
        return $this?->scene ?: 0;
    }
    public function getProductIdAttr($value, $data = [])
    {
        return $this?->material_uuid ?: 0;
    }
    public function getProductSkuIdAttr($value, $data = [])
    {
        return $this?->material_uuid ?: 0;
    }
    public function getReviewStatusAttr($value, $data = [])
    {
        return $this?->status ?: 0;
    }
    public function getRejectedTimeAttr($value, $data = [])
    {
        return $this?->status == self::STATUS_REJECTED ? date("Y-m-d H:i:s", intval($this?->update_time ?: 0)) : '';
    }
    public function getApprovedTimeAttr($value, $data = [])
    {
        return $this?->status == self::STATUS_APPROVED ? date("Y-m-d H:i:s", intval($this?->update_time ?: 0)) : '';
    }
    public function getOperatorIdAttr($value, $data = [])
    {
        return $this?->operator_uuid ?: 0;
    }
    public function getRefusedAttr($value, $data = [])
    {
        return $this?->reject_reason ?: 0;
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
        return $this->belongsTo(Product::class, 'material_uuid', 'uuid');
    }

    /**
     * 操作人
     */
    public function operator()
    {
        return $this->belongsTo(User::class, 'operator_uuid', 'uuid')->field(['uuid', 'uuid as shop_user_id', 'username as user_name', 'real_name']);
    }

    /**
     * 关联商品规格表
     */
    public function sku()
    {
        return $this->belongsTo('app\\common\\model\\product\\ProductBom', 'material_uuid', 'uuid');
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
        return $model
            ->with(['sku', 'operator'])
            ->when($type && $type > 0, function ($q) use ($type) {
                $q->where('scene', self::OLD_SCENE_LOSS[$type]);
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
            ->leftJoin('product_package p', 'dr.material_uuid = p.uuid')
            ->leftJoin('product_category c', 'p.category_uuid = c.uuid')
            ->where('dr.status', self::STATUS_APPROVED)
            ->when($type && $type > 0, function ($q) use ($type) {
                $q->where('dr.scene', $type);
            })
            ->when($start_time && $end_time, function ($q) use ($start_time, $end_time) {
                $q->where('dr.create_time', 'between', [strtotime($start_time), strtotime($end_time)]);
            })
            ->field([
                'IF(c.parent_uuid > 0, c.parent_uuid, p.category_uuid) AS category_uuid',
                'c.parent_uuid',
                'c.name',
                'SUM(num) as damage_count'
            ])
            ->group('p.category_uuid')
            ->order('p.category_uuid')
            ->select()->toArray();
        // 对结果进行处理
        foreach ($list as &$damageCount) {
            $category_uuid = $damageCount['category_uuid'];
            $damageCount['name'] = Category::getNameTextAttr($damageCount['name']);
            // 入库数
            $entry_num = (new ErpWarehouseForm())->alias('eir')
                ->leftJoin('product_package p', 'eir.material_uuid = p.uuid')
                ->leftJoin('product_category c', 'p.category_uuid = c.uuid')
                ->where(function ($q) use ($category_uuid) {
                    $q->where('c.uuid', $category_uuid)->whereOr('c.parent_uuid', $category_uuid);
                })
                ->where('eir.create_time', '<=', strtotime($end_time))
                ->where('eir.status', '=', ErpWarehouseForm::STATUS_SUCCESS)
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
        $product = ProductBom::where('uuid', $params['product_sku_id'])->find();
        $material = Material::where('uuid', $params['product_sku_id'])->find();

        $stock_num = 0;
        if ($product) {
            $stock_num = ProductBom::where('uuid', $params['product_sku_id'])->value('stock_num');
        }

        // 原料
        if ($material) {
            $stock_num = Material::where('uuid', $params['product_sku_id'])->value('stock_num');
        }

        if ($num > $stock_num) {
            $this->error = '不能大于库存数量';
            return false;
        }
        //
        $data['form_no'] = $model->generateNumber();
        $data['scene'] = self::OLD_SCENE_LOSS[$params['type'] ?? 1];
        $data['material_uuid'] = $params['product_sku_id'] ?? 0;
        $data['num'] = $num;
        $data['remark'] = $params['remark'] ?? 0;
        $data['status'] = self::STATUS_PENDING;
        $data['operator_uuid'] = $params['operator_id'];
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
        $updateArr['scene'] = self::OLD_SCENE_LOSS[$params['type'] ?? 1];
        $updateArr['material_uuid'] = $params['product_sku_id'] ?? 0;
        $updateArr['num'] = $params['num'] ?? 0;
        $updateArr['remark'] = $params['remark'] ?? 0;
        $updateArr['operator_uuid'] = $params['operator_id'];

        return $detail->save($updateArr);
    }

    /**
     * 审核
     */
    public function review($params)
    {
        if ($this->status != 0) {
            $this->error = '当前状态不可操作';
            return false;
        }
        $updateArr = [];
        $this->startTrans();
        try {
            if (($params['review_status'] ?? 0) == 1) {
                $updateArr['status'] = 1;
                $updateArr['approved_time'] = time();
                // 减少库存
                $product = ProductBom::where('uuid', $this->material_uuid)->find();
                $material = Material::where('uuid', $this->material_uuid)->find();
                // 成品
                if ($product) {
                    $productSku = ProductBom::where('uuid', $this->material_uuid)->find();
                    if ($this->num > $productSku->stock_num) {
                        $this->error = '报损数量大于剩余库存数量';
                        return false;
                    }
                    $skuStock = helper::bcsub($productSku->stock_num, $this->num);
                    $productSku->save(['stock_num' => $skuStock]);
                    $productStock = helper::bcsub($product->stock_num, $this->num);
                    $product->save(['stock_num' => $productStock]);
                }
                // 材料
                else if ($material) {
                    $productSku = Material::where('uuid', $this->material_uuid)->find();
                    if ($this['stock_num'] > $productSku->stock_num) {
                        $this->error = '报损数量大于剩余库存数量';
                        return false;
                    }
                    $skuStock = helper::bcsub($productSku->stock_num, $this->num, 4);
                    $productSku->save(['stock_num' => $skuStock]);
                    $productMaterialStock = helper::bcsub($product->stock_num, $this->num, 4);
                    $product->save(['stock_num' => $productMaterialStock]);
                } else {
                    $this->error = '记录不存在';
                    return false;
                }
            }
            if (($params['review_status'] ?? 0) == 2) {
                $updateArr['status'] = 2;
                $updateArr['rejected_time'] = time();
                $updateArr['reject_reason'] = $params['refused'] ?? '';
            }
            //
            $this->save($updateArr);
            $this->commit();

            // 更新跟材料相关的所有产品总库存、产品规格库存、加料库存
            if (($params['review_status'] ?? 0) == 1 && $material) {
                /** @var Product $product */
                $product->reCalProductStock([$this->material_uuid]);
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
