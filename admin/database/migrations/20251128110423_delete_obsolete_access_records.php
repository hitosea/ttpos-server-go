<?php

use think\facade\Db;
use think\migration\Migrator;

class DeleteObsoleteAccessRecords extends Migrator
{
    /**
     * 删除废弃的权限记录
     * 
     * 删除 ttpos_access 和 ttpos_role_access 表中的指定权限记录
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 需要删除的 access_uuid 列表
        $obsoleteAccessUuids = [
            2856459440128000,
            2856488800256000,
            2856522354688000,
            2856572686336000,
            2856606240768000,
            2856652378112000,
            2856685932544000,
            2856744652800000,
            2858422374400000,
            2858560786432000,
            2859252846592001,
        ];

        try {
            // 开启事务
            $db->startTrans();

            // 1. 删除 ttpos_role_access 表中的关联记录
            $deletedRoleAccess = $db->table('ttpos_role_access')
                ->whereIn('access_uuid', $obsoleteAccessUuids)
                ->delete();

            // 2. 删除 ttpos_access 表中的权限记录
            $deletedAccess = $db->table('ttpos_access')
                ->whereIn('uuid', $obsoleteAccessUuids)
                ->delete();

            // 提交事务
            $db->commit();

        } catch (\Exception $e) {
            // 回滚事务
            $db->rollback();
            echo "删除失败: " . $e->getMessage() . "\n";
            throw $e;
        }
    }
}
