<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 为 ttpos_multi_language_name 表添加 is_full_overwrite 字段
 */
class AddIsFullOverwriteToMultiLanguageName extends Migrator
{
    /**
     * 执行迁移
     */
    public function change()
    {
        $table = $this->table('multi_language_name');
        
        // 检查字段是否已存在
        if (!$table->hasColumn('not_overwrite')) {
            $table->addColumn('not_overwrite', 'integer', [
                'limit' => 11,
                'default' => 0,
                'null' => false,
                'comment' => '不要覆盖 0-否 1-是',
                'after' => 'sv_name'
            ])
            ->update();
        }
    }
}
