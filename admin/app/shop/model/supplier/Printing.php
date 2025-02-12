<?php

namespace app\shop\model\supplier;

use app\common\model\supplier\Printing as PrintingModel;

/**
 * 菜品打印模型
 */
class Printing extends PrintingModel
{
    /**
     * 获取列表数据
     */
    public function getLists($params, $user)
    {
        $model = $this;
        // 查询列表数据
        return $model->order(['create_time' => 'desc'])->paginate($params);
    }

    /**
     * 添加
     */
    public function add($data, $user)
    {
        $detail = $this->where('name', '=', $data['name'])->find();
        if ($detail) {
            $this->error = '名称已存在';
            return false;
        }
        // 开启事务
        $this->startTrans();
        try {
            $this->save($data);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        $detail = $this->where('name', '=', $data['name'])->where('id', '<>', $this['id'])->find();
        if ($detail) {
            $this->error = '名称已存在';
            return false;
        }
        // 开启事务
        $this->startTrans();
        try {
            if (!isset($data['category_id'])) {
                $data['category_id'] = '';
            }
            if (!isset($data['label_id'])) {
                $data['label_id'] = '';
            }
            $this->save($data);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 软删除
     */
    public function setDelete()
    {
        return $this->delete();
    }
}
