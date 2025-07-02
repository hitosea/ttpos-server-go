<?php

use think\migration\Migrator;

class UpdateDataDeliveryToPrinterTemplate extends Migrator
{
    // 迁移目标
    const TARGET = 'all';
    
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
        $table = $this->table('ttpos_sale_order');
        if (!$table->hasIndex('idx_tso_bill_qry')) {
            $table->addIndex(['delete_time','sale_bill_uuid'], ['name' => 'idx_tso_bill_qry']);
        }
        $table = $this->table('ttpos_sale_order_product');
        if (!$table->hasIndex('idx_tsop_order_qry')) {
            $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tsop_order_qry']);
        }
        $table = $this->table('ttpos_sale_order_coupon');
        if (!$table->hasIndex('idx_tsoc_order_qry')) {
            $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tsoc_order_qry']);
        }
        $table = $this->table('ttpos_payment_order');
        if (!$table->hasIndex('idx_tpo_order_qry')) {
            $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tpo_order_qry']);
        }
        $table = $this->table('ttpos_sale_order_buffet_customer_type');
        if (!$table->hasIndex('idx_tsobcf_order_qry')) {
            $table->addIndex(['delete_time','sale_order_uuid'], ['name' => 'idx_tsobcf_order_qry']);
        }

    }
}
