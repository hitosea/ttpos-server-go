<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateMaterialCategoryTable extends Migrator
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
        if ($this->hasTable('material_category')) {
            return;
        }

        $table = $this->table('material_category', [
            'id' => false,
            'primary_key' => ['id'],
            'engine' => 'InnoDB',
            'collation' => 'utf8mb4_unicode_ci',
            'comment' => '原料分类表'
        ]);

        $table->addColumn('id', 'integer', [
            'identity' => true,
            'signed' => false,
            'limit' => 11,
            'comment' => '自增ID'
        ])
        ->addColumn('uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '原料分类ID'
        ])
        ->addColumn('name', 'string', [
            'limit' => 255,
            'default' => '',
            'comment' => '原料分类名称'
        ])
        ->addColumn('multi_language_name_uuid', 'biginteger', [
            'signed' => false,
            'default' => 0,
            'comment' => '多语言名称ID'
        ])
        ->addColumn('create_time', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '创建时间(时间戳)'
        ])
        ->addColumn('update_time', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '更新时间(时间戳)'
        ])
        ->addColumn('delete_time', 'integer', [
            'signed' => false,
            'limit' => 10,
            'default' => 0,
            'comment' => '删除时间(时间戳)'
        ])
        ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
        ->create();
    }
} 