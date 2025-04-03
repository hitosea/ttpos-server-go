<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddServiceFeeServiceTaxGiveNumFreeNumToStatisticsProduct extends Migrator
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
        $table = $this->table('statistics_product');
        if (!$table->hasColumn('service_fee') && !$table->hasColumn('service_tax') && !$table->hasColumn('give_num') && !$table->hasColumn('free_num')) {
            $table->addColumn('service_fee', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '服务费', 'after' => 'tax_fee']);
            $table->addColumn('service_tax', 'decimal', ['precision' => 14, 'scale' => 2, 'null' => false, 'default' => 0.00, 'comment' => '服务税费', 'after' => 'service_fee']);
            $table->addColumn('give_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '赠菜数量', 'after' => 'service_tax']);
            $table->addColumn('free_num', 'integer', ['null' => false, 'default' => 0, 'comment' => '免单数量', 'after' => 'give_num']);
            $table->update();
        }
    }
}
