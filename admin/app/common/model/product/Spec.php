<?php

namespace app\common\model\product;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

/**
 * 规格/属性(组)模型
 */
class Spec extends BaseModel
{
    use SoftDelete;
    protected $name = 'product_flavor';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    /**
     * 追加字段
     */
    protected $append = ['spec_id', 'spec_name', 'spec_name_text'];

    /**
     * 兼容字段
     */
    public function getSpecIdAttr($value, $data)
    {
        return $this->uuid ?: 0;
    }
    public function getSpecNameAttr($value)
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
     * 规格名称
     * @param mixed $value
     * @param mixed $data
     * @return array|string
     */
    public function getSpecNameTextAttr($value, $data)
    {
        return extractLanguage($data['name']);
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
    public function productSku($spec_id)
    {
        return $this->alias('spec')
            ->field('pb.product_package_uuid as product_id')
            ->leftJoin('product_bom pb', 'pb.product_flavor_uuid = spec.uuid')
            ->where('pb.delete_time', 0)
            ->where('pb.product_flavor_uuid', $spec_id)
            ->select();
    }

    /**
     * 关联材料
     */
    public function material()
    {
        return $this->hasMany('app\\common\\model\\product\\ProductSkuMaterial', 'spec_id')->where('product_sku_id', '=', 0)->with(['materialProduct']);
    }

    /**
     * 更新规格组
     * @param mixed $data
     * @return void
     */
    public function updateSpec($data)
    {
        
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        $list = [];
        $flavors = $this->with(['multiLanguageName'])->order(['create_time' => 'desc'])->select();
        foreach ($flavors as $flavor) {
            $name = [
                'zh' => $flavor['multiLanguageName']['zh_name'],
                'zhtw' => $flavor['multiLanguageName']['zh_tw_name'],
                'en' => $flavor['multiLanguageName']['en_name'],
                'ja' => $flavor['multiLanguageName']['ja_name'],
                'ko' => $flavor['multiLanguageName']['ko_name'],
                'my' => $flavor['multiLanguageName']['my_name'],
                'th' => $flavor['multiLanguageName']['th_name'],
                'tr' => $flavor['multiLanguageName']['tr_name'],
                'sv' => $flavor['multiLanguageName']['sv_name'],
            ];
            $list[] = [
                'spec_id' => $flavor['uuid'],
                'spec_name' => json_encode($name),
                'spec_name_text' => extractLanguage(json_encode($name)),
                'id' => $flavor['id'],
                'uuid' => $flavor['uuid'],
                'name' => json_encode($name),
                'multi_language_name_uuid' => $flavor['multi_language_name_uuid'],
                'create_time' => $flavor['create_time'],
                'update_time' => $flavor['update_time'],
                'delete_time' => $flavor['delete_time'],
            ];
        }
        return $list;
    }

    /**
     * 获取列表数据
     */
    public function getLists($shop_supplier_id)
    {
        return $this->order(['create_time' => 'desc'])->select();
    }

    /**
     * 详情
     */
    public static function detail($spec_id)
    {
        return self::where('uuid', $spec_id)->find();
    }

    /**
     * 检查是否被关联
     */
    public function isUseWithProduct($spec_id)
    {
        return false;
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
