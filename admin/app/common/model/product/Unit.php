<?php

namespace app\common\model\product;

use think\facade\Db;
use think\facade\Env;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;
use app\common\service\websocket\Websocket;

/**
 * 单位模型
 */
class Unit extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_unit';
    protected $pk   = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     */
    protected $append = ['unit_id', 'unit_name', 'unit_name_text'];

    /**
     * 商品更新后推送通知
     */
    public static function onAfterWrite(Unit $model)
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
    public static function onAfterDelete(Unit $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }
    
    /**
     * 兼容字段
     */
    public function getUnitIdAttr()
    {
        return $this->uuid ?: 0;
    }
    public function getUnitNameAttr()
    {
        return $this->getData('name') ?: '';
    }

    /**
     * 单位名称
     */
    public function getUnitNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
    }

    /**
     * 关联产品ids
     */
    public function getProductIdsAttr($value, $data = [])
    {
        $prefix = Env::get('DB_PREFIX');
        $dbName = 'shop' . request()->appId;
        $results = Db::connect($dbName)->query("SELECT uuid FROM (
            SELECT uuid FROM " . $prefix . "product_package
            WHERE unit_uuid = ? AND delete_time = 0
            UNION ALL
            SELECT uuid FROM " . $prefix . "material
            WHERE unit_uuid = ? AND delete_time = 0
        ) as product_ids", [$this->uuid, $this->uuid]);

        return array_column($results, 'uuid');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->hasMany('app\\common\\model\\product\\Product', 'unit_uuid', 'uuid');
    }

    /**
     * 关联原料
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\Material', 'unit_uuid', 'uuid');
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        return $this->order(['create_time' => 'desc'])->select()?->append(['product_ids'], true);
    }

    /**
     * 获取列表数据
     */
    public function getLists($shop_supplier_id)
    {
        return $this->order(['id' => 'asc', 'create_time' => 'desc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($unit_id)
    {
        return self::where('uuid', $unit_id)->find();
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($unit_id)
    {
        return Product::where('unit_uuid', 'in', $unit_id)->count() > 0;
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
}
