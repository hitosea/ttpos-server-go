<?php

namespace app\shop\model\product;

use app\common\model\product\RelatedMaterial as RelatedMaterialModel;

class RelatedMaterial extends RelatedMaterialModel
{

    /**
     * 添加规格关联材料
     */
    public static function addRelatedMaterial($materialList, $relatedUuid)
    {
        $saveData = []; 
        foreach ($materialList as $material) {
            $saveData[] = [
                'related_uuid' => $relatedUuid,
                'material_uuid' => $material['product_id'],
                'num' => $material['material_num'] ?? 0,
            ];
        }
        if (!empty($saveData)) {
            (new self())->saveAll($saveData);
        }
    }

    /**
     * 更新规格关联材料
     */
    public static function updateRelatedMaterial($materialList, $relatedUuid)
    {
        // 规格关联材料uuid列表
        if (empty($materialList)) {
            self::deleteRelatedMaterial($relatedUuid);
            return;
        }
        // 规格关联材料uuid列表
        $relatedMaterialUuidList = [];

        foreach ($materialList as $material) {
            $num = $material['material_num'] ?? 0;
            $relatedMaterial = RelatedMaterial::where('related_uuid', $relatedUuid)
                ->where('material_uuid', $material['product_id'])
                ->find();
            // 如果规格关联材料不存在，则创建
            if (!$relatedMaterial) {
                $relatedMaterial = RelatedMaterial::create([
                    'related_uuid' => $relatedUuid,
                    'material_uuid' => $material['product_id'],
                    'num' => $num,
                ]);
            } else {
                // 如果规格关联材料存在，则更新
                $relatedMaterial->save([ 'num' => $num ]);
            }
            // 规格关联材料uuid列表
            $relatedMaterialUuidList[] = $relatedMaterial['uuid'];
        }
        // 如果规格关联材料uuid列表不为空，则删除规格关联材料
        if (!empty($relatedMaterialUuidList)) {
            $relatedMaterialList = RelatedMaterial::whereNotIn('uuid', $relatedMaterialUuidList)
                ->where('related_uuid', $relatedUuid)
                ->select();
            foreach ($relatedMaterialList as $relatedMaterial) {
                $relatedMaterial->delete();
            }
        }
    }

    /**
     * 删除规格关联材料
     */
    public static function deleteRelatedMaterial($relatedUuid)
    {
        $relatedMaterialList = RelatedMaterial::where('related_uuid', $relatedUuid)->select();
        foreach ($relatedMaterialList as $relatedMaterial) {
            $relatedMaterial->delete();
        }
    }
}