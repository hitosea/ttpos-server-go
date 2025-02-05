<?php


namespace app\shop\model\store;

use app\common\model\store\TableArea as TableAreaModel;

/**
 * 桌位区域模型
 */
class TableArea extends TableAreaModel
{

    const FORM_SCENE_ADD = 'add';
    const FORM_SCENE_EDIT = 'edit';

    /**
     * 获取列表数据
     */
    public function getList($params, $shop_supplier_id)
    {
        $model = $this;
        if (isset($params['search']) && $params['search'] != '') {
            $model = $model->like('area_name', $params['search']);
        }
        // 查询列表数据
        $list = $model->where('shop_supplier_id', '=', $shop_supplier_id)
            ->order(['create_time' => 'desc'])
            ->paginate($params);
        // 检查每个区域是否被桌台关联
        foreach ($list as &$area) {
            $isAssociated = $this->isAreaAssociatedWithTable($area['area_id']);
            $area['can_delete'] = !$isAssociated ? 1 : 0;
        }
        return $list;
    }

    /**
     * 检查区域是否被桌台关联
     */
    private function isAreaAssociatedWithTable($areaId)
    {
        return Table::where('area_id', $areaId)->count() > 0;
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        // 表单验证
        if (!$this->validateForm($data, self::FORM_SCENE_ADD)) {
            return false;
        }
        $data['app_id'] = self::$app_id;
        return self::create($data);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        // 表单验证
        if (!$this->validateForm($data, self::FORM_SCENE_EDIT)) {
            return false;
        }
        return $this->save($data);
    }

    /**
     * 删除
     */
    public function setDelete()
    {
        if ($this->isAreaAssociatedWithTable($this['area_id'])) {
            $this->error = '当前区域下存在桌台，不允许删除';
            return false;
        }
        return $this->delete();
    }

    /**
     * 表单验证
     */
    private function validateForm($data, $scene = self::FORM_SCENE_ADD)
    {
        if ($scene === self::FORM_SCENE_ADD) {
            //查询桌号是否存在
            $count = $this->where('shop_supplier_id', '=', $data['shop_supplier_id'])
                ->where('area_name', '=', $data['area_name'])
                ->count();
            if ($count) {
                $this->error = '桌号区域名称已存在';
                return false;
            }
        } else {
            $count = $this->where('shop_supplier_id', '=', $this['shop_supplier_id'])
                ->where('area_name', '=', $data['area_name'])
                ->where('area_id', '<>', $data['area_id'])
                ->count();
            if ($count) {
                $this->error = '桌号区域名称已存在';
                return false;
            }
        }
        return true;
    }
}
