<?php

namespace app\shop\model\product;

use app\common\model\product\RelatedMaterial as RelatedMaterialModel;
use think\facade\Db;
use think\facade\Env;

class RelatedMaterial extends RelatedMaterialModel
{
    /**
     * 更新规格/加料关联材料
     */
    public static function updateRelatedMaterial($materialList, $relatedUuid)
    {
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
        // 更新规格/加料关联材料库存
        self::updateStock($relatedMaterialUuidList);
    }

    /**
     * 删除规格/加料关联材料
     */
    public static function deleteRelatedMaterial($relatedUuid)
    {
        $relatedMaterialList = RelatedMaterial::where('related_uuid', $relatedUuid)->select();
        foreach ($relatedMaterialList as $relatedMaterial) {
            $relatedMaterial->delete();
        }
    }

    /**
     * 更新规格/加料关联材料库存
     */
    public static function updateStock($relatedMaterialUuidList)
    {
        return;
        // if (empty($relatedMaterialUuidList)) {
        //     return;
        // }
        // // 提取材料uuid列表
        // $relatedMaterialUuidList = implode(',', $relatedMaterialUuidList);

        // $db = Db::connect((new self())->getConnection());
        // $prefix = Env::get('DB_PREFIX');

        // // 更新规格/加料关联材料库存
        // $db->startTrans();
        // try {
        //     $db->execute("
        //         UPDATE {$prefix}product_bom AS rm 
        //         JOIN (
        //             SELECT a.related_uuid AS related_uuid, MIN(a.min_stock_num) AS min_stock_num
        //             FROM (
        //                 SELECT rms.related_uuid, LEAST(FLOOR(MIN(m.stock / rms.num)), 99999999) AS min_stock_num
        //                 FROM {$prefix}related_material AS rms
        //                 JOIN {$prefix}warehouse_item AS m ON rms.material_uuid = m.material_uuid
        //                 JOIN {$prefix}warehouse AS w ON m.warehouse_uuid = w.uuid
        //                 WHERE rms.uuid IN ({$relatedMaterialUuidList})
        //                 AND (w.is_default = 1 AND w.headquarter_uuid = 0)
        //                 GROUP BY rms.uuid
        //             ) AS a
        //         ) AS sub ON rm.uuid = sub.related_uuid
        //         SET rm.stock_num = sub.min_stock_num
        //         WHERE rm.uuid IN (
        //             SELECT related_uuid FROM {$prefix}related_material WHERE uuid IN ({$relatedMaterialUuidList})
        //         );
        //     ");
        //     $db->commit();
        // } catch (\Exception $e) {
        //     // 出现异常时，回滚事务
        //     $db->rollback();
        //     // 记录或处理异常
        //     trace('Error: ' . $e->getMessage());
        // }
        
    }
}