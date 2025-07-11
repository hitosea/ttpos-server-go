<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateMemberToSass extends Migrator
{
    // 迁移目标
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        if (!$this->hasTable('member')) {
            $table = $this->table('member', ['comment' => '会员表']);
            $table->addColumn('company_uuid', 'integer', ['limit' => 10, 'null' => false, 'comment' => '公司ID', 'after' => 'uuid']);
            $table->addColumn('device_id', 'string', ['limit' => 255, 'null' => false, 'comment' => '设备ID', 'after' => 'company_uuid']);
            $table->addColumn('is_visitor', 'integer', ['limit' => 1, 'null' => false, 'comment' => '是否游客 0否 1是', 'after' => 'device_id']);
            $table->addColumn('nickname', 'string', ['limit' => 255, 'default' => '', 'comment' => '昵称', 'after' => 'is_visitor']);
            $table->addColumn('phone', 'string', ['limit' => 20, 'default' => '', 'comment' => '手机号', 'after' => 'nickname']);
            $table->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间', 'after' => 'phone']);
            $table->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间', 'after' => 'create_time']);
            $table->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间', 'after' => 'update_time']);
            $table->create();
        }
    }
} 