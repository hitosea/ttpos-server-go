<?php

namespace app\admin\model\settings;

use app\common\model\settings\Message as MessageModel;

class Message extends MessageModel
{
    /**
     * 获取全部
     */
    public static function getAll()
    {
        $model = new static;
        return $model->order(['sort' => 'asc'])->select();
    }

    /**
     * 新增
     */
    public function add($data)
    {
        return $this->save($data);
    }


    /**
     * 软删除
     */
    public function setDelete()
    {
        return $this->delete();
    }

    /**
     * 更新记录
     */
    public function edit($data)
    {
        return $this->save($data);
    }
}
