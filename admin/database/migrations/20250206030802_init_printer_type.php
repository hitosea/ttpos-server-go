<?php

use think\facade\Db;
use think\facade\Log;
use think\migration\Migrator;
use think\migration\db\Column;

class AddValueToPrinterType extends Migrator
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
                    'en' => 'Sunmi printer (LAN)',
                    'zh' => '商米打印机（局域网）',
                    'zhtw' => '商米打印机（局域网）',
                    'th' => 'เครื่องพิมพ์ SUNMI (LAN)',
                    'my' => 'SUNMI ပရင်တာ (LAN)',
                    'ja' => 'SUNMIプリンター（LAN）',
                    'ko' => '샹미프린터 (근거리통신망)',
                    'tr' => 'SUNMI yazıcı (LAN)',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'SUNMI_LAN',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Sunmi Printer (Cloud Printing)',
                    'zh' => '商米打印机（云打印）',
                    'zhtw' => '商米打印機（雲打印）',
                    'th' => 'เครื่องพิมพ์ Sunmi (การพิมพ์บนคลาวด์)',
                    'my' => 'Sunmi ပရင်တာ (Cloud Printing)',
                    'ja' => 'Sunmiプリンター（クラウド印刷）',
                    'ko' => 'Sunmi 프린터(클라우드 인쇄)',
                    'tr' => 'Sunmi Yazıcı (Bulut Baskı)',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'SUNMI_CLOUD',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Xinye printer (Wired)',
                    'zh' => '芯烨打印机（有线）',
                    'zhtw' => '芯燁印表機（有線）',
                    'th' => 'เครื่องพิมพ์ Xinye (แบบใช้สาย)',
                    'my' => 'Xinye ပရင်တာ (ကြိုးတပ်)',
                    'ja' => 'Xinyeプリンター（有線）',
                    'ko' => 'Xinye 프린터(유선)',
                    'tr' => 'Xinye yazıcı (kablolu)',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'XPRINTER_LAN',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Xinye Printer (WIFI)',
                    'zh' => '芯烨打印机（WIFI）',
                    'zhtw' => '芯燁印表機（WIFI）',
                    'th' => 'เครื่องพิมพ์ Xinye (WIFI)',
                    'my' => 'Xinye ပရင်တာ (WIFI)',
                    'ja' => '新業プリンター（WIFI）',
                    'ko' => '신예 프린터(와이파이)',
                    'tr' => 'Xinye Yazıcı (WIFI)',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'XPRINTER_WIFI',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Codesoft (Ethernet)',
                    'zh' => 'Codesoft（网口）',
                    'zhtw' => 'Codesoft（網口）',
                    'th' => 'Codesoft (RJ45)',
                    'my' => 'Codesoft (Ethernet)',
                    'ja' => 'Codesoft（LAN）',
                    'ko' => 'Codesoft (LAN)',
                    'tr' => 'Codesoft (RJ45)',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'CODESOFT_LAN',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'name' => [
                    'en' => 'Codesoft（WIFI）',
                    'zh' => 'Codesoft（WIFI）',
                    'zhtw' => 'Codesoft（WIFI）',
                    'th' => 'Codesoft（WIFI）',
                    'my' => 'Codesoft（WIFI）',
                    'ja' => 'Codesoft（WIFI）',
                    'ko' => 'Codesoft（WIFI）',
                    'tr' => 'Codesoft（WIFI）',
                ],
                'multi_language_name_uuid' => 0,
                'key' => 'CODESOFT_WIFI',
                'create_time' => time(),
                'update_time' => time(),
            ]
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
