<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddCancelSceneAndIsAutoAcceptToMemberSaleOrder extends Migrator
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
        if (!$table->hasColumn('cancel_scene')) {
            $table->addColumn('cancel_scene', 'string', ['limit' => 50, 'null' => false, 'default' => '', 'comment' => '取消场景：merchant_cancel-商家取消；member_cancel-用户取消；merchant_reject-商家拒单', 'after' => 'status']);
            $table->addColumn('is_auto_accept', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否自动接单：0-否；1-是', 'after' => 'cancel_scene']);
            $table->update();
        }
    }
}
