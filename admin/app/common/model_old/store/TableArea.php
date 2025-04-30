<?php


namespace app\common\model_old\store;

use app\common\model_old\BaseModel;

/**
 * 门店区域模型
 */
class TableArea extends BaseModel
{
    protected $pk = 'area_id';
    protected $name = 'table_area';

    /**
     * 关联门店
     */
    public function supplier()
    {
        return $this->BelongsTo('app\\common\\model_old\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id');
    }

    /**
     * 桌位详情
     */
    public static function detail($where)
    {
        $filter = is_array($where) ? $where : ['area_id' => $where];
        return static::where($filter)->find();
    }

    /**
     * 获取所有列表
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
            'area_name' => $name,
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['area_id', '<>', $id];
        }
        return static::where($filter)->value('area_id') ? true : false;
    }

    /**
     * 获取区域列表
     */
    public static function getSucList()
    {
        return (new self)->field(['area_id', 'area_name'])
            ->order(['sort' => 'asc', 'create_time' => 'desc'])
            ->select()->toArray();
    }
}
