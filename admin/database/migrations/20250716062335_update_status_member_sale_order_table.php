<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateStatusMemberSaleOrderTable extends Migrator
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
          // 添加字段
          $table = $this->table('member_sale_order');
          if ($table->hasColumn('status')) {
              $table->changeColumn('status', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '订单状态 0-选购中 1-待支付 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消']);
              $table->update();
          }
    }
}
