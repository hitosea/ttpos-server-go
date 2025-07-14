<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InitFreeReason extends Migrator
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
        $list = [
            [
                'name' => [
                    'en_name' => 'Customer complaint about quality',
                    'zh_name' => '顾客投诉质量',
                    'zh_tw_name' => '顧客投訴質量',
                    'th_name' => 'ลูกค้าแจ้งเรื่องคุณภาพ',
                    'my_name' => 'ဖောက်သည် အရည်အသွေး ပယ်ဖျက်မှု',
                    'ja_name' => '顧客の品質に対する苦情',
                    'ko_name' => '고객의 품질 불만',
                    'tr_name' => 'Müşteri kalite şikayeti',
                    'sv_name' => 'Kundkvalitetsskrivelse',
                ]
            ],
            [
                'name' => [
                    'en_name' => 'Friendship Discount',
                    'zh_name' => '友情打折',
                    'zh_tw_name' => '友情打折',
                    'th_name' => 'ส่วนลดมิตรภาพ',
                    'my_name' => 'ငှားနေရန် ဂုဏ်ယူစရာ စျေးလျှော့',
                    'ja_name' => '友情割引',
                    'ko_name' => '우정 할인',
                    'tr_name' => 'Arkadaş İndirimi',
                    'sv_name' => 'Vännerabatt',
                ]
            ],
            [
                'name' => [
                    'en_name' => 'temporary activity',
                    'zh_name' => '临时活动',
                    'zh_tw_name' => '臨時活動',
                    'th_name' => 'กิจกรรมชั่วคราว',
                    'my_name' => 'အချိန်ပိုင်းလုပ်ငန်း',
                    'ja_name' => '臨時活動',
                    'ko_name' => '애드혹 활동',
                    'tr_name' => 'geçici etkinlik',
                    'sv_name' => 'Tillfällig aktivitet',
                ]
            ],
        ];
        foreach ($list as $item) {
            $resaon = $db->name('free_reason')->where('name', $item['name']['zh_name'])->find();
            if ($resaon) {
                continue;
            }
            $multiLanguage = $db->name('multi_language_name')->where('zh_name', $item['name']['zh_name'])->find();
            if (!$multiLanguage) {
                $nameUuid = createUuid();
                $language = $item['name'];
                $language['uuid'] = $nameUuid;
                $language['create_time'] = time();
                $language['update_time'] = time();
                $db->name('multi_language_name')->insert($language);
            } else {
                $nameUuid = $multiLanguage['uuid'];
            }
            $db->name('free_reason')->insert([
                'uuid' => createUuid(),
                'name' => $item['name']['zh_name'],
                'multi_language_name_uuid' => $nameUuid,
                'create_time' => time(),
                'update_time' => time(),
            ]);
        }
    }
}
