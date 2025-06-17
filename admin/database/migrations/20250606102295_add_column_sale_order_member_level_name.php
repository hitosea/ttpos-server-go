<?php

use think\migration\Migrator;

class AddColumnSaleOrderMemberLevelName extends Migrator
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
        if (!$table->hasColumn('gift_points_type')) {
            $table->addColumn('gift_points_type', 'integer', ['default' => 0, 'comment' => '赠送积分类型, 0-按比例赠送 1-按人数固定金额赠送', 'after' => 'gift_points_rate']);
        }
        if (!$table->hasColumn('member_level_name')) {
            $table->addColumn('member_level_name', 'string', ['default' => '', 'comment' => '会员等级名称', 'after' => 'member_balance', 'limit' => 255]);
        }
        $table->update();
    }
}

