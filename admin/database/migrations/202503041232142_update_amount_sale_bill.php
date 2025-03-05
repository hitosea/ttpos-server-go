<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateAmountSaleBill extends Migrator
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
        if ($table->hasColumn('product_original_fee')) {
            $table->changeColumn('product_original_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0])->update();
        }
        if (!$table->hasColumn('production_time')) {
            $table->addColumn('production_time', 'integer', ['default' => 0])->update();
        }
        if ($table->hasColumn('status')) {
            $table->changeColumn('status', 'tinyinteger', ['default' => 0, 'comment' => '订单状态, 0-待付款、1-已完成、2-已取消。'])->update();
        }
    }
}
