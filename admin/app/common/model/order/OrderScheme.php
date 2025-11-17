<?php

namespace app\common\model\order;

use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\store\TableArea;
use app\common\model\order\OrderSchemeArea;
use app\common\model\order\OrderSchemeProduct;

/**
 * 订单方案模型
 */
class OrderScheme extends BaseModel
{
    use SoftDelete;
    protected $pk = 'id';
    protected $name = 'product_must_plan';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    // 点餐方式、桌台方式
    const USE_CHANNEL_ORDER = 10; // 点餐方式
    const USE_CHANNEL_TABLE = 20; // 桌台方式

    // 必点规则-旧
    const OLD_MUST_RULE_MUST = 1; // 固定商品
    const OLD_MUST_RULE_OPTIONAL = 2; // 可选商品

    // 必点规则-新
    const NEW_MUST_RULE_MUST = 0; // 固定商品
    const NEW_MUST_RULE_OPTIONAL = 1; // 可选商品

    // 必点规则, 旧->新
    const OLD_MUST_RULE_MAP = [
        self::OLD_MUST_RULE_MUST => self::NEW_MUST_RULE_MUST,
        self::OLD_MUST_RULE_OPTIONAL => self::NEW_MUST_RULE_OPTIONAL,
    ];

    // 必点规则, 新-旧
    const NEW_MUST_RULE_MAP = [
        self::NEW_MUST_RULE_MUST => self::OLD_MUST_RULE_MUST,
        self::NEW_MUST_RULE_OPTIONAL => self::OLD_MUST_RULE_OPTIONAL,
    ];

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['name_text', 'table_area_ids', 'product_ids'];

    /**
     * 兼容字段
     */
    public function getIdAttr($value, $data)
    {
        return $this->uuid ?: 0;
    }

    /**
     * 名称
     */
    public static function getNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name'] ?? '');
    }

    /**
     * 获取使用渠道
     */
    public function getUseChannelAttr($value)
    {
        return $value ? explode(',', $value) : [];
    }

    /**
     * 获取桌台区域ids
     */
    public function getTableAreaIdsAttr($value, $data)
    {
        $list = OrderSchemeArea::alias('oa')
            ->field(['oa.desk_region_uuid'])
            ->where('oa.product_must_plan_uuid', $data['uuid'] ?? 0)
            ->select();

        return array_column($list->toArray(), 'desk_region_uuid');
    }

    /**
     * 获取商品ids列表
     */
    public function getProductIdsAttr($value, $data)
    {
        $list = OrderSchemeProduct::alias('op')
            ->leftJoin('product_package pp', 'pp.uuid = op.product_package_uuid')
            ->field(['op.product_package_uuid', 'pp.name as product_name'])
            ->where('op.product_must_plan_uuid', $data['uuid'] ?? 0)
            ->select();
        foreach ($list as &$item) {
            $item['product_name_text'] = extractLanguage($item['product_name']);
        }
        return $list->toArray();
    }

    /**
     * 设置使用渠道
     */
    public function setUseChannelAttr($value)
    {
        return is_array($value) ? implode(',', $value) : $value;
    }

    /**
     * 设置桌台区域ids
     */
    public function setTableAreaIdsAttr($value)
    {
        return is_array($value) ? implode(',', $value) : $value;
    }

    /**
     * 设置商品ids
     */
    public function setProductIdsAttr($value, $data)
    {
        return json_encode($value) ?: '';
    }

    /**
     * 设置必点规则
     */
    public function setMustRuleAttr($value)
    {
        return self::OLD_MUST_RULE_MAP[$value];
    }

    /**
     * 获取必点规则
     */
    public function getMustRuleAttr($value, $data)
    {
        return self::NEW_MUST_RULE_MAP[$value];
    }

    /**
     * 列表
     */
    public function getList($params, $app_id, $shop_supplier_id)
    {
        return $this->alias('a')
            ->leftJoin('product_must_plan_region pmpr', 'pmpr.product_must_plan_uuid = a.uuid')
            ->leftJoin('desk_region ta', 'ta.uuid = pmpr.desk_region_uuid')
            ->when($params['name'] ?? '', function ($query) use ($params) {
                $query->like('a.name', $params['name']);
            })
            ->when(isset($params['status']) && $params['status'] !== '', function ($query) use ($params) {
                $query->where('a.status', $params['status']);
            })
            ->field('a.*, IFNULL(GROUP_CONCAT(ta.name), "") as area_names')
            ->group('a.uuid')
            ->order('create_time', 'desc')
            ->paginate($params);
    }

    /**
     * 详情
     */
    public function detail($id)
    {
        return $this->where('uuid', $id)->find();
    }

    /**
     * 验证方案参数
     * @param array $params 参数
     * @param int $id 编辑时的ID
     * @return bool
     */
    private function validateSchemeParams(array $params, int $id = 0): bool
    {
        // 验证方案名称
        if ($params['name'] == '') {
            $this->error = '方案名称不能为空';
            return false;
        }
        if (mb_strlen($params['name']) > 50) {
            $this->error = '方案名称不能超过50个字符';
            return false;
        }

        // 使用渠道 10-点餐方式 20-桌台方式
        if (empty($params['use_channel'])) {
            $this->error = '使用渠道不能为空';
            return false;
        }
        $allowedChannels = [10, 20];
        $invalidChannels = array_diff($params['use_channel'], $allowedChannels);
        if (!empty($invalidChannels)) {
            $this->error = '使用渠道参数错误';
            return false;
        }


        // 桌台区域ids 判断是否存在
        if (!empty($params['table_area_ids'])) {
            $tableAreaIds = is_array($params['table_area_ids']) ? $params['table_area_ids'] : explode(',', $params['table_area_ids']);
            $tableAreaModel = new TableArea();
            $tableAreaList = $tableAreaModel->where('uuid', 'in', $tableAreaIds)->select();
            if (count($tableAreaList) != count($tableAreaIds)) {
                $this->error = '桌台区域参数错误';
                return false;
            }
        }

        // 必点类型 1-每人必点1份 2-每笔订单必点1份
        if (empty($params['must_type'])) {
            $this->error = '必点类型不能为空';
            return false;
        }
        if (!in_array($params['must_type'], [1, 2])) {
            $this->error = '必点类型参数错误';
            return false;
        }
        // 如果是点餐方式，则只能设置每笔订单必点1份；如果是桌台方式，则都可以设置
        if ($params['use_channel'][0] == 10 && $params['must_type'] != 2) {
            $this->error = '使用渠道不支持的必点类型';
            return false;
        }

        // 必点规则 1-固定商品 2-可选商品
        if (empty($params['must_rule'])) {
            $this->error = '必点规则不能为空';
            return false;
        }
        if (!in_array($params['must_rule'], [1, 2])) {
            $this->error = '必点规则参数错误';
            return false;
        }

        // 商品ids 判断是否存在
        if (!empty($params['product_ids'])) {
            $productIds = is_array($params['product_ids']) ? $params['product_ids'] : explode(',', $params['product_ids']);
            $productModel = new Product();
            $productList = $productModel->where('uuid', 'in', $productIds)->select();
            if (count($productList) != count($productIds)) {
                $this->error = '商品参数错误';
                return false;
            }
        }

        return true;
    }

    /**
     * 新增
     */
    public function add($params)
    {
        if (!$this->validateSchemeParams($params)) {
            return false;
        }

        $this->startTrans();
        try {
            $data = [
                'name' => $params['name'] ?? '',
                'use_channel' => $params['use_channel'] ?? '',
                'must_type' => $params['must_type'] ?? 1,
                'must_rule' => $params['must_rule'] ?? 1,
                'status' => $params['status'] ?? 1,
                'auto_cart' => $params['auto_cart'] ?? 1,
                'auto_change' => $params['auto_change'] ?? 1,
                'auto_check' => $params['auto_check'] ?? 1,
                'auto_checkout' => $params['auto_checkout'] ?? 1,
            ];
            $this->save($data);

            // 保存桌台区域
            if (!empty($params['table_area_ids'])) {
                $tableAreaIds = is_array($params['table_area_ids']) ? $params['table_area_ids'] : explode(',', $params['table_area_ids']);
                $tableAreaData = array_map(fn($id) => [
                    'product_must_plan_uuid' => $this->uuid,
                    'desk_region_uuid' => $id,
                ], $tableAreaIds);
                (new OrderSchemeArea)->saveAll($tableAreaData);
            }

            // 保存商品
            if (!empty($params['product_ids'])) {
                $productIds = is_array($params['product_ids']) ? $params['product_ids'] : explode(',', $params['product_ids']);
                $productData = array_map(fn($id) => [
                    'product_must_plan_uuid' => $this->uuid,
                    'product_package_uuid' => $id,
                ], $productIds);
                (new OrderSchemeProduct)->saveAll($productData);
            }

            $this->commit();
            return true;
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
        if (!$this->validateSchemeParams($params, $this['uuid'])) {
            return false;
        }

        $this->startTrans();
        try {
            $data = [
                'name' => $params['name'] ?? '',
                'use_channel' => $params['use_channel'] ?? '',
                'must_type' => $params['must_type'] ?? 1,
                'must_rule' => $params['must_rule'] ?? 1,
                'status' => $params['status'] ?? 1,
                'auto_cart' => $params['auto_cart'] ?? 1,
                'auto_change' => $params['auto_change'] ?? 1,
                'auto_check' => $params['auto_check'] ?? 1,
                'auto_checkout' => $params['auto_checkout'] ?? 1,
            ];

            $areaList = OrderSchemeArea::where('product_must_plan_uuid', $this->uuid)->select();
            foreach ($areaList as $area) {
                $area->force()->delete();
            }
            // 保存桌台区域
            if (!empty($params['table_area_ids'])) {
                $tableAreaIds = is_array($params['table_area_ids']) ? $params['table_area_ids'] : explode(',', $params['table_area_ids']);
                $tableAreaData = array_map(fn($id) => [
                    'product_must_plan_uuid' => $this->uuid,
                    'desk_region_uuid' => $id,
                ], $tableAreaIds);
                (new OrderSchemeArea)->saveAll($tableAreaData);
            }

            // 保存商品
            if (!empty($params['product_ids'])) {
                $productList = OrderSchemeProduct::where('product_must_plan_uuid', $this->uuid)->select();
                foreach ($productList as $product) {
                    $product->force()->delete();
                }
                $productIds = is_array($params['product_ids']) ? $params['product_ids'] : explode(',', $params['product_ids']);

                $numType0Count = Product::where('uuid', 'in', $productIds)->where('num_type', 0)->count();
                if ($numType0Count != count($productIds)) {
                    $this->error = '必点方案不能包含按小数计价商品';
                    return false;
                }

                $productData = array_map(fn($id) => [
                    'product_must_plan_uuid' => $this->uuid,
                    'product_package_uuid' => $id,
                ], $productIds);
                (new OrderSchemeProduct)->saveAll($productData);
            }

            $this->save($data);
            $this->commit();
            return true;
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
        $this->startTrans();
        try {
            $productList = OrderSchemeProduct::where('product_must_plan_uuid', $this->uuid)->select();
            foreach ($productList as $product) {
                $product->delete();
            }
            $areaList = OrderSchemeArea::where('product_must_plan_uuid', $this->uuid)->select();
            foreach ($areaList as $area) {
                $area->delete();
            }
            $this->delete();
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }

    /**
     * 设置状态
     */
    public function setStatus($params)
    {
        return $this->where('uuid', $params['id'])->update(['status' => $params['status']]);
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            'name' => $name,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('id') ? true : false;
    }
}
