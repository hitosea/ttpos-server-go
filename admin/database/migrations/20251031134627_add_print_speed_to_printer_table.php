<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddPrintSpeedToPrinterTable extends Migrator
{
    /**
     * 添加 print_speed 字段到 ttpos_printer 表
     * 打印速度：1-流畅（不分片打印），2-稳定（分片大包打印），3-兼容（分片小包打印）
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('printer')) {
            $table = $this->table('printer');
            
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('print_speed')) {
                $table->addColumn('print_speed', 'integer', ['limit' => 1, 'default' => 3, 'comment' => '打印速度：1-流畅（不分片打印），2-稳定（分片大包打印），3-兼容（分片小包打印）', 'after' => 'status'])->update();
            }
        }

        // 检查表是否存在
        if ($this->hasTable('printer_log')) {
            $table = $this->table('printer_log');
            
            // 检查字段是否已存在，如果不存在则添加
            if (!$table->hasColumn('print_speed')) {
                $table->addColumn('print_speed', 'integer', ['limit' => 1, 'default' => 3, 'comment' => '打印速度：1-流畅（不分片打印），2-稳定（分片大包打印），3-兼容（分片小包打印）', 'after' => 'copies'])->update();
            }
        }
    }
}
