<?php


namespace app\shop\model\store;

use app\common\model\order\Order;
use app\common\model\shop\BindRecord;
use app\common\enum\order\OrderStatusEnum;
use app\common\enum\order\OrderPayStatusEnum;
use app\common\model\store\Table as TableModel;

/**
 * 桌位模型
 */
class Table extends TableModel
{

    const FORM_SCENE_ADD = 'add';
    const FORM_SCENE_EDIT = 'edit';

    /**
     * 区域名称
     */
    public function getAreaNameAttr($value, $data)
    {
        return (new TableArea)->where('area_id', '=', $data['area_id'])->value('area_name');
    }

    /**
     * 类型名称
     */
    public function getTypeNameAttr($value, $data)
    {
        return (new TableType)->where('type_id', '=', $data['type_id'])->value('type_name');
    }

    /**
     * 获取列表数据
     */
    public function getList($params, $shop_supplier_id)
    {
        $model = $this;
        if (isset($params['area_id']) && $params['area_id']) {
            $model = $model->where('area_id', '=', $params['area_id']);
        }
        if (isset($params['type_id']) && $params['type_id']) {
            $model = $model->where('type_id', '=', $params['type_id']);
        }
        if (isset($params['search']) && $params['search'] != '') {
            $model = $model->like('table_no', $params['search']);
        }
        // 查询列表数据
        return $model->where('shop_supplier_id', '=', $shop_supplier_id)
            ->order(['create_time' => 'desc'])
            ->paginate($params);
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        $licenses = request()->licenses;
        $tl = ($licenses['z_l'] ?? 0);
        if ($tl != -1 && $this->where('shop_supplier_id', '=', $data['shop_supplier_id'])->count() >= $tl) {
            $this->error = '桌台数量已达上限，如有需要，请联系销售代表';
            return false;
        }
        // 表单验证
        if (!$this->validateForm($data, self::FORM_SCENE_ADD)) {
            return false;
        }
        $data = $this->sortData($data);
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
        $areaInfo = (new TableArea)->where('area_id', '=', $data['area_id'])->findOrEmpty();
        if ($areaInfo->isEmpty()) {
            $this->error = '区域不存在';
            return false;
        }
        $data = $this->sortData($data);
        // 同步修改订单
        if ($data['table_no'] ?? '') {
            $order = Order::where('table_id', $this->table_id)
                ->where("order_status", OrderStatusEnum::NORMAL)
                ->where('pay_status', OrderPayStatusEnum::PENDING)
                ->find();
            if ($order) {
                $order->table_no = $data['table_no'];
                $order->save();
            }
        }
        // 
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
        $model = $this->whereIn('table_id', $table_ids);
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
        (new BindRecord)->unbindByKey(BindRecord::SOURCE_TABLET, $this['bind_info']);
        return $this->save(['is_bind' => 0, 'bind_info' => '']);
    }

    /**
     * 表单验证
     */
    private function validateForm($data, $scene = self::FORM_SCENE_ADD)
    {
        if ($scene === self::FORM_SCENE_ADD) {
            //查询桌号是否存在
            $count = $this->where('shop_supplier_id', '=', $data['shop_supplier_id'])
                // ->where('area_id', '=', $data['area_id'])
                ->where('table_no', '=', $data['table_no'])
                ->count();
            if ($count) {
                $this->error = '桌号已存在';
                return false;
            }
        } else {
            $count = $this->where('shop_supplier_id', '=', $this['shop_supplier_id'])
                // ->where('area_id', '=', $data['area_id'])
                ->where('table_no', '=', $data['table_no'])
                ->where('table_id', '<>', $data['table_id'])
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
        $data['area_name'] = (new TableArea)->where('area_id', '=', $data['area_id'])->value('area_name');
        $typeInfo = (new TableType)->where('type_id', '=', $data['type_id'])->field('type_name,min_num,max_num')->find();
        $data['min_num'] = $typeInfo['min_num'];
        $data['max_num'] = $typeInfo['max_num'];
        $data['type_name'] = $typeInfo['type_name'];
        return $data;
    }

    /**
     * 桌台开关状态 0-关 1-开
     */
    public function setSwitchStatus($switch_status)
    {
        if ($this['status'] == 30 || $this['is_bind'] == 1) {
            $this->error = '当前桌位状态不允许该操作';
            return false;
        }
        return $this->save(['switch_status' => $switch_status]);
    }
}
