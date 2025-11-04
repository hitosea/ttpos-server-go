<?php


namespace app\shop\model\store;

use app\common\model\bill\SaleBill;
use app\common\model\shop\BindRecord;
use app\common\model\store\Table as TableModel;

/**
 * 桌位模型
 */
class Table extends TableModel
{

    const FORM_SCENE_ADD = 'add';
    const FORM_SCENE_EDIT = 'edit';

    /**
     * 获取列表数据
     */
    public function getList($params, $shop_supplier_id)
    {
        $model = $this->alias('a')
            ->field('a.*, a.uuid as org_uuid, dr.uuid as area_id, dr.name as area_name, dt.uuid as type_id, dt.name as type_name, dt.range_min as min_num, dt.range_max as max_num')
            ->leftJoin('desk_region dr', 'dr.uuid = a.region_uuid')
            ->leftJoin('desk_type dt', 'dt.uuid = a.type_uuid');

        if (isset($params['area_id']) && $params['area_id']) {
            $model = $model->where('region_uuid', '=', $params['area_id']);
        }
        if (isset($params['type_id']) && $params['type_id']) {
            $model = $model->where('type_uuid', '=', $params['type_id']);
        }
        if (isset($params['search']) && $params['search'] != '') {
            $model = $model->like('desk_no', $params['search']);
        }
        // 查询列表数据
        return $model->order(['a.id' => 'desc'])->paginate($params);
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        $licenses = request()->licenses;
        $tl = ($licenses['z_l'] ?? 0);
        if ($tl != -1 && $this->count() >= $tl) {
            $this->error = '桌台数量已达上限，如有需要，请联系销售代表';
            return false;
        }
        $data = $this->sortData($data);
        $data['uuid'] = createUuid();
        $data['is_disable'] = 0;
        return self::create($data);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        $areaInfo = (new TableArea)->where('uuid', '=', $data['area_id'])->findOrEmpty();
        if ($areaInfo->isEmpty()) {
            $this->error = '区域不存在';
            return false;
        }
        $data = $this->sortData($data);
        return $this->save($data);
    }

    /**
     * 删除
     */
    public function setDelete()
    {
        if ($this['status'] == 30) {
            $this->error = '当前桌位已开台，不允许该操作';
            return false;
        }
        return $this->delete();
    }

    /**
     * 批量删除
     * @param $table_ids
     */
    public function batchDelete($table_ids)
    {
        $model = $this->whereIn('uuid', $table_ids);
        $list = $model->select();
        if ($list->isEmpty()) {
            $this->error = '数据不存在';
            return false;
        }
        foreach ($list as $table) {
            if ($table['status'] == 30) {
                $this->error = '当前桌位已开台，不允许该操作';
                return false;
            }
        }
        return $model->delete();
    }

    /**
     * 解绑
     */
    public function setUnbind()
    {
        if ($this['status'] == 30) {
            $this->error = '当前桌位已开台，不允许该操作';
            return false;
        }
        // 解绑设备记录
        (new BindRecord)->unbindByKey(BindRecord::SOURCE_TABLET, $this['device_uuid']);
        return $this->save(['is_bind' => 0, 'device_uuid' => '']);
    }

    /**
     * 表单验证
     */
    private function validateForm($data, $scene = self::FORM_SCENE_ADD)
    {
        if ($scene === self::FORM_SCENE_ADD) {
            //查询桌号是否存在
            $count = $this->where('desk_no', '=', $data['table_no'])->count();
            if ($count) {
                $this->error = '桌号已存在';
                return false;
            }
        } else {
            $count = $this->where('desk_no', '=', $data['table_no'])
                ->where('uuid', '<>', $data['table_id'])
                ->count();
            if ($count) {
                $this->error = '桌号已存在';
                return false;
            }
        }
        return true;
    }

    /**
     * 整理数据
     * @param mixed $data
     * @return mixed
     */
    public function sortData($data)
    {
        $data['area_name'] = (new TableArea)->where('uuid', '=', $data['area_id'])->value('name');
        $typeInfo = (new TableType)->where('uuid', '=', $data['type_id'])->field('name, range_min, range_max')->find();
        $data['min_num'] = $typeInfo['range_min'];
        $data['max_num'] = $typeInfo['range_max'];
        $data['type_name'] = $typeInfo['name'];
        //
        $data['desk_no'] = $data['table_no'] ?? '';
        $data['region_uuid'] = $data['area_id'] ?? 0;
        $data['type_uuid'] = $data['type_id'] ?? 0;
        return $data;
    }

    /**
     * 桌台开关状态 0-关 1-开
     */
    public function setSwitchStatus($switch_status)
    {
        if ($this['status'] == 1 || $this['device_uuid'] > 0) {
            $this->error = '当前桌位状态不允许该操作';
            return false;
        }
        $bill = SaleBill::where('desk_uuid', $this['uuid'])->where('status', 0)->find();
        if ($bill) {
            $this->error = '桌台使用中，无法禁用';
            return false;
        }
        return $this->save(['is_disable' => $switch_status == 0 ? 1 : 0]);
    }
}
