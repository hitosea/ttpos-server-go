<?php

use Phinx\Migration\AbstractMigration;

class AlterMarketingActivityNameLength extends AbstractMigration
{
    /**
     * 修改活动名称字段长度
     */
    public function change()
    {
        if ($this->hasTable('marketing_activity')) {
            $table = $this->table('marketing_activity');
            if ($table->hasColumn('name')) {
                $table->changeColumn('name', 'string', ['limit' => 2500, 'default' => '', 'comment' => '活动名称'])
                    ->update();
            }
            if ($table->hasColumn('description')) {
                $table->changeColumn('description', 'string', ['limit' => 5000, 'default' => '', 'comment' => '活动描述'])
                    ->update();
            }
        }
       
    }
} 