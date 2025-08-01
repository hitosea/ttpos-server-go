<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRiderAcceptTimeoutToMemberSaleOrder extends Migrator
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
        if (!$table->hasColumn('rider_accept_timeout')) {
            $table->addColumn('rider_accept_timeout', 'integer', ['null' => false, 'default' => 0, 'comment' => '骑手接单超时时间（秒）', 'after' => 'delivery_fee_per_km']);
            $table->update();
        }
    }
} 