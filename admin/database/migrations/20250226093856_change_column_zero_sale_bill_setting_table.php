<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ChangeColumnZeroSaleBillSettingTable extends Migrator
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
        $table = $this->table('sale_bill_setting');
        if ($table->hasColumn('zero')) {
            $table->renameColumn('zero', 'zero_rule')
                ->update();
        }
        if ($table->hasColumn('zero_rule')) {
            $table->changeColumn('zero_rule', 'tinyinteger', ['default' => 0, 'comment' => '优惠折扣抹零, 0-实款实收 1-抹分 2-抹角 3-四舍五入保留一位小数 4-四舍五入保留整数'])
                ->update();
        }
        if ($table->hasColumn('zero_checkout')) {
            $table->renameColumn('zero_checkout', 'zero_checkout_rule')
                ->update();
        }
        if ($table->hasColumn('zero_checkout_rule')) {
            $table->changeColumn('zero_checkout_rule', 'tinyinteger', ['default' => 0, 'comment' => '结账抹零, 0-实款实收 1-抹分 2-抹角 3-抹元'])
                ->update();
        }
    }
}
