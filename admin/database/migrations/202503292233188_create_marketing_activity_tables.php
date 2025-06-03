<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateMarketingActivityTables extends Migrator
{
    public function change()
    {
        // 活动主表
        if (!$this->hasTable('marketing_activity')) {
            $table = $this->table('marketing_activity', ['comment' => '会员营销-活动表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '活动唯一ID'])
                ->addColumn('name', 'string', ['limit' => 50, 'default' => '', 'comment' => '活动名称'])
                ->addColumn('type', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '活动类型 0邀请有礼 1积分商城'])
                ->addColumn('multi_language_name_uuid', 'biginteger', ['default' => 0, 'comment' => '活动名称多语言uuid'])
                ->addColumn('description', 'string', ['limit' => 100, 'default' => '', 'comment' => '活动文案'])
                ->addColumn('multi_language_desc_uuid', 'biginteger', ['default' => 0, 'comment' => '活动文案多语言uuid'])
                ->addColumn('start_time', 'integer', ['default' => 0, 'comment' => '活动开始时间'])
                ->addColumn('end_time', 'integer', ['default' => 0, 'comment' => '活动结束时间'])
                ->addColumn('reward_condition_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '奖励条件金额'])
                ->addColumn('is_open_reward_limit', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否开启奖励次数限制 0否 1是'])
                ->addColumn('reward_limit', 'integer', ['default' => 0, 'comment' => '奖励次数限制'])
                ->addColumn('is_invalid', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否失效 0否 1是'])
                ->addColumn('image_base64', 'text', ['null' => true, 'comment' => '活动图片base64'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }

        // 活动礼品表
        if (!$this->hasTable('marketing_activity_prize')) {
            $prize = $this->table('marketing_activity_prize', ['comment' => '活动礼品表']);
            $prize->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '礼品唯一ID'])
                ->addColumn('activity_uuid', 'biginteger', ['default' => 0, 'comment' => '活动uuid'])
                ->addColumn('prize_type', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '奖品类型 1优惠券 2未知'])
                ->addColumn('prize_uuid', 'biginteger', ['default' => 0, 'comment' => '奖品uuid'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }

        // 奖励发放记录表
        if (!$this->hasTable('marketing_activity_record')) {
            $record = $this->table('marketing_activity_record', ['comment' => '活动奖励发放记录']);
            $record->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '记录唯一ID'])
                ->addColumn('activity_uuid', 'biginteger', ['default' => 0, 'comment' => '活动uuid'])
                ->addColumn('prize_uuid', 'biginteger', ['default' => 0, 'comment' => '奖品uuid'])
                ->addColumn('member_uuid', 'biginteger', ['default' => 0, 'comment' => '会员uuid'])
                ->addColumn('reward_count', 'integer', ['default' => 0, 'comment' => '已获得奖励次数'])
                ->addColumn('last_reward_time', 'integer', ['default' => 0, 'comment' => '最后一次获得奖励时间'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
} 