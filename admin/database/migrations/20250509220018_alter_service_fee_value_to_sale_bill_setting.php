<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AlterServiceFeeValueToSaleBillSetting extends Migrator
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
        if ($table->hasColumn('service_fee_value')) {
            $table->changeColumn('service_fee_value','decimal', ['precision' => 12, 'scale' => 4, 'null' => false, 'default' => 0, 'comment' => '服务费值,服务费类型为1时,服务费值为固定金额,服务费类型为2和3时,服务费值为%比例', 'after' => 'service_fee_type']);
        }
        $table->update();
    }
}
