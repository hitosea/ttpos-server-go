<?php
/**
 * 为商品标签表添加 is_show_kiosk 字段
 * 
 * 任务: story-admin-self-service-kiosk-client Phase 4.11
 * 需求: R2.1, R2.7
 */

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsShowKioskToProductLabel extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        // 检查表是否存在
        if ($this->hasTable('product_label')) {
            $table = $this->table('product_label');
            
            // 检查字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('is_show_kiosk')) {
                $table->addColumn('is_show_kiosk', 'integer', [
                    'limit' => 3,
                    'default' => 0,
                    'null' => false,
                    'comment' => '是否在自助点餐机显示, 0-否 1-是',
                    'after' => 'is_show_menu'
                ]);
            }
            
            $table->update();
        }
    }
}

