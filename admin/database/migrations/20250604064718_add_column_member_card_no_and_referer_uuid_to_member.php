<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnMemberCardNoAndRefererUuidToMember extends Migrator
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
        $table = $this->table('member');
        if (!$table->hasColumn('referrer_uuid')) {
            $table->addColumn('referrer_uuid', 'biginteger', ['default' => 0, 'null' => false, 'comment' => '推荐人Uuid', 'after' => 'member_card_uuid']);
        }
        if (!$table->hasColumn('member_card_no')) {
            $table->addColumn('member_card_no', 'string', ['default' => '', 'null' => false, 'comment' => '会员卡号', 'after' => 'member_card_uuid']);
        }
        $table->update();
    }
}
