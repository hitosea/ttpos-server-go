<?php
/**
 * 为商家设置表添加Grab外卖开关字段
 * 
 * 任务: story-cloud-platform-merchant-grab-delivery-toggle Phase 1.1
 * 需求: R1.1, R1.7
 */

use think\migration\Migrator;

class AddEnableGrabDeliveryToCompanySetting extends Migrator
{
    // 迁移目标：所有商户数据库
    const TARGET = 'all';

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
        if ($this->hasTable('company_setting')) {
            $table = $this->table('company_setting');
            
            // 检查 enable_grab_delivery 字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('enable_grab_delivery')) {
                $table->addColumn('enable_grab_delivery', 'integer', [
                    'limit' => 3,
                    'default' => 0,
                    'null' => false,
                    'comment' => '是否启用Grab外卖：0-否；1-是',
                    'after' => 'enable_kiosk'
                ]);
            }
            
            $table->update();
        }
    }
}

