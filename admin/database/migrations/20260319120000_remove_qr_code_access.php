<?php

use think\migration\Migrator;

/**
 * 删除业务设置下的"二维码"权限（已迁移至手机点餐模块）
 */
class RemoveQrCodeAccess extends Migrator
{
    public function up()
    {
        // 删除角色关联的二维码权限
        $this->execute("DELETE FROM `ttpos_role_access` WHERE `access_uuid` = '2859412230144000'");
        // 删除二维码权限记录
        $this->execute("DELETE FROM `ttpos_access` WHERE `uuid` = '2859412230144000'");
    }

    public function down()
    {
        // 不可逆
    }
}
