<?php

use think\migration\Migrator;

/**
 * 将采购订单操作日志表中的 content 字段类型由 TEXT 调整为 LONGTEXT
 *
 * 目标表：ttpos_purchase_order_log（迁移中使用不带 ttpos_ 前缀的表名）
 * 变更内容：
 * - content: TEXT -> LONGTEXT
 */
class ModifyContentLongtextInPurchaseOrderLog extends Migrator
{
    /**
     * Change Method.
     *
     * 修改 ttpos_purchase_order_log.content 字段为 LONGTEXT
     */
    public function change()
    {
        // 表不存在则直接返回，避免异常
        // 注意：这里使用的是不带 ttpos_ 前缀的表名（框架会自动加表前缀）
        if (!$this->hasTable('purchase_order_log')) {
            return;
        }

        $table = $this->table('purchase_order_log');

        // 修改 content 字段类型为 LONGTEXT，保留原注释
        if ($table->hasColumn('content')) {
            $table->changeColumn('content', 'text', [
                'limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_LONG,
                'null' => true,
                'comment' => '操作内容详情',
            ])->update();
        }
    }
}

