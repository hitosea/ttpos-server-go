<?php

use think\migration\Migrator;

/**
 * 添加时间节点字段到操作耗时记录表
 *
 * 用于记录请求处理过程中各阶段的耗时数据
 */
class AddTimeNodesToOperationDuration extends Migrator
{
    // 迁移目标：仅应用到 SaaS 主库
    const TARGET = 'main';

    /**
     * Change Method.
     */
    public function change()
    {
        $table = $this->table('order_operation_duration');

        // 检查字段是否已存在
        if (!$table->hasColumn('time_nodes')) {
            $table->addColumn('time_nodes', 'text', [
                'null' => true,
                'after' => 'error_msg',
                'comment' => '时间节点JSON([{name,offset_ms}])',
            ])->update();
        }
    }
}
