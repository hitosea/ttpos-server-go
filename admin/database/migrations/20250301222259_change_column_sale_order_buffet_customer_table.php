<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ChangeColumnSaleOrderBuffetCustomerTable extends Migrator
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
        $table = $this->table('sale_order_buffet_customer_type');
        if ($table->hasColumn('buffet_package_multi_language_name_uuid')) {
            $table->removeColumn('buffet_package_multi_language_name_uuid')->update();
        }
        if ($table->hasColumn('buffet_customer_type_multi_language_name_uuid')) {
            $table->removeColumn('buffet_customer_type_multi_language_name_uuid')->update();
        }
        if ($table->hasColumn('customer_price')) {
            $table->renameColumn('customer_price', 'sale_price')->update();
        }
        if ($table->hasColumn('custom_discount_rate')) {
            $table->changeColumn('custom_discount_rate', 'decimal', ['precision' => 12, 'scale' => 4, 'default' => 1, 'comment' => '自定义折扣率, 值为0-1之间(0-100%)'])->update();
        }
        if ($table->hasColumn('amount')) {
            $table->changeColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '应收金额(单人)。自助餐顾客类型已含税时，应收金额(单人)=(自助餐顾客类型原价-自助餐顾客类型税费)+服务费+自助餐顾客类型税费；自助餐顾客类型未含税时，应收金额(单人)=自助餐顾客类型原价+服务费+自助餐顾客类型税费'])->update();
            $table->renameColumn('amount', 'total_price')->update();
        }
    }
}
