<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMemberUuidToMemberSaleOrder extends Migrator
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
        $table =  $this->table('member_sale_order');
        if (!$table->hasColumn('member_uuid')) {
            $table->addColumn('member_uuid', 'biginteger', ['default' => 0, 'after' => 'uuid', 'comment' => '会员UUID']); // 会员UUID
        }
        $table->update(); 
    }
}
