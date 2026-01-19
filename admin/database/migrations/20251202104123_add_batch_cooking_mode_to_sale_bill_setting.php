<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddBatchCookingModeToSaleBillSetting extends Migrator
{
    /**
     * 为销售账单设置表新增 batch_cooking_mode 字段
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill_setting')) {
            $table = $this->table('sale_bill_setting');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('batch_cooking_mode')) {
                $table->addColumn('batch_cooking_mode', 'string', [
                    'limit' => 10,
                    'default' => 'post',
                    'null' => false,
                    'comment' => '分批送厨模式: pre-前置 / post-后置，默认 post',
                    'after' => 'auto_points_exchange',
                ]);
            }
            
            $table->update();
            
            // 为历史数据设置默认值
            $this->execute("UPDATE `ttpos_sale_bill_setting` SET `batch_cooking_mode` = 'post' WHERE `batch_cooking_mode` IS NULL OR `batch_cooking_mode` = ''");
        }
    }
}

