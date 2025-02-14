<?php

use think\migration\Migrator;

class UpdateDeviceKeyToPrinterLog extends Migrator
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
        $table = $this->table('printer_log');
        if ($table->hasColumn('cashier_device_key')) {
            $table->renameColumn('cashier_device_key', 'cashier_device_id');
            $table->update();
        }
        // 
        $table = $this->table('device');
        if ($table->hasColumn('device_key')) {
            $table->renameColumn('device_key', 'device_id');
            $table->update();
        }
    }
}