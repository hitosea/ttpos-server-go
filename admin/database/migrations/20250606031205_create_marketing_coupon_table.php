<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateMarketingCouponTable extends Migrator
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
        // 优惠券表
        if (!$this->hasTable('marketing_coupon')) {
            $table = $this->table('marketing_coupon', ['comment' => '会员营销-优惠券表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '优惠券唯一ID'])
                ->addColumn('name', 'string', ['limit' => 50, 'default' => '', 'comment' => '优惠券名称'])
                ->addColumn('sort', 'integer', ['default' => 0, 'comment' => '排序, 1-99'])
                ->addColumn('type', 'string', ['limit' => 20, 'default' => '', 'comment' => '优惠券类型: deduction - 抵扣券'])
                ->addColumn('deduction_type', 'string', ['limit' => 20, 'default' => '', 'comment' => '抵扣类型: taxed - 税后抵扣'])
                ->addColumn('amount', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '优惠券金额'])
                ->addColumn('count', 'integer', ['default' => 0, 'comment' => '优惠券数量, 最大999999'])
                ->addColumn('day_start_time', 'string', ['limit' => 5, 'default' => '', 'comment' => '每日适用时段开始时间, hh:mm 格式'])
                ->addColumn('day_end_time', 'string', ['limit' => 5, 'default' => '', 'comment' => '每日适用时段结束时间, hh:mm 格式'])
                ->addColumn('requirement', 'string', ['limit' => 20, 'default' => '', 'comment' => '获得优惠券所需条件: none - 都可以获取; marketing - 营销活动'])
                ->addColumn('valid_start_time', 'integer', ['default' => 0, 'comment' => '优惠券有效开始时间, requirement = none 时有效'])
                ->addColumn('valid_end_time', 'integer', ['default' => 0, 'comment' => '优惠券有效结束时间, requirement = none 时有效'])
                ->addColumn('valid_days', 'integer', ['default' => 0, 'comment' => '领取优惠券后n天内有效, requirement = marketing 时有效'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }
        // 优惠券记录表
        if (!$this->hasTable('marketing_coupon_record')) {
            $table = $this->table('marketing_coupon_record', ['comment' => '会员营销-优惠券记录表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '优惠券记录唯一ID'])
                ->addColumn('coupon_uuid', 'biginteger', ['default' => 0, 'comment' => '优惠券唯一ID'])
                ->addColumn('serial_no', 'string', ['default' => '', 'comment' => '记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用'])
                ->addColumn('type', 'integer', ['default' => '1', 'comment' => '记录类型：1-首次添加、2-调整添加、3-调整扣减'])
                ->addColumn('count', 'integer', ['default' => 0, 'comment' => '变动数量'])
                ->addColumn('left_count', 'integer', ['default' => 0, 'comment' => '剩余有效张数'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
}
