<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ChangeDiscountDefaultValueOne extends Migrator
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
        if ($table->hasColumn('member_discount_rate')) {
            $table->changeColumn('member_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
        if ($table->hasColumn('member_card_discount_rate')) {
            $table->changeColumn('member_card_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
        if ($table->hasColumn('custom_discount_rate')) {
            $table->changeColumn('custom_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }

        $table = $this->table('sale_order_product');
        if ($table->hasColumn('custom_discount_rate')) {
            $table->changeColumn('custom_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
        if ($table->hasColumn('member_discount_rate')) {
            $table->changeColumn('member_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
        if ($table->hasColumn('member_card_discount_rate')) {
            $table->changeColumn('member_card_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }

        $table = $this->table('member_level');
        if ($table->hasColumn('discount')) {
            $table->changeColumn('discount', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
        $table = $this->table('member_card');
        if ($table->hasColumn('discount')) {
            $table->changeColumn('discount', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
        $table = $this->table('member_card_log');
        if ($table->hasColumn('discount')) {
            $table->changeColumn('discount', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1.00])->update();
        }
    }
}
