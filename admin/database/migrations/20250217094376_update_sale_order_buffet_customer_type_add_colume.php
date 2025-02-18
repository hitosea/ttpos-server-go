<?php

use think\migration\Migrator;
use think\migration\db\Column;

class UpdateSaleOrderBuffetCustomerTypeAddColume extends Migrator
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
        $table = $this->table('sale_order_buffet_customer_type');
        if ($table->hasColumn('price')) {
            $table->removeColumn('price')
                  ->update();
        }
        if (!$table->hasColumn('customer_price')) {
            $table->addColumn('customer_price', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变', 'after' => 'num'])
                  ->update();
        }
        if (!$table->hasColumn('price')) {
            $table->addColumn('price', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '价格（折后价），只进行自定义打折，不进行会员打折', 'after' => 'customer_price'])
                  ->update();
        }
        if (!$table->hasColumn('custom_discount_rate')) {
            $table->addColumn('custom_discount_rate', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '自定义折扣率(0-100%)', 'after' => 'price'])
                  ->update();
        }
        if (!$table->hasColumn('custom_discount_fee')) {
            $table->addColumn('custom_discount_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*(1-自定义折扣率)', 'after' => 'custom_discount_rate'])
                  ->update();
        }
        if (!$table->hasColumn('tax_rate')) {
            $table->addColumn('tax_rate', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '税率,单位%.加购时记录税率,结账时再重新核算', 'after' => 'custom_discount_fee'])
                  ->update();
        }
        if (!$table->hasColumn('service_tax_fee')) {
            $table->addColumn('service_tax_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率', 'after' => 'tax_rate'])
                  ->update();
        }
        if (!$table->hasColumn('tax_fee')) {
            $table->addColumn('tax_fee', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '税费（单人）,0-不收取税费；收取时，税费=自助餐顾客类型原价*税率', 'after' => 'service_tax_fee'])
                  ->update();
        }
        if (!$table->hasColumn('amount')) {
            $table->addColumn('amount', 'decimal', ['precision' => 12, 'scale' => 2, 'default' => 0, 'comment' => '应收金额(单人)。自助餐顾客类型已含税时，应收金额(单人)=(自助餐顾客类型原价-自助餐顾客类型税费)+服务费+自助餐顾客类型税费；自助餐顾客类型未含税时，应收金额(单人)=自助餐顾客类型原价+服务费+自助餐顾客类型税费', 'after' => 'tax_fee'])
                  ->update();
        }
        
    }
}
