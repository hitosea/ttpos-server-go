<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMultiLanguageUuidToOrderBuffetCustomerType extends Migrator
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
        if (!$table->hasColumn('buffet_package_multi_language_name_uuid')) {
            $table->addColumn('buffet_package_multi_language_name_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '自助餐套餐多语言ID',
                'after' => 'buffet_customer_type_uuid'
            ])->update();
        }
        if (!$table->hasColumn('buffet_customer_type_multi_language_name_uuid')) {
            $table->addColumn('buffet_customer_type_multi_language_name_uuid', 'biginteger', [
                'signed' => false,
                'null' => false,
                'default' => 0,
                'comment' => '自助餐客户类型多语言ID',
                'after' => 'buffet_package_multi_language_name_uuid'
            ])->update();
        }
    }
}
