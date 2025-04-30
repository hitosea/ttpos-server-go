<?php


namespace app\common\model_old\store;

use app\common\model_old\BaseModel;

/**
 * 门店类型模型
 */
class TableType extends BaseModel
{
    protected $pk = 'type_id';
    protected $name = 'table_type';

    /**
     * 关联门店
     */
    public function supplier()
    {
        return $this->BelongsTo('app\\common\\model_old\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id');
    }

    /**
     * 桌位类型详情
     */
    public static function detail($where)
    {
        $filter = is_array($where) ? $where : ['type_id' => $where];
        return static::where($filter)->find();
    }

    /**
     * 获取所有门店列表
     */
    public static function getAllList($shop_supplier_id)
    {
        return (new self)->where('shop_supplier_id', '=', $shop_supplier_id)
            ->order(['sort' => 'asc', 'create_time' => 'desc'])
            ->select();
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
    {
        $filter = [
            'type_name' => $name,
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['type_id', '<>', $id];
        }
        return static::where($filter)->value('type_id') ? true : false;
    }
}
