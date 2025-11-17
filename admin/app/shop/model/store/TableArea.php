<?php


namespace app\shop\model\store;

use app\common\service\websocket\Websocket;
use app\common\model\store\TableArea as TableAreaModel;

/**
 * 桌位区域模型
 */
class TableArea extends TableAreaModel
{

    const FORM_SCENE_ADD = 'add';
    const FORM_SCENE_EDIT = 'edit';

    /**
     * 分类更新后推送通知
     */
    public static function onAfterWrite(TableArea $model)
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
    public static function onAfterDelete(TableArea $model)
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
        if (isset($params['search']) && $params['search'] != '') {
            $model = $model->like('name', $params['search']);
        }
        // 查询列表数据
        $list = $model->order(['create_time' => 'desc'])->paginate($params);
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
        return Table::where('uuid', $areaId)->count() > 0;
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        $data['uuid'] = createUuid();
        $data['name'] = $data['area_name'] ?? '';
        return self::create($data);
    }

    /**
     * 编辑记录
     */
    public function edit($data)
    {
        $data['name'] = $data['area_name'] ?? '';
        return $this->save($data);
    }

    /**
     * 删除
     */
    public function setDelete()
    {
        if ($this->isAreaAssociatedWithTable($this['uuid'])) {
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
            $count = $this->where('name', '=', $data['area_name'])->count();
            if ($count) {
                $this->error = '桌号区域名称已存在';
                return false;
            }
        } else {
            $count = $this->where('name', '=', $data['area_name'])->where('uuid', '<>', $data['area_id'])->count();
            if ($count) {
                $this->error = '桌号区域名称已存在';
                return false;
            }
        }
        return true;
    }
}
