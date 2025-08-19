<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFieldsToProductPackageTable extends Migrator
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
        $table = $this->table('product_package');
        
        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('price')) {
            $table->addColumn('price', 'decimal', [
                'precision' => 12,
                'scale' => 2,
                'null' => false,
                'default' => 0,
                'comment' => '套餐价格',
                'after' => 'describe'
            ])
            ->update();
        }
        
        if (!$table->hasColumn('product_type')) {
            $table->addColumn('product_type', 'integer', [
                'null' => false,
                'default' => 0,
                'comment' => '商品类型, 0-商品 1-套餐',
                'after' => 'price'
            ])
            ->update();
        }
    }
} 