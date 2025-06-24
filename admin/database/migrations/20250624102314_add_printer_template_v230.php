<?php

use think\facade\Db;
use think\facade\Log;
use think\migration\Migrator;
use think\migration\db\Column;

class AddPrinterTemplateV230 extends Migrator
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
        $db = Db::connect(Db::getConfig('default'), true);

        $printerTemplateList = [
            [
                'id' => 10,
                'uuid' => 10,
                'name' => '出菜单',
                'template' => 1,
                'create_time' => time(),
                'update_time' => time(),
            ],
        ];

        foreach ($printerTemplateList as $item) {
            // 判断重复
            $printerTemplate = $db->name('printer_template')->where('uuid', $item['uuid'])->find();
            if ($printerTemplate) {
                continue;
            }
            // 添加打印机模板
            $db->name('printer_template')->insert($item);
        }
    }
}
