<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTtposMenuToTakeoutTable extends Migrator
{
    /**
     * 添加 ttpos_menu 字段到外卖平台管理表
     * - 用于存储 TTPOS 导出的菜单数据
     */
    public function change()
    {
        $table = $this->table('takeout');
        
        // 检查字段是否已存在，避免重复添加
        if (!$table->hasColumn('ttpos_menu')) {
            $table->addColumn('ttpos_menu', 'json', ['null' => true, 'after' => 'menu', 'comment' => 'TTPOS导出的菜单数据(JSON格式)'])->update();
        }
    }
}

