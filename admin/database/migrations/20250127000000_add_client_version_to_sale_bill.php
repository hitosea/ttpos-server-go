<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddClientVersionToSaleBill extends Migrator
{
    /**
     * 为销售账单表新增 client_version 字段
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('sale_bill')) {
            $table = $this->table('sale_bill');
            
            // 检查字段是否不存在
            if (!$table->hasColumn('client_version')) {
                $table->addColumn('client_version', 'string', [
                    'limit' => 20,
                    'default' => '',
                    'null' => false,
                    'comment' => '客户端版本号（如 2.10.0、2.9.0）',
                    'after' => 'source',
                ]);
            }
            
            $table->update();
        }
    }
}

