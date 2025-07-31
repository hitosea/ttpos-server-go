<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddFieldsToStatisticsProduct extends Migrator
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
        $table = $this->table('statistics_product');
        if (!$table->hasColumn('is_takeout')) {
            $table->addColumn('is_takeout', 'integer', ['default' => 0, 'comment' => '是否外送', 'after' => 'refund_num']);
            $table->addIndex(['is_takeout'], ['name' => 'idx_is_takeout']);
            $table->update();
        }
        if (!$table->hasColumn('member_order_discount_rate')) {
            $table->addColumn('member_order_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'null' => false, 'default' => 1.0000, 'comment' => '会员端商品价格上浮比例1%-300%', 'after' => 'is_takeout'])
                ->update();
        }
        $table->update();
    }
}
