<?php

use think\facade\Db;
use think\migration\Migrator;

/**
 * 去重手机点餐相关权限（mobile_order_setting、mobile_order_qrcode、mobile_order_menu_style）
 * 保留最新记录，删除最早的重复记录及其 role_access 关联
 */
class DedupMobileOrderAccess extends Migrator
{
    public function up()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        $paths = ['mobile_order_setting', 'mobile_order_qrcode', 'mobile_order_menu_style'];

        foreach ($paths as $path) {
            $records = $db->name('access')
                ->where('path', '=', $path)
                ->where('delete_time', '=', 0)
                ->order('id', 'desc')
                ->select();

            if (count($records) <= 1) {
                continue;
            }

            // 跳过第一条（最新的），删除其余重复记录
            $deleteUuids = [];
            foreach ($records as $index => $record) {
                if ($index === 0) {
                    continue;
                }
                $deleteUuids[] = $record['uuid'];
            }

            if (!empty($deleteUuids)) {
                // 删除关联的 role_access
                $db->name('role_access')->whereIn('access_uuid', $deleteUuids)->delete();
                // 删除重复的 access
                $db->name('access')->whereIn('uuid', $deleteUuids)->delete();
            }
        }
    }

    public function down()
    {
        // 不可逆
    }
}
