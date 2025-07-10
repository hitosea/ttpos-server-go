<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddRelatedPrinterUuidToDeviceTable extends Migrator
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
        $table = $this->table('device');
        if (!$table->hasColumn('related_printer_uuid')) {
            $table->addColumn('related_printer_uuid', 'biginteger', ['signed' => false, 'null' => false, 'default' => 0, 'comment' => '关联打印机uuid,表示该设备关联的打印机uuid', 'after' => 'device_id'])
                ->update();
        }
    }
}
