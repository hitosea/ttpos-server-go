<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRelatedUuidToMemberBalanceLog extends Migrator
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
        $table = $this->table('member_balance_log');
        if (!$table->hasColumn('related_uuid')) {
            $table->addColumn('related_uuid', 'biginteger', [
                'default' => 0,
                'comment' => '关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景90)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40)',
                'null' => false,
                'after' => 'processed',
            ]);
        }
        $table->update();
    }
}
