<?php

namespace app\common\model\file;

use app\common\model\BaseModel;
use app\shop\model\file\UploadFile;

/**
 * 文件库分组模型
 */
class UploadGroup extends BaseModel
{
    protected $name = 'file_group';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['group_id'];

    /**
     * 兼容ID字段
     */
    public function getGroupIdAttr()
    {
        return $this->uuid ?? 0;
    }

    /**
     * 分组详情
     */
    public static function detail($group_id, $shop_supplier_id = 0)
    {
        return self::where('uuid', '=', $group_id)->find();
    }

    /**
     * 获取列表记录
     */
    public function getList($groupType, $shop_supplier_id = 0)
    {
        !empty($groupType) && $this->where('group_type', '=', trim($groupType));
        return $this
            ->field('*, uuid as group_id')
            ->where('group_type', '=', trim($groupType))
            ->order(['sort' => 'asc', 'create_time' => 'desc'])
            ->select();
    }

    /**
     * 添加新记录
     */
    public function add($data)
    {
        return $this->save(array_merge([
            'uuid'=> createUuid(),
            'sort' => 100
        ], $data));
    }

    /**
     * 更新记录
     */
    public function edit($data)
    {
        return $this->save($data) !== false;
    }

    /**
     * 删除记录
     */
    public function remove()
    {
        // 更新该分组下的所有文件
        (new UploadFile())->where('group_uuid', '=', $this['uuid'])->update(['group_uuid' => 0]);
        // 删除分组
        return $this->delete();
    }
}
