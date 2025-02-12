<?php

namespace app\shop\model\product;

use help\ValidateHelp;
use app\common\model\product\Unit as UnitModel;

/**
 * 单位模型
 */
class Unit extends UnitModel
{
    /**
     * 获取列表数据
     */
    public function getList($data, $shop_supplier_id)
    {
        $model = $this;
        if (isset($data['unit_name']) && $data['unit_name'] != '') {
            $model = $model->jsonLike('unit_name', $data['unit_name']);
        }
        $list = $model->order(['create_time' => 'desc'])->paginate($data)?->append(['product_ids'], true);

        // 是否关联产品
        foreach ($list as &$item) {
            $item['is_used'] = $this->isUseWithProduct($item['unit_id']) ? 1 : 0;
        }
        return $list;
    }

    /**
     * 添加
     */
    public function add($data, $shop_supplier_id)
    {
        if (ValidateHelp::hasEmptyValue($data['unit_name'] ?? '')) {
            $this->error = '单位名称不能为空';
            return false;
        }
        $isExist = $this->where('unit_name', '=', $data['unit_name'])->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        $data['shop_supplier_id'] = $shop_supplier_id;
        $data['app_id']           = self::$app_id;
        $this->save($data);
        return array_merge($data, ['unit_id' => $this->unit_id]);
    }

    /**
     * 修改
     */
    public function edit($data)
    {
        if (ValidateHelp::hasEmptyValue($data['unit_name'] ?? '')) {
            $this->error = '单位名称不能为空';
            return false;
        }
        $isExist = $this->where('unit_name', '=', $data['unit_name'])
            ->where('unit_id', '<>', $this['unit_id'])
            ->count();
        if ($isExist) {
            $this->error = '名称已存在';
            return false;
        }
        // 更新关联产品的单位
        $this->product()->update(['product_unit' => $data['unit_name']]);
        return $this->save($data);
    }

    /**
     * 删除
     */
    public function setDelete($unit_id)
    {
        // 判断是否关联产品
        if ($this->isUseWithProduct($unit_id)) {
            $this->error = '该单位下存在商品，不允许删除';
            return false;
        }
        return $this->where('unit_id', 'in', $unit_id)->delete();
    }

    /**
     * 关联菜品
     *
     * @param int $unit_id ID
     * @param array $product_ids 产品ID数组
     * @return bool
     */
    public function relatedProduct($unit_id, $product_ids)
    {
        // 开始事务
        $this->startTrans();
        try {
            // 删除原有关系
            $this->product()->update(['unit_id' => 0, 'product_unit' => '']);
            // 添加新关系
            Product::whereIn('product_id', $product_ids)->update(['unit_id' => $unit_id, 'product_unit' => $this['unit_name']]);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
