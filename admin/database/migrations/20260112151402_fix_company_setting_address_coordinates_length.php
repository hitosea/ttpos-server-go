<?php

use think\migration\Migrator;
use think\migration\db\Column;

class FixCompanySettingAddressCoordinatesLength extends Migrator
{
    const TARGET = 'all';
    /**
     * 修改 ttpos_company_setting 表的 address 和 coordinates 字段长度为 500
     * 
     * 问题背景：
     * - saas 库的 address 字段长度为 varchar(255)，导致保存较长地址时报错
     * - 商家库已经是 varchar(500)，但为了保持一致性，此迁移会检查并更新
     * 
     * 修复：将 address 和 coordinates 字段长度统一改为 varchar(500)
     */
    public function change()
    {
        $table = $this->table('company_setting');

        // 修改 address 字段长度为 500
        if ($table->hasColumn('address')) {
            $table->changeColumn('address', 'string', [
                'limit' => 500,
                'null' => false,
                'default' => '',
                'comment' => '联系地址'
            ])->update();
        }

        // 修改 coordinates 字段长度为 500
        if ($table->hasColumn('coordinates')) {
            $table->changeColumn('coordinates', 'string', [
                'limit' => 500,
                'null' => false,
                'default' => '',
                'comment' => '经纬度，如：13.721899,100.52900'
            ])->update();
        }
    }
}
