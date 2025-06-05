<?php

use think\migration\Migrator;
use think\migration\db\Column;

class CreateMarketingActivityConsumptionTable extends Migrator
{
    public function change()
    {
        if (!$this->hasTable('marketing_activity_consumption')) {
            $table = $this->table('marketing_activity_consumption', ['comment' => '活动消费记录表']);
            $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '消费记录唯一ID'])
                ->addColumn('activity_uuid', 'biginteger', ['default' => 0, 'comment' => '活动uuid'])
                ->addColumn('referrer_uuid', 'biginteger', ['default' => 0, 'comment' => '推荐人uuid'])
                ->addColumn('consumer_uuid', 'biginteger', ['default' => 0, 'comment' => '消费人uuid'])
                ->addColumn('consumption_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '消费金额'])
                ->addColumn('reward_amount', 'decimal', ['precision' => 14, 'scale' => 2, 'default' => 0, 'comment' => '奖励金额'])
                ->addColumn('reward_status', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '奖励状态 0未发放 1已发放'])
                ->addColumn('reward_num', 'integer', ['default' => 0, 'comment' => '奖励次数'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->create();
        }
    }
} 