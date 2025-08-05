<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIsShowSkuToPrinterTemplate extends Migrator
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
        $table = $this->table('printer_template');
        // 检查字段是否已经存在
        if (!$table->hasColumn('is_show_sku')) {
            // 添加 is_show_sku 字段
            $table->addColumn('is_show_sku', 'integer', [
                'limit' => 1,
                'default' => 1,
                'comment' => '是否显示SKU：0=不显示，1=显示',
                'after' => 'template'
            ])->update();
        }
    }
}