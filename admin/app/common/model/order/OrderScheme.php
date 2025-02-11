<?php

namespace app\common\model\order;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\model\product\Product;
use app\common\model\store\TableArea;

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

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['name_text'];

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
    public function getTableAreaIdsAttr($value)
    {
        return $value ? explode(',', $value) : [];
    }

    /**
     * 获取商品ids列表
     */
    public function getProductIdsAttr($value, $data)
    {
        $product_ids = $data['product_ids'] ? json_decode($data['product_ids'], true) : [];
        $list = Product::alias('product')
            ->field(['product.product_id', 'product.product_name'])
            ->where('product_id', 'in', $product_ids)
            ->where('is_delete', 0)
            ->where('product_status', 10)
            ->select()?->append(['product_name_text']);
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
        return $this->where('id', $id)->find();
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
        if ($this->checkNameExist($params['name'], $id ? $this['shop_supplier_id'] : $params['shop_supplier_id'], $id)) {
            $this->error = '方案名称已存在';
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
            $tableAreaList = $tableAreaModel->where('area_id', 'in', $tableAreaIds)->select();
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
            $productList = $productModel->where('product_id', 'in', $productIds)
                ->where('type', 10)
                ->where('is_delete', 0)
                ->select();
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

        $data = [
            'name' => $params['name'] ?? '',
            'use_channel' => $params['use_channel'] ?? '',
            'table_area_ids' => $params['table_area_ids'] ?? '',
            'must_type' => $params['must_type'] ?? 1,
            'must_rule' => $params['must_rule'] ?? 1,
            'product_ids' => $params['product_ids'] ?? '',
            'status' => $params['status'] ?? 1,
            'auto_cart' => $params['auto_cart'] ?? 1,
            'auto_change' => $params['auto_change'] ?? 1,
            'auto_check' => $params['auto_check'] ?? 1,
            'auto_checkout' => $params['auto_checkout'] ?? 1,
            'shop_supplier_id' => $params['shop_supplier_id'] ?? '',
            'app_id' => $params['app_id'] ?? ''
        ];

        return $this->save($data);
    }

    /**
     * 编辑
     */
    public function edit($params)
    {
        if (!$this->validateSchemeParams($params, $this['id'])) {
            return false;
        }

        return $this->save($params);
    }

    /**
     * 删除
     */
    public function del()
    {
        return $this->delete();
    }

    /**
     * 设置状态
     */
    public function setStatus($params)
    {
        return $this->where('id', $params['id'])->update(['status' => $params['status']]);
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
            $filter[] = ['id', '<>', $id];
        }
        return static::where($filter)->value('id') ? true : false;
    }
}
