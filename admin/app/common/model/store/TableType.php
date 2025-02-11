<?php


namespace app\common\model\store;

use app\common\model\BaseModel;

/**
 * 门店类型模型
 */
class TableType extends BaseModel
{
    protected $name = 'desk_type';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['type_id', 'type_name', 'min_num', 'max_num'];

    /**
     * 兼容字段
     */
    public function getTypeIdAttr($value, $data)
    {
        return $this->uuid ?: 0;
    }
    public function getTypeNameAttr($value, $data)
    {
        return $this->getData('name') ?: '';
    }
    public function getMinNumAttr($value, $data)
    {
        return $this->range_min ?: 0;
    }
    public function getMaxNumAttr($value, $data)
    {
        return $this->range_max ?: 0;
    }

    /**
     * 桌位类型详情
     */
    public static function detail($where)
    {
        $filter = is_array($where) ? $where : ['uuid' => $where];
        return static::where($filter)->find();
    }

    /**
     * 获取所有门店列表
     */
    public static function getAllList($shop_supplier_id)
    {
        return (new self)->order(['sort' => 'asc', 'create_time' => 'desc'])
            ->select();
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
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('id') ? true : false;
    }
}
