<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTtposStaffTableInSaas extends Migrator
{
    const TARGET = 'main';
    /**
     * 在 saas 数据库中创建 ttpos_staff 表（统一账号表）
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('staff')) {
            $table = $this->table('staff', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '员工表（统一账号表）'
            ]);

            $table
                // 基础字段
                ->addColumn('id', 'integer', ['identity' => true, 'signed' => false, 'comment' => '自增ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '员工ID'])
                
                // 账号信息字段
                ->addColumn('email', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '邮箱（全平台唯一）'])
                ->addColumn('phone', 'string', ['limit' => 20, 'null' => true, 'default' => '', 'comment' => '手机号（全平台唯一，允许空字符串）'])
                ->addColumn('real_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '姓名'])
                
                // 密码相关字段
                ->addColumn('password', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '登录密码（加密）'])
                ->addColumn('password_change_count', 'integer', ['default' => 0, 'null' => true, 'comment' => '修改密码次数'])
                ->addColumn('password_change_time', 'integer', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '修改密码时间'])
                
                // 状态字段
                ->addColumn('is_disable', 'integer', ['limit' => 1, 'default' => 0, 'null' => false, 'comment' => '是否禁用1禁用,0未禁用'])
                
                // 门店相关字段
                ->addColumn('last_company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '上次登录新管理端的商家UUID'])
                
                // 时间字段
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '创建时间(时间戳)'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '更新时间(时间戳)'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'null' => false, 'comment' => '删除时间(时间戳)'])
                
                // 索引
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
                ->addIndex(['email'], ['unique' => true, 'name' => 'uk_email'])
                ->addIndex(['phone'], ['name' => 'idx_phone'])
                ->addIndex(['last_company_uuid'], ['name' => 'idx_last_company_uuid'])
                ->addIndex(['delete_time'], ['name' => 'idx_delete_time'])
                
                ->create();
        }
    }
}
