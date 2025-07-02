<?php

use think\migration\Migrator;

class UpdateDataDeliveryToPrinterTemplate extends Migrator
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
        if ($table->hasColumn('delivery_config')) {
            $table->changeColumn('delivery_config', 'text', ['comment' => '外送配置', 'after' => 'address']);
            $table->update();
        }
    }
}
