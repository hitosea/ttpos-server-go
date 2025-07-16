<?php

use think\facade\Db;
use think\migration\Migrator;

class AddColumnIsVerifiedPhoneToMemberSaleOrder extends Migrator
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
        // 添加字段
        $table = $this->table('member_sale_order');
        if (!$table->hasColumn('is_verified_phone')) {
            $table->addColumn('is_verified_phone', 'integer', ['limit' => 11, 'null' => false, 'default' => 0, 'comment' => '订单是否已经验证手机号,0-未验证 1-已验证,不再弹出验证手机号', 'after' => 'remark']);
            $table->update();
        }
        if (!$table->hasColumn('payment_method_uuid')) {
            $table->addColumn('payment_method_uuid', 'biginteger', ['limit' => 20, 'null' => false, 'default' => 0, 'comment' => '支付方式UUID,订单已选择的支付方式', 'after' => 'is_verified_phone']);
            $table->update();
        }
    }
}
