<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddIsKitchenConfirmToSaleBillAndInitNumToProductionOrderProduct extends Migrator
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
        $table = $this->table('sale_bill');
        if (!$table->hasColumn('is_kitchen_confirm')) {
            $table->addColumn('is_kitchen_confirm', 'integer', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '厨显是否确认退菜，确认后不在厨显端显示已经整单取消的菜品,0:未确认,1:已确认', 'after' => 'finish_time'])
                ->update();
        }

        $db = Db::connect(Db::getConfig('default'), true);
        $db->name("sale_bill")->where("is_kitchen_confirm", "=", "0")->update(["is_kitchen_confirm" => 1]);

        $table = $this->table('production_order_product');
        if (!$table->hasColumn('init_num')) {
            $table->addColumn('init_num', 'decimal', ['precision' => 12, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '初始送厨数量，退菜后，init_num肯定大于num', 'after' => 'num'])
                ->update();
        }
        $db = Db::connect(Db::getConfig('default'), true);
        $db->name("production_order_product")->where('id', '>', '0')->update(["init_num" => Db::raw("num")]);
    }
}
