<?php

use Phinx\Db\Adapter\MysqlAdapter;
use think\migration\Migrator;
use think\migration\db\Column;

class CreateWarehourseMonthlyProductBomForm extends Migrator
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
        if (!$this->hasTable('warehouse_monthly_product_bom_form')) {
            $table = $this->table('warehouse_monthly_product_bom_form', ['engine' => 'InnoDB', 'collation' => 'utf8mb4_unicode_ci', 'comment' => '月度商品bom报表']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '月度报表uuid']);
            $table->addColumn('year', 'integer', ['default' => 0, 'comment' => '年']);
            $table->addColumn('month', 'integer', ['default' => 0, 'comment' => '月']);
            $table->addColumn('scene', 'integer', ['default' => 0, 'comment' => '记录类型,0-月初 1-月末']);
            $table->addColumn('product_bom_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '商品bom uuid']);
            $table->addColumn('stock', 'decimal', ['precision' => 20, 'scale' => 4, 'default' => 0.0000, 'comment' => '库存']);
            $table->addColumn('create_time', 'integer', ['limit' => MysqlAdapter::INT_REGULAR, 'null' => false, 'default' => 0, 'signed' => true, 'comment' => '创建时间']);
            $table->addColumn('update_time', 'integer', ['limit' => MysqlAdapter::INT_REGULAR, 'null' => false, 'default' => 0, 'signed' => true, 'comment' => '更新时间']);
            $table->addColumn('delete_time', 'integer', ['limit' => MysqlAdapter::INT_REGULAR, 'null' => false, 'default' => 0, 'signed' => true, 'comment' => '更新时间']);
            $table->create();
        }
    }
}
