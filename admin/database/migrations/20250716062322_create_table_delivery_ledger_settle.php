<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateTableDeliveryLedgerSettle extends Migrator
{
    // 迁移目标
    const TARGET = 'main';
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
        if (!$this->hasTable('delivery_ledger_settle')) {
            $table = $this->table('delivery_ledger_settle', ['comment' => '外送台账结清数据']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID']);
            $table->addColumn('company_uuid', 'biginteger', ['default' => 0, 'comment' => '公司ID']);
            $table->addColumn('month', 'string', ['limit' => 7, 'null' => false, 'default' => '', 'comment' => '月份']);
            $table->addColumn('order_count', 'integer', ['null' => false, 'default' => 0, 'comment' => '订单数']);
            $table->addColumn('delivery_fee_amount', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '总配送费']);
            $table->addColumn('channel_data', 'text', ['comment' => '渠道数据']);
            $table->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间']);
            $table->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间']);
            $table->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间']);
            $table->create();
        }
    }
}
