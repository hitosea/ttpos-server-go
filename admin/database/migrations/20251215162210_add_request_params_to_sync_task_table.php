<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRequestParamsToSyncTaskTable extends Migrator
{
    /**
     * 添加请求参数字段到 ttpos_sync_task 表
     */
    public function change()
    {
        $table = $this->table('sync_task');

        // 检查字段是否已存在，如果不存在则添加
        if (!$table->hasColumn('request_params')) {
            $table->addColumn('request_params', 'text', [
                'null' => true,
                'comment' => '请求参数(JSON格式)',
                'after' => 'end_time'
            ])->update();
        }
    }
}
