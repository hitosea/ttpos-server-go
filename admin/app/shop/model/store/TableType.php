<?php


namespace app\shop\model\store;

use app\common\service\websocket\Websocket;
use app\common\model\store\TableType as TableTypeModel;

/**
 * 桌位类型模型
 */
class TableType extends TableTypeModel
{

    const FORM_SCENE_ADD = 'add';
    const FORM_SCENE_EDIT = 'edit';

   /**
     * 分类更新后推送通知
     */
    public static function onAfterWrite(TableType $model)
    {
        $msgData = [
            'type' => 'update',
            'type_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_DESK_TYPE, 0, $msgData);
    }

    /**
     * 分类删除后推送通知
     */
    public static function onAfterDelete(TableType $model)
    {   
        $msgData = [
            'type' => 'delete',
            'type_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_DESK_TYPE, 0, $msgData);
    }

    /**
     * 获取列表数据
     */
    public function getList($params, $shop_supplier_id)
    {
        $model = $this;
        if (!empty($search)) {
            $model = $model->like('name', $search);
        }
        // 查询列表数据
        $list = $model->order(['create_time' => 'desc'])->paginate($params);
        // 检查每个类型是否被桌台关联
        foreach ($list as &$type) {
            $isAssociated = $this->isTypeAssociatedWithTable($type['type_id']);
            $type['can_delete'] = !$isAssociated ? 1 : 0;
        }
        return $list;
    }

    /**
     * 检查区域是否被桌台关联
     */
    private function isTypeAssociatedWithTable($type_id)
    {
        return Table::where('type_uuid', $type_id)->count() > 0;
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
        $data['uuid'] = createUuid();
        $data['name'] = $data['type_name'] ?? '';
        $data['range_min'] = $data['min_num'] ?? 0;
        $data['range_max'] = $data['max_num'] ?? 0;
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
        $data['name'] = $data['type_name'] ?? '';
        $data['range_min'] = $data['min_num'] ?? 0;
        $data['range_max'] = $data['max_num'] ?? 0;
        return $this->save($data);
    }

    /**
     * 删除
     */
    public function setDelete()
    {
        if ($this->isTypeAssociatedWithTable($this['uuid'])) {
            $this->error = '该类型下存在桌台，不允许删除';
            return false;
        }
        return $this->delete();
    }

    /**
     * 表单验证
     */
    private function validateForm($data, $scene = self::FORM_SCENE_ADD)
    {
        if (strlen($data['type_name'] ?? '') > 50) {
            $this->error = '类型名称不得超过50字符';
            return false;
        }
        if ($data['min_num'] < 1 || $data['min_num'] > 100) {
            $this->error = '请输入1-100之间的数';
            return false;
        }
        if ($data['max_num'] < 1 || $data['max_num'] > 100) {
            $this->error = '请输入1-100之间的数';
            return false;
        }
        if ($data['min_num'] > $data['max_num']) {
            $this->error = '最多人数不可小于最少人数';
            return false;
        }
        return true;
    }
}
