<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateProductLabelTable extends Migrator
{
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
        // 检查表是否已存在
        if ($this->hasTable('product_label')) {
            return;
        }

        // 创建 product_label 表
        $table = $this->table('product_label', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_general_ci',
            'comment' => '商品标签表'
        ]);

        $table->addColumn('id', 'integer', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
            ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识UUID'])
            ->addColumn('name', 'text', ['null' => true, 'comment' => '标签名称'])
            ->addColumn('style', 'string', ['limit' => 50, 'default' => '', 'comment' => '标签样式'])
            ->addColumn('is_show_cashier', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否在收银机显示, 0-否 1-是'])
            ->addColumn('is_show_tablet', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否在平板显示, 0-否 1-是'])
            ->addColumn('is_show_assistant', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否在助手显示, 0-否 1-是'])
            ->addColumn('is_show_h5', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否在H5显示, 0-否 1-是'])
            ->addColumn('is_show_delivery', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否在外送显示, 0-否 1-是'])
            ->addColumn('is_show_menu', 'integer', ['signed' => false, 'default' => 0, 'comment' => '是否在电子菜单显示, 0-否 1-是'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
            ->create();
    }
}

