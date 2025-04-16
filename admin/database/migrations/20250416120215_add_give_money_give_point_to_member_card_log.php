<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddGiveMoneyGivePointToMemberCardLog extends Migrator
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
        $table = $this->table('member_card_log');
        if (!$table->hasColumn('give_money')) {
            $table->addColumn('give_money', 'decimal', ['precision' => 10, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '赠送余额', 'after' => 'member_uuid'])
                ->update();
        }
        if (!$table->hasColumn('give_point')) {
            $table->addColumn('give_point', 'decimal', ['precision' => 10, 'scale' => 2, 'null' => false, 'default' => 0, 'comment' => '赠送积分', 'after' => 'give_money'])
                ->update();
        }
    }
}
