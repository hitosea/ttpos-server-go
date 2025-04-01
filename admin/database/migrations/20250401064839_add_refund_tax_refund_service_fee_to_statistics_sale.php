<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRefundTaxRefundServiceFeeToStatisticsSale extends Migrator
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
        $table = $this->table('statistics_sale');
        if (!$table->hasColumn('refund_tax') && !$table->hasColumn('refund_service_fee')) {
            $table->addColumn('refund_tax', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款税费', 'after' => 'refund_amount']);
            $table->addColumn('refund_service_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '退款服务费', 'after' => 'refund_tax']);
            $table->update();
        }
    }
}
