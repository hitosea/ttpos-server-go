<?php

use think\facade\Db;
use think\migration\Migrator;

class AddNewPrinterTypes extends Migrator
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

        $printerTypeList = [
            [
                'name' => [
                    'en' => 'GP Cloud 80mm',
                    'zh' => '佳博（云打印）80mm',
                    'zhtw' => '佳博（雲打印）80mm',
                    'th' => 'เครื่องพิมพ์ GP (คลาวด์)80mm',
                    'my' => 'GP ပရင်တာ (Cloud)80mm',
                    'ja' => 'GPプリンター（クラウド）80mm',
                    'ko' => 'GP 프린터(클라우드)80mm',
                    'tr' => 'GP Yazıcı (Bulut)80mm',
                    'sv' => 'GP skrivare (Moln)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'GP_CLOUD',
                'create_time' => time(),
                'update_time' => time(),
            ],
        ];
        foreach ($printerTypeList as $item) {
            // 判断重复
            $printerType = $db->name('printer_type')->where('key', $item['key'])->find();
            if ($printerType) {
                continue;
            }
            // 添加语言包
            $multiLanguage = $db->name('multi_language_name')->where('zh_name', $item['name']['zh'])->find();
            if (!$multiLanguage) {
                $nameUuid = createUuid();
                $language = [
                    'uuid' => $nameUuid,
                    'create_time' => $item['create_time'],
                    'update_time' => $item['update_time'],
                ];
                foreach ($item['name'] as $key => $value) {
                    if ($key == 'zhtw') {
                        $language['zh_tw_name'] = $value;
                    } else {
                        $language[$key . '_name'] = $value;
                    }
                }
                $db->name('multi_language_name')->insert($language);
            } else {
                $nameUuid = $multiLanguage['uuid'];
            }
            // 添加打印机类型
            $item['uuid'] = createUuid();
            $item['name'] = json_encode($item['name'], JSON_UNESCAPED_UNICODE);
            $item['multi_language_name_uuid'] = $nameUuid;
            $db->name('printer_type')->insert($item);
        }
    }
}