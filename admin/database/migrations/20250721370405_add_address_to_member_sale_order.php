<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddAddressToMemberSaleOrder extends Migrator
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


        $table = $this->table('member_sale_order_address');
        if ($table->exists()) {
            $table->drop()->save();
        }

        $table = $this->table('member_sale_order');
        if (!$table->hasColumn('member_address_uuid')) {
            $table->addColumn('member_address_uuid', 'biginteger', [unsigned => true, 'null' => false, 'default' => 0, 'comment' => '会员收货地址UUID', 'after' => 'related_order_type']);
            $table->update();
        }
         if (!$table->hasColumn('contact_location')) {
            $table->addColumn('contact_location', 'string', ['null' => false, 'default' => '', 'comment' => '位置坐标', 'after' => 'member_address_uuid']);
            $table->update();
        }
         if (!$table->hasColumn('contact_address')) {
            $table->addColumn('contact_address', 'string', ['null' => false, 'default' => '', 'comment' => '详细地址', 'after' => 'contact_location']);
            $table->update();
        }
        if (!$table->hasColumn('contact_address_detail')) {
            $table->addColumn('contact_address_detail', 'string', ['null' => false, 'default' => '', 'comment' => '详细地址', 'after' => 'contact_address']);
            $table->update();
        }
         if (!$table->hasColumn('contact_name')) {
            $table->addColumn('contact_name', 'string', ['null' => false, 'default' => '', 'comment' => '联系人', 'after' => 'contact_address_detail']);
            $table->update();
        }
          if (!$table->hasColumn('contact_phone')) {
            $table->addColumn('contact_phone', 'string', ['null' => false, 'default' => '', 'comment' => '联系电话', 'after' => 'contact_name']);
            $table->update();
        }
        if (!$table->hasColumn('contact_phone_prefix')) {
            $table->addColumn('contact_phone_prefix', 'string', ['null' => false, 'default' => '', 'comment' => '联系电话前缀', 'after' => 'contact_phone']);
            $table->update();
        }
        if (!$table->hasColumn('contact_gender')) {
            $table->addColumn('contact_gender', 'integer', ['null' => false, 'default' => 0, 'comment' => '联系人性别, 0-女士 1-先生', 'after' => 'contact_phone_prefix']);
            $table->update();
        }
    }
}
