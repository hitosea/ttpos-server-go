<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddColumnsToCompanySetting extends Migrator
{
    // 迁移目标
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
        $table = $this->table('company_setting');
        if (!$table->hasColumn('delivery_config')) {
            $table->addColumn('delivery_config', 'text', ['default' => '', 'comment' => '外送配置', 'after' => 'address']);
        }
        if (!$table->hasColumn('delivery_status')) {
            $table->addColumn('delivery_status', 'integer', ['default' => '0', 'comment' => '外送配置状态：0-关,1-开', 'after' => 'address']);
        }
        $table->update();
    }
}
