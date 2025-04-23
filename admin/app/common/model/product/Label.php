<?php

namespace app\common\model\product;

use app\common\model\BaseModel;
use app\common\service\websocket\Websocket;

/**
 * 打印标签模型
 */
class Label extends BaseModel
{
    protected $name = 'printer_tag';
    protected $pk = 'id';

    /**
     * 追加字段
     */
    protected $append = ['label_id', 'label_name', 'label_name_text'];

    /**
     * 商品更新后推送通知
     */
    public static function onAfterWrite(Label $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_KITCHEN, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 商品删除后推送通知
     */
    public static function onAfterDelete(Label $model)
    {
        $msgData = [
            'type' => 'update',
            'product_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_KITCHEN, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);
    }

    /**
     * 兼容字段
     */
    public function getLabelIdAttr()
    {
        return $this->uuid ?: 0;
    }
    public function getLabelNameAttr()
    {
        return $this->getData('name') ?: '';
    }

    /**
     * 标签名称
     */
    public function getLabelNameTextAttr($value, $data = [])
    {
        return extractLanguage($value ?: $data['name']);
    }

    /**
     * 关联产品ids
     */
    // public function getProductIdsAttr($value, $data = [])
    // {
    //     return $this->product()->column('uuid');
    // }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->hasMany(Product::class, 'printer_tag_uuid', 'uuid');
    }

    /**
     * 更新标签
     * @param mixed $data
     * @return void
     */
    public function updateLabel($data)
    {
        if ($data) {
            $addData = [];
            foreach ($data as $item) {
                $isExit = $this->where('name', '=', $item['label_name'])->count();
                if ($isExit == 0) {
                    $addData[] = [
                        'name' => $item['label_name'],
                    ];
                }
            }
            $addData && $this->saveAll($addData);
        }
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        $list = $this->with(['product'])->order(['create_time' => 'desc'])->select();
        foreach ($list as $item) {
            $product_ids = [];
            foreach ($item->product as $product) {
                $product_ids[] = $product->uuid;
            }
            $item->product_ids = $product_ids;
            unset($item->product);
        }

        return $list;
    }

    /**
     * 详情
     */
    public static function detail($label_id)
    {
        return self::where('uuid', $label_id)->find();
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($label_id)
    {
        return Product::where('printer_tag_uuid', 'in', $label_id)->count() > 0;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
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
