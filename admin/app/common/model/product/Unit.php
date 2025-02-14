<?php

namespace app\common\model\product;

use think\facade\Db;
use app\common\model\BaseModel;
use think\model\concern\SoftDelete;

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
     * 多语言数据处理
     * @param mixed $model
     * @return void
     */
    public static function onBeforeDelete($model)
    {
        dump(1111111);
        die;
        if ($model instanceof \think\Collection) {
            $uuids = $model->column('multi_language_name_uuid');
            if (!empty($uuids)) {
                $model->multiLanguageName()->where('uuid', 'in', $uuids)->delete();
            }
        } else {
            if ($model->multi_language_name_uuid) {
                $model->multiLanguageName()->delete();
            }
        }
    }

    /**
     * 多语言关联
     */
    public function multiLanguageName()
    {
        return $this->hasOne('app\common\model\store\MultiLanguageName', 'uuid', 'multi_language_name_uuid');
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
        return $this->product()->column('uuid');
    }

    /**
     * 关联产品
     */
    public function product()
    {
        return $this->hasMany('app\\common\\model\\product\\Product', 'unit_uuid', 'uuid');
    }

    /**
     * 更新单位
     * @param mixed $unit_name
     * @param mixed $shop_supplier_id
     * @return void
     */
    public function updateUnit($unit_name, $shop_supplier_id)
    {
        // todo 不用再次校验单位新增，单位添加是已固定单位，先注释掉
        // if ($unit_name) {
        //     $isExit = $this->where('unit_name', '=', $unit_name)->count();
        //     if ($isExit == 0) {
        //         $addData = [
        //             'unit_name'        => $unit_name,
        //         ];
        //         $this->save($addData);
        //     }
        // }
    }

    /**
     * 获取列表数据
     */
    public function getAllList($shop_supplier_id)
    {
        return $this->order(['sort' => 'asc', 'create_time' => 'desc'])->select()?->append(['product_ids'], true);
    }

    /**
     * 获取列表数据
     */
    public function getLists($shop_supplier_id)
    {
        return $this->order(['sort' => 'asc', 'create_time' => 'desc'])->select();
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
