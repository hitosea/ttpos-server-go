<?php

use think\migration\Migrator;

class ChangeColumnSaleOrderRateDefaultValue extends Migrator
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
            $table->removeColumn('member_discount_rate')
                  ->update();
        }
        if ($table->hasColumn('member_card_discount_rate')) {
            $table->removeColumn('member_card_discount_rate')
                  ->update();
        }
        if ($table->hasColumn('custom_discount_rate')) {
            $table->removeColumn('custom_discount_rate')
                  ->update();
        }


        if (!$table->hasColumn('member_discount_rate')) {
            $table->addColumn('member_discount_rate', 'decimal', ['precision' => 10, 'scale' => 2, 'default' => 0, 'comment' => '会员折扣率(0-100%)，默认0%，取值范围0-1，如折扣率为10%，则取值为0.1'])
                             ->update();
        }
        if (!$table->hasColumn('member_card_discount_rate')) {
            $table->addColumn('member_card_discount_rate', 'decimal', ['precision' => 10, 'scale' => 2, 'default' => 0, 'comment' => '会员卡折扣率(0-100%)，默认0%，取值范围0-1，如折扣率为10%，则取值为0.1'])
                             ->update();
        }
        if (!$table->hasColumn('custom_discount_rate')) {
            $table->addColumn('custom_discount_rate', 'decimal', ['precision' => 10, 'scale' => 2, 'default' => 0, 'comment' => '自定义折扣率(0-100%)，默认0%，取值范围0-1，如折扣率为10%，则取值为0.1'])
                             ->update();
        }
        $table->update();
    }
}
