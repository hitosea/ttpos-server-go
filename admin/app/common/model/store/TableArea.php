<?php


namespace app\common\model\store;

use app\common\model\BaseModel;

/**
 * 门店区域模型
 */
class TableArea extends BaseModel
{
    protected $name = 'desk_region';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['area_id', 'area_name'];

    /**
     * 兼容字段
     */
    public function getAreaIdAttr($value, $data)
    {
        return $this->uuid ?: 0;
    }
    public function getAreaNameAttr($value, $data)
    {
        return $this->getData('name') ?: '';
    }

    /**
     * 桌位详情
     */
    public static function detail($where)
    {
        $filter = is_array($where) ? $where : ['uuid' => $where];
        return static::where($filter)->find();
    }

    /**
     * 获取所有列表
     */
    public static function getAllList($shop_supplier_id)
    {
        return (new self)->order(['sort' => 'asc', 'create_time' => 'desc'])->select();
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

    /**
     * 获取区域列表
     */
    public static function getSucList()
    {
        return (new self)->field('uuid as area_id, name as area_name')->order(['sort' => 'asc', 'create_time' => 'desc'])->select()->toArray();
    }
}
