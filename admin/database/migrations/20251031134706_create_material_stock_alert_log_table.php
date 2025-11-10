<?php

use Phinx\Migration\AbstractMigration;

/**
 * 物料库存预警邮件记录表
 */
class CreateMaterialStockAlertLogTable extends AbstractMigration
{
    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否存在
        if (!$this->hasTable('material_stock_alert_log')) {
            $table = $this->table('material_stock_alert_log', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_unicode_ci',
                'comment' => '物料库存预警邮件记录表'
            ]);

            $table->addColumn('id', 'biginteger', ['signed' => false, 'identity' => true, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识UUID'])
                ->addColumn('message_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '消息UUID，每次发送时随机生成'])
                ->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '公司UUID'])
                ->addColumn('material_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '物料UUID'])
                ->addColumn('warehouse_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '仓库UUID，0表示全部维度'])
                ->addColumn('alert_type', 'integer', ['signed' => false, 'default' => 1, 'comment' => '预警类型：1-公司维度 2-仓库维度'])
                ->addColumn('current_stock', 'decimal', ['precision' => 14, 'scale' => 4, 'default' => 0, 'comment' => '当前库存数量'])
                ->addColumn('safety_stock', 'decimal', ['precision' => 14, 'scale' => 4, 'default' => 0, 'comment' => '安全库存数量'])
                ->addColumn('last_alert_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '上次预警时间（时间戳）'])
                ->addColumn('alert_count', 'integer', ['signed' => false, 'default' => 0, 'comment' => '预警次数'])
                ->addColumn('send_status', 'integer', ['signed' => false, 'default' => 0, 'comment' => '发送状态：0-待发送 1-发送成功 2-发送失败'])
                ->addColumn('recipient', 'string', ['limit' => 500, 'default' => '', 'comment' => '收件人邮箱（多个用逗号分隔）'])
                ->addColumn('error_message', 'text', ['null' => true, 'comment' => '错误信息'])
                ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间（时间戳）'])
                ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间（时间戳）'])
                ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间（时间戳）'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'idx_uuid'])
                ->addIndex(['company_uuid', 'material_uuid', 'warehouse_uuid'], ['name' => 'idx_company_material_warehouse'])
                ->create();
        }
    }
}

