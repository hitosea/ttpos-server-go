<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnMemberUuidToMarketingCouponRecord extends Migrator
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
        // 优惠券记录关联会员
        $table = $this->table('marketing_coupon_record');
        if (!$table->hasColumn('member_uuid')) {
            $table->addColumn('member_uuid', 'biginteger', ['default' => 0, 'null' => false, 'comment' => '关联会员Uuid', 'after' => 'left_count']);
        }
        $table->update();
    }
}
