<?php

use think\migration\Migrator;

class AddOrdersIdx extends Migrator
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
            $table = $this->table('sale_order');
            if (!$table->hasIndexByName('idx_tso_bill_qry')) {
                $table->addIndex(['delete_time','sale_bill_uuid'], ['name' => 'idx_tso_bill_qry']);
                $table->update();
            }
            $table = $this->table('sale_order_product');
            if (!$table->hasIndexByName('idx_tsop_order_qry')) {
                $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tsop_order_qry']);
                $table->update();
            }
            $table = $this->table('sale_order_coupon');
            if (!$table->hasIndexByName('idx_tsoc_order_qry')) {
                $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tsoc_order_qry']);
                $table->update();
            }
            $table = $this->table('payment_order');
            if (!$table->hasIndexByName('idx_tpo_order_qry')) {
                $table->addIndex(['delete_time','related_uuid'], ['name' => 'idx_tpo_order_qry']);
                $table->update();
            }
            $table = $this->table('sale_order_buffet_customer_type');
            if (!$table->hasIndexByName('idx_tsobcf_order_qry')) {
                $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tsobcf_order_qry']);
                $table->update();
            }
        } catch (\Exception $e) {
            trace($e->getMessage(), 'error');
            throw $e;
        }
    }
}
