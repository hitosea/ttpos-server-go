<?php

use think\facade\Db;
use think\facade\Log;
use think\migration\Migrator;
use think\migration\db\Column;

class InitPrinterType extends Migrator
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
                    'en' => 'Sunmi printer (LAN)80mm',
                    'zh' => '商米打印机（局域网）80mm',
                    'zhtw' => '商米打印机（局域网）80mm',
                    'th' => 'เครื่องพิมพ์ SUNMI (LAN)80mm',
                    'my' => 'SUNMI ပရင်တာ (LAN)80mm',
                    'ja' => 'SUNMIプリンター（LAN）80mm',
                    'ko' => '샹미프린터 (근거리통신망)80mm',
                    'tr' => 'SUNMI yazıcı (LAN)80mm',
                    'sv' => 'SUNMI skrivare (LAN)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'SUNMI_LAN',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Sunmi Printer (Cloud Printing)80mm',
                    'zh' => '商米打印机（云打印）80mm',
                    'zhtw' => '商米打印機（雲打印）80mm',
                    'th' => 'เครื่องพิมพ์ Sunmi (การพิมพ์บนคลาวด์)80mm',
                    'my' => 'Sunmi ပရင်တာ (Cloud Printing)80mm',
                    'ja' => 'Sunmiプリンター（クラウド印刷）80mm',
                    'ko' => 'Sunmi 프린터(클라우드 인쇄)80mm',
                    'tr' => 'Sunmi Yazıcı (Bulut Baskı)80mm',
                    'sv' => 'Sunmi skrivare (Molnbaskylla)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'SUNMI_CLOUD',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Xinye printer (Wired)80mm',
                    'zh' => '芯烨打印机（有线）80mm',
                    'zhtw' => '芯燁印表機（有線）80mm',
                    'th' => 'เครื่องพิมพ์ Xinye (แบบใช้สาย)80mm',
                    'my' => 'Xinye ပရင်တာ (ကြိုးတပ်)80mm',
                    'ja' => 'Xinyeプリンター（有線）80mm',
                    'ko' => 'Xinye 프린터(유선)80mm',
                    'tr' => 'Xinye yazıcı (kablolu)80mm',
                    'sv' => 'Xinye skrivare (kablolagd)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'XPRINTER_LAN',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Xinye Printer (WIFI)80mm',
                    'zh' => '芯烨打印机（WIFI）80mm',
                    'zhtw' => '芯燁印表機（WIFI）80mm',
                    'th' => 'เครื่องพิมพ์ Xinye (WIFI)80mm',
                    'my' => 'Xinye ပရင်တာ (WIFI)80mm',
                    'ja' => '新業プリンター（WIFI）80mm',
                    'ko' => '신예 프린터(와이파이)80mm',
                    'tr' => 'Xinye Yazıcı (WIFI)80mm',
                    'sv' => 'Xinye skrivare (WIFI)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'XPRINTER_WIFI',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Codesoft (Ethernet)80mm',
                    'zh' => 'Codesoft（网口）80mm',
                    'zhtw' => 'Codesoft（網口）80mm',
                    'th' => 'Codesoft (RJ45)80mm',
                    'my' => 'Codesoft (Ethernet)80mm',
                    'ja' => 'Codesoft（LAN）80mm',
                    'ko' => 'Codesoft (LAN)80mm',
                    'tr' => 'Codesoft (RJ45)80mm',
                    'sv' => 'Codesoft (kablolagd)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'CODESOFT_LAN',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Codesoft（WIFI）80mm',
                    'zh' => 'Codesoft（WIFI）80mm',
                    'zhtw' => 'Codesoft（WIFI）80mm',
                    'th' => 'Codesoft（WIFI）80mm',
                    'my' => 'Codesoft（WIFI）80mm',
                    'ja' => 'Codesoft（WIFI）80mm',
                    'ko' => 'Codesoft（WIFI）80mm',
                    'tr' => 'Codesoft（WIFI）80mm',
                    'sv' => 'Codesoft (WIFI)80mm',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'CODESOFT_WIFI',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'GP Cloud',
                    'zh' => '佳博（云打印）',
                    'zhtw' => '佳博（雲打印）',
                    'th' => 'เครื่องพิมพ์ GP (คลาวด์)',
                    'my' => 'GP ပရင်တာ (Cloud)',
                    'ja' => 'GPプリンター（クラウド）',
                    'ko' => 'GP 프린터(클라우드)',
                    'tr' => 'GP Yazıcı (Bulut)',
                    'sv' => 'GP skrivare (Moln)',
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
