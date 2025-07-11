<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class UpdatePrinterType extends Migrator
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
        $data = $db->name('printer_type')->where('key', 'SUNMI_LAN')->find();
        if ($data) {
            $name = json_decode($data['name'], true);
            $name['sv'] = 'SUNMI skrivare (LAN)80mm';
            $db->name('printer_type')->where('key', 'SUNMI_LAN')->update(['name' => json_encode($name)]);
        }

        $data = $db->name('printer_type')->where('key', 'SUNMI_CLOUD')->find();
        if ($data) {
            $name = json_decode($data['name'], true);
            $name['sv'] = 'Sunmi skrivare (Molnbaskylla)80mm';
            $db->name('printer_type')->where('key', 'SUNMI_CLOUD')->update(['name' => json_encode($name)]);
        }

        $data = $db->name('printer_type')->where('key', 'XPRINTER_LAN')->find();
        if ($data) {
            $name = json_decode($data['name'], true);
            $name['sv'] = 'Xinye skrivare (kablolagd)80mm';
            $db->name('printer_type')->where('key', 'XPRINTER_LAN')->update(['name' => json_encode($name)]);
        }

        $data = $db->name('printer_type')->where('key', 'XPRINTER_WIFI')->find();
        if ($data) {
            $name = json_decode($data['name'], true);
            $name['sv'] = 'Xinye skrivare (WIFI)80mm';
            $db->name('printer_type')->where('key', 'XPRINTER_WIFI')->update(['name' => json_encode($name)]);
        }

        $data = $db->name('printer_type')->where('key', 'CODESOFT_LAN')->find();
        if ($data) {
            $name = json_decode($data['name'], true);
            $name['sv'] = 'Codesoft (kablolagd)80mm';
            $db->name('printer_type')->where('key', 'CODESOFT_LAN')->update(['name' => json_encode($name)]);
        }

        $data = $db->name('printer_type')->where('key', 'CODESOFT_WIFI')->find();
        if ($data) {
            $name = json_decode($data['name'], true);
            $name['sv'] = 'Codesoft (WIFI)80mm';
            $db->name('printer_type')->where('key', 'CODESOFT_WIFI')->update(['name' => json_encode($name)]);
        }
    }
}
