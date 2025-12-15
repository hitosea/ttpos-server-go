<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyCompanySettingAddressCoordinatesLength extends Migrator
{
    /**
     * 修改 ttpos_company_setting 表的 address 和 coordinates 字段长度为 500
     * 适用于 saas 库和商家数据库
     */
    public function change()
    {
        $table = $this->table('company_setting');

        // 修改 address 字段长度为 500
        if ($table->hasColumn('address')) {
            $table->changeColumn('address', 'string', ['limit' => 500, 'null' => false, 'default' => '', 'comment' => '联系地址'])->update();
        }

        // 修改 coordinates 字段长度为 500
        if ($table->hasColumn('coordinates')) {
            $table->changeColumn('coordinates', 'string', ['limit' => 500, 'null' => false, 'default' => '', 'comment' => '经纬度，如：13.721899,100.52900'])->update();
        }
    }
}
