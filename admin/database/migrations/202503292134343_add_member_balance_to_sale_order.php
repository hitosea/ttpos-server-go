<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMemberBalanceToSaleOrder extends Migrator
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
        $table = $this->table('sale_order');
        if (!$table->hasColumn('member_balance')) {
            $table->addColumn('member_balance', 'decimal', [
                'precision' => 12,
                'scale' => 2,
                'default' => 0,
                'comment' => '会员余额. 会员消费本单后剩余的余额',
            ]);
        }
        $table->update();
    }
}
