<?php

namespace app\common\model\product;

use think\Model;
use help\StringHelp;
use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\service\websocket\Websocket;

/**
 * 加料库模型
 */
class Feed extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_sauce';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    // 最大材料数量限制
    const MAX_MATERIAL_NUM = 99999999;

    /**
     * 追加字段
     */
    protected $append = ['feed_id', 'feed_name', 'feed_name_text'];

    /**
     * 商品更新后推送通知
     */
    public static function onAfterWrite(Feed $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 商品删除后推送通知
     */
    public static function onAfterDelete(Feed $model)
    {
        $msgData = [
            'type' => 'delete',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 兼容字段
     */
    public function getFeedIdAttr()
    {
        return $this->uuid ?: 0;
    }
    public function getFeedNameAttr()
    {
        return $this->getData('name') ?: '';
    }

    /**
     * 多语言关联
     */
    public function multiLanguageName()
    {
        return $this->hasOne('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
    }

    /**
     * 规格关联材料
     */
    public function relatedMaterial()
    {
        return $this->hasMany('app\common\model\product\RelatedMaterial', 'related_uuid', 'uuid');
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
     * 加料名称
     *
     * @param [type] $value
     * @param array $data
     * @return string
     */
    public function getFeedNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
    }

    /**
     * 关联产品
     */
    public function productFeed($feed_id)
    {
        return $this->alias('feed')
            ->field('pb.product_package_uuid as product_id')
            ->leftJoin('product_bom pb', 'pb.product_sauce_uuid = feed.uuid')
            ->where('pb.delete_time', 0)
            ->where('feed.uuid', $feed_id)
            ->select();
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        return $this->order(['sort' => 'asc', 'create_time' => 'asc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($feed_id)
    {
        return self::where('uuid', $feed_id)->find();
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($feed_id)
    {
        return ProductBom::where('product_sauce_uuid', 'in', $feed_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null, $lang = 'zh')
    {
        $filter = [
            [Db::raw("JSON_UNQUOTE(JSON_EXTRACT(name, '$.$lang'))"), '=', $name],
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('uuid') ? true : false;
    }

    /**
     * todo 兼容 维护产品表中的加料数组
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
