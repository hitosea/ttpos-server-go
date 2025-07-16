<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddMemberSaleOrderAddressTable extends Migrator
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
        if (!$this->hasTable('member_sale_order_address')) {
            $table = $this->table('member_sale_order_address', ['comment' => '会员销售订单地址表,订单的地址信息']);
            $table->addColumn('uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '会员销售订单地址ID'])
            ->addColumn('member_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '会员UUID'])
            ->addColumn('member_address_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '会员收货地址UUID'])
            ->addColumn('longitude', 'decimal', ['precision' => 12, 'scale' => 6, 'null' => false, 'default' => 0.00, 'comment' => '经度'])
            ->addColumn('latitude', 'decimal', ['precision' => 12, 'scale' => 6, 'null' => false, 'default' => 0.00, 'comment' => '纬度'])
            ->addColumn('address', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '地址'])
            ->addColumn('detail_address', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '详细地址'])
            ->addColumn('contact_name', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系人'])
            ->addColumn('contact_phone', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '联系电话'])
            ->addColumn('contact_gender', 'integer', ['null' => false, 'default' => 0, 'comment' => '联系人性别, 0-女士 1-先生'])
            ->addColumn('member_sale_order_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '会员销售订单UUID'])
            ->addColumn('create_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '创建时间(时间戳)'])
            ->addColumn('update_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '更新时间(时间戳)'])
            ->addColumn('delete_time', 'integer', ['null' => false, 'default' => 0, 'comment' => '删除时间(时间戳)'])
            ->create();
        }
    }
}
