<?php

namespace app\shop\model\product;

use help\ValidateHelp;
use app\common\model\product\Material;
use app\common\service\websocket\Websocket;
use app\common\model\store\MultiLanguageName;
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
            $model = $model->jsonLike('name', $data['unit_name']);
        }
        $list = $model->order(['sort' => 'asc', 'create_time' => 'asc'])->paginate($data)?->append(['product_ids'], true);

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
        // 获取当前最大的排序值
        $maxSort = $this->where('uuid', '<>', $this['uuid'])->max('sort');
        $data['sort'] = $maxSort + 1;
        $data['name'] = $data['unit_name'] ?? '';
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($data['unit_name']);
        // 保存单位
        $this->save($data);
        return array_merge($data, ['unit_id' => $this->uuid]);
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
        //
        $data['name'] = $data['unit_name'] ?? '';
        (new MultiLanguageName)->saveNames($data['unit_name'], $this['multi_language_name_uuid']);
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
        $this->startTrans();
        try {
            // 删除多语言数据
            $models = $this->whereIn('uuid', $unit_id)->select();
            foreach ($models as $model) {
                if ($model['multi_language_name_uuid']) {
                    (new MultiLanguageName)->where('uuid', $model['multi_language_name_uuid'])->find()?->delete();
                }
                $model->delete();
            }
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
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
            $this->product()->update(['unit_uuid' => 0]);
            $this->material()->update(['unit_uuid' => 0]);
            // 添加新关系
            Product::whereIn('uuid', $product_ids)->update(['unit_uuid' => $unit_id]);
            Material::whereIn('uuid', $product_ids)->update(['unit_uuid' => $unit_id]);

            $this->commit();
            
            // 推送
            $msgData = [
                'type' => 'update',
                'product_uuid' => 0,
                'update_time' => time()
            ];
            Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_PRODUCT, 0, $msgData);

            return true;
        } catch (\Exception $e) {
            $this->rollback();
            $this->error = $e->getMessage();
            return false;
        }
    }
}
