<?php

use think\migration\Migrator;

class AddColumnSaleOrderProductOpenMember extends Migrator
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
        // 销售订单表
        $table = $this->table('sale_order_product');
        if (!$table->hasColumn('open_member_discount')) {
            $table->addColumn('open_member_discount', 'integer', ['default' => 0, 'comment' => '是否开启会员折扣, 0-否 1-是。添加商品时记录下状态不受后台改变，结账时检查是否改变'])
                  ->update();
        }
        $table->update();
        $table = $this->table('product_package');
        if ($table->hasColumn('open_discount')) {
            $table->renameColumn('open_discount', 'open_member_discount')
                  ->update();
        }
        $table->update();
    }
}
