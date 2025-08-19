<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddPackageGroupUuidToSaleOrderProductTable extends Migrator
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
        $table = $this->table('sale_order_product');
        
        // 检查字段是否已存在
        $hasColumn = $table->hasColumn('package_group_uuid');
        if (!$hasColumn) {
            // 添加套餐分组UUID字段
            $table->addColumn('package_group_uuid', 'biginteger', [
                'null' => false,
                'default' => 0,
                'comment' => '套餐分组UUID',
                'after' => 'package_uuid'
            ])->update();
            
            // 添加索引
            $table->addIndex(['package_group_uuid'], ['name' => 'idx_package_group_uuid'])->update();
        }
    }
} 