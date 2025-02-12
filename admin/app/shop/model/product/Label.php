<?php

namespace app\shop\model\product;

use app\common\model\product\Label as LabelModel;

/**
 * 标签模型
 */
class Label extends LabelModel
{
    /**
     * 获取列表数据
     */
    public function getList($data, $shop_supplier_id)
    {
        $model = $this;
        if (isset($data['label_name']) && $data['label_name'] != '') {
            $model = $model->like('label_name', trim($data['label_name']));
        }
        $list =  $model->order(['create_time' => 'desc'])->paginate($data)?->append(['product_ids'], true);

        // 是否关联产品
        foreach ($list as &$item) {
            $item['is_used'] = $this->isUseWithProduct($item['label_id']) ? 1 : 0;
        }
        return $list;
    }

    /**
     * 添加
     */
    public function add($data, $shop_supplier_id)
    {
        $isExist = $this->where('label_name', '=', $data['label_name'])->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $data['shop_supplier_id'] = $shop_supplier_id;
        $data['app_id'] = self::$app_id;
        return $this->save($data);
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        $isExist = $this->where('label_name', '=', $data['label_name'])
            ->where('label_id', '<>', $this['label_id'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        return $this->save($data);
    }

    /**
     * 删除
     */
    public function setDelete($label_id)
    {
        // 判断是否关联产品
        if ($this->isUseWithProduct($label_id)) {
            $this->error = '该标签下存在商品，不允许删除';
            return false;
        }
        return $this->where('label_id', 'in', $label_id)->delete();
    }

    /**
     * 更新产品标签
     *
     * @param int $label_id 标签ID
     * @param array $product_ids 产品ID数组
     * @return bool
     */
    public function relatedProduct($label_id, $product_ids)
    {
        // 开始事务
        $this->startTrans();
        try {
            // 删除原有关系
            $this->product()->update(['label_id' => 0]);
            // 添加新关系
            Product::whereIn('product_id', $product_ids)->update(['label_id' => $label_id]);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
