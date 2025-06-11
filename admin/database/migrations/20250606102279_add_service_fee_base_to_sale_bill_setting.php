<?php

use think\facade\Db;
use think\facade\Config;
use think\migration\Migrator;

class AddServiceFeeBaseToSaleBillSetting extends Migrator
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
        if (!$table->hasColumn('service_fee_base')) {
            $table->addColumn('service_fee_base', 'integer', ['default' => 0, 'comment' => '服务费计算基准, 0-商品惠后价 1-商品价格合计', 'after' => 'service_apply']);
            $table->update();
        }
    }
}
