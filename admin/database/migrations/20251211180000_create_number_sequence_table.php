<?php
use think\migration\Migrator;
use think\migration\db\Column;

class CreateNumberSequenceTable extends Migrator
{

    const TARGET = 'main';

    /**
     * 创建通用编号序列表（saas库）
     * 用于管理各类编号的按日递增序列（发票、订单、收据等）
     * Requirement: add-invoice-number-for-printing
     */
    public function change()
    {
        if (!$this->hasTable('number_sequence')) {
            $table = $this->table('number_sequence', ['id' => false, 'primary_key' => ['id'], 'engine' => 'InnoDB', 'collation' => 'utf8mb4_general_ci', 'comment' => '通用编号序列表']);
        
            $table->addColumn('id', 'integer', ['limit' => 10, 'signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('company_uuid', 'biginteger', ['limit' => 20, 'signed' => false, 'comment' => '商家UUID'])
                ->addColumn('type', 'string', ['limit' => 32, 'comment' => '编号类型(invoice:发票,order:订单,receipt:收据等)'])
                ->addColumn('date', 'date', ['comment' => '日期'])
                ->addColumn('sequence', 'integer', ['limit' => 10, 'signed' => false, 'default' => 0, 'comment' => '当日序列号'])
                ->addColumn('create_time', 'biginteger', ['limit' => 20, 'signed' => false, 'default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'biginteger', ['limit' => 20, 'signed' => false, 'default' => 0, 'comment' => '更新时间'])
                ->addIndex(['company_uuid', 'type', 'date'], ['unique' => true, 'name' => 'idx_company_uuid_type_date'])
                ->addIndex(['type', 'date'], ['name' => 'idx_type_date'])
                ->create();
        }
        
    }
}

