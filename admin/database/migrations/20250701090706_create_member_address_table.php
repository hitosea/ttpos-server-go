<?php

use think\migration\Migrator;

class CreateMemberAddressTable extends Migrator
{
    /**
     * 创建会员地址表
     */
    public function change()
    {
        if (!$this->hasTable('member_address')) {
            $table = $this->table('member_address', ['comment' => '会员地址表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
                ->addColumn('member_uuid', 'biginteger', ['default' => 0, 'comment' => '会员ID'])
                ->addColumn('name', 'string', ['limit' => 30, 'default' => '', 'comment' => '联系人'])
                ->addColumn('phone', 'string', ['limit' => 20, 'default' => '', 'comment' => '手机号'])
                ->addColumn('country', 'string', ['limit' => 10, 'default' => '+66', 'comment' => '国家代码'])
                ->addColumn('province', 'string', ['limit' => 30, 'default' => '', 'comment' => '省'])
                ->addColumn('city', 'string', ['limit' => 30, 'default' => '', 'comment' => '市'])
                ->addColumn('area', 'string', ['limit' => 30, 'default' => '', 'comment' => '区'])
                ->addColumn('address', 'string', ['limit' => 255, 'default' => '', 'comment' => '详细地址'])
                ->addColumn('street', 'string', ['limit' => 255, 'default' => '', 'comment' => '街道/门牌号'])
                ->addColumn('is_default', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否默认 0否 1是'])
                ->addColumn('location', 'string', ['limit' => 100, 'default' => '', 'comment' => '位置坐标'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
} 