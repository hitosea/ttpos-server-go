<?php

use think\migration\Migrator;
use think\migration\db\Column;

/**
 * 任务 #39752: Grab/LINE MAN-单规格不推送规格内的选项
 *
 * 为 ttpos_takeout 表添加 single_flavor_mapping 字段
 * 用于存储被跳过推送的单规格商品信息，订单回传时用于自动补全规格
 */
class AddSingleFlavorMappingToTakeout extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        $table = $this->table('takeout');

        // 检查字段是否存在
        if (!$table->hasColumn('single_flavor_mapping')) {
            $table->addColumn('single_flavor_mapping', 'json', [
                'null' => true,
                'default' => null,
                'comment' => '单规格映射(JSON格式)：记录被跳过推送的单规格商品信息',
                'after' => 'ttpos_menu',
            ])->update();
        }
    }
}
