<?php

use think\migration\Migrator;

class AddIsDistanceCalculatedFromMemberSaleOrderTable extends Migrator
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
        $table = $this->table('member_sale_order');
        if (!$table->hasColumn('is_distance_calculated')) {
            $table->addColumn('is_distance_calculated', 'integer', ['null' => false, 'default' => -1, 'comment' => '是否已计算距离费，-1-未计算，1-已计算', 'after' => 'refund_amount']);
            $table->update();
        }
    }
} 