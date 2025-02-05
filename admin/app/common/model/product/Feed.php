<?php

namespace app\common\model\product;

use help\StringHelp;
use think\facade\Db;
use app\common\model\BaseModel;

/**
 * 加料库模型
 */
class Feed extends BaseModel
{
    protected $name = 'feed';
    protected $pk = 'feed_id';

    /**
     * 处理多语言
     */
    protected $append = ['feed_name_text'];

    // 最大材料数量限制
    const MAX_MATERIAL_NUM = 99999999;

    /**
     * 加料名称
     *
     * @param [type] $value
     * @param array $data
     * @return string
     */
    public function getFeedNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['feed_name']);
    }

    /**
     * 关联产品ids
     */
    public function getProductIdsAttr($value, $data = [])
    {
        $product_ids = $data['product_ids'] ?? $value ?? '';
        if (empty($product_ids)) {
            return [];
        }
        $arr = array_map('intval', explode(',', $product_ids));
        return array_values($arr);
    }

    /**
     * 关联产品
     */
    public function productFeed($feed_id)
    {
        return $this->alias('feed')
            ->field('product.product_id')
            ->leftJoin('product_feed pf', 'feed.feed_id = pf.feed_id')
            ->leftJoin('product product', 'product.product_id = pf.product_id')
            ->where('product.is_delete', 0)
            ->where('feed.feed_id', $feed_id)
            ->select();
    }

    /**
     * 关联材料加料
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\ProductFeedMaterial', 'feed_id')->where('product_feed_id', '=', 0)->with(['materialProduct']);
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        $prefix = env('DB_PREFIX');
        return $this->alias('feed')
        ->with(['material'])
        ->field('feed.*')
        ->field("IF(pf.feed_count IS NULL, 0, 1) AS is_used")
        ->field("IFNULL(pf.product_ids, '') AS product_ids")
        ->leftJoin("
            (
                SELECT pf.feed_id, GROUP_CONCAT(DISTINCT product.product_id) AS product_ids, COUNT(DISTINCT pf.feed_id) AS feed_count
                FROM {$prefix}product_feed pf
                LEFT JOIN {$prefix}product product ON pf.product_id = product.product_id
                WHERE product.is_delete = 0
                GROUP BY pf.feed_id
            ) pf
        ", 'feed.feed_id = pf.feed_id')
        ->where('shop_supplier_id', '=', $shop_supplier_id)->order(['sort' => 'asc', 'create_time' => 'desc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($feed_id)
    {
        return self::find($feed_id);
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($feed_id)
    {
        // 兼容旧数据，先删除产品已删除的关联数据
        ProductFeed::where('product_id', 'in', function ($query) {
            $query->name('product')->where('is_delete', '=', 1)->field('product_id');
        })->delete();
        return ProductFeed::where('feed_id', 'in', $feed_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(feed_name, '$.$lang'))"), '=', $name],
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['feed_id', '<>', $id];
        }
        return static::where($filter)->value('feed_id') ? true : false;
    }

    /**
     * 维护产品表中的加料数组
     *
     * @param array $total_product_ids 产品ID数组
     */
    public function maintainProductFeed($total_product_ids, $delete_product_ids = [])
    {
        if (!empty($total_product_ids)) {
            $chunks = array_chunk($total_product_ids, 1000);
            foreach ($chunks as $chunk) {
                // 查询加料表
                $product_feeds = ProductFeed::with(['feed', 'material'])->whereIn('product_id', $chunk)->select()->toArray();
                // 格式化数据
                $product_feed_map = [];
                foreach ($product_feeds as $item) {
                    $product_id = $item['product_id'];
                    $feed_id = $item['feed_id'];
                    //
                    if (!isset($product_feed_map[$product_id])) {
                        $product_feed_map[$product_id] = [];
                    }
                    if (!isset($product_feed_map[$product_id][$feed_id])) {
                        $product_feed_map[$product_id][$feed_id] = [
                            'feed_id'        => $feed_id,
                            'feed_name'      => $item['feed']['feed_name'] ?? '',
                            'stock_num'      => $item['stock_num'] ?? '',
                            'price'          => $item['price'] ?? 0,
                            'default_select' => 0,
                            'uuid'           => StringHelp::getGuidV4(),
                            'material'       => $item['material'] ?? [],
                        ];
                    }
                }
                // 更新产品表
                $prefix = env('DB_PREFIX');
                $product = new Product;
                $product_ids = array_keys($product_feed_map);
                $product_feeds = array_values($product_feed_map);
                if (!empty($product_ids)) {
                    $update_sql = "UPDATE {$prefix}product SET product_feed = CASE product_id ";
                    foreach ($product_ids as $index => $product_id) {
                        $product_feed = json_encode(array_values($product_feeds[$index])) ?? '[]';
                        $product_feed = addslashes($product_feed); // 防止SQL注入并确保JSON数据正确转义
                        $update_sql .= "WHEN $product_id THEN '$product_feed' ";
                    }
                    $update_sql .= "END WHERE product_id IN (" . implode(',', $product_ids) . ")";
                    Db::connect($product->getConnection())->execute($update_sql);
                }
                // 如果有全部删除的产品ID，则清空对应的加料数组
                if (!empty($delete_product_ids)) {
                    $delete_product_ids = array_diff($delete_product_ids, $product_ids);
                    if (!empty($delete_product_ids)) {
                        $product->where('product_id', 'in', $delete_product_ids)->update(['product_feed' => '[]']);
                    }
                }
            }
        }
        return true;
    }
}
