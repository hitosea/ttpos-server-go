<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIndexToStatistics extends Migrator
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
        try {
            $table = $this->table('statistics_sale');
            if ($table->hasColumn('sale_bill_uuid') && !$table->hasIndexByName('idx_sale_bill_uuid')) {
                $table->addIndex(['sale_bill_uuid'], ['name' => 'idx_sale_bill_uuid']);
            }
            $table->update();
    
            $table = $this->table('statistics_payment');
            if ($table->hasColumn('sale_bill_uuid') && !$table->hasIndexByName('idx_sale_bill_uuid')) {
                $table->addIndex(['sale_bill_uuid'], ['name' => 'idx_sale_bill_uuid']);
            }
            $table->update();
    
            $table = $this->table('statistics_member_payment');
            if ($table->hasColumn('member_recharge_order_uuid') && !$table->hasIndexByName('idx_member_recharge_order_uuid')) {
                $table->addIndex(['member_recharge_order_uuid'], ['name' => 'idx_member_recharge_order_uuid']);
            }
            $table->update();
    
            $table = $this->table('statistics_member');
            if ($table->hasColumn('member_recharge_order_uuid') && !$table->hasIndexByName('idx_member_recharge_order_uuid')) {
                $table->addIndex(['member_recharge_order_uuid'], ['name' => 'idx_member_recharge_order_uuid']);
            }
            $table->update();
        } catch (\Exception $e) {
            trace($e->getMessage(), 'error');
            throw $e;
        }
    }
}
