<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddProductPackageGroupTables extends Migrator
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
        // 创建商品套餐组表
        $this->createProductPackageGroupTable();
        
        // 创建商品套餐组商品表
        $this->createProductPackageGroupItemTable();
    }
    
    /**
     * 创建商品套餐组表
     */
    private function createProductPackageGroupTable()
    {
        $table = $this->table('product_package_group');
        
        // 检查表是否已存在
        if (!$this->hasTable('product_package_group')) {
            $table->addColumn('uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '商品套餐组ID'
            ])
            ->addColumn('name', 'text', [
                'null' => true,
                'comment' => '名称'
            ])
            ->addColumn('multi_language_name_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '多语言名称ID'
            ])
            ->addColumn('product_package_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '商品套餐UUID'
            ])
            ->addColumn('create_time', 'integer', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '创建时间(时间戳)'
            ])
            ->addColumn('update_time', 'integer', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '更新时间(时间戳)'
            ])
            ->addColumn('delete_time', 'integer', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '删除时间(时间戳)'
            ])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->addIndex(['product_package_uuid'], ['name' => 'idx_product_package_uuid'])
            ->addIndex(['multi_language_name_uuid'], ['name' => 'idx_multi_language_name_uuid'])
            ->create();
        }
    }
    
    /**
     * 创建商品套餐组商品表
     */
    private function createProductPackageGroupItemTable()
    {
        $table = $this->table('product_package_group_item');
        
        // 检查表是否已存在
        if (!$this->hasTable('product_package_group_item')) {
            $table->addColumn('uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '商品套餐组商品ID'
            ])
            ->addColumn('product_package_group_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '商品套餐组ID'
            ])
            ->addColumn('related_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '关联商品UUID, product_package_uuid'
            ])
            ->addColumn('product_bom_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '商品BOM UUID,商品规格uuid'
            ])
            ->addColumn('num', 'decimal', [
                'precision' => 12,
                'scale' => 4,
                'null' => false,
                'default' => 0,
                'comment' => '数量'
            ])
            ->addColumn('sort', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '排序'
            ])
            ->addColumn('create_time', 'integer', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '创建时间(时间戳)'
            ])
            ->addColumn('update_time', 'integer', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '更新时间(时间戳)'
            ])
            ->addColumn('delete_time', 'integer', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '删除时间(时间戳)'
            ])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->addIndex(['product_package_group_uuid'], ['name' => 'idx_product_package_group_uuid'])
            ->addIndex(['related_uuid'], ['name' => 'idx_related_uuid'])
            ->addIndex(['product_bom_uuid'], ['name' => 'idx_product_bom_uuid'])
            ->addIndex(['sort'], ['name' => 'idx_sort'])
            ->create();
        }
    }
} 