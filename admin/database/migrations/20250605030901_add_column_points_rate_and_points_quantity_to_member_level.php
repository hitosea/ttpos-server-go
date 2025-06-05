<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnPointsRateAndPointsQuantityToMemberLevel extends Migrator
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
        // 会员等级赠送积分规则设置的值
        $table = $this->table('member_level');
        if (!$table->hasColumn('points_quantity')) {
            $table->addColumn('points_quantity', 'string', ['default' => '', 'null' => false, 'comment' => '购物赠送积分按照桌台人数赠送时的数量', 'after' => 'remark']);
        }
        if (!$table->hasColumn('points_rate')) {
            $table->addColumn('points_rate', 'string', ['default' => '', 'null' => false, 'comment' => '购物赠送积分按照付款金额比例赠送时的比例', 'after' => 'remark']);
        }
        $table->update();
    }
}
