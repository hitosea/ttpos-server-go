<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InitReturnFoodReason extends Migrator
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
            ['name' => [
                'en_name' => 'Long waiting time',
                'zh_name' => '等待时间长',
                'zh_tw_name' => '等待時間長',
                'th_name' => 'เวลารอคอยนาน',
                'my_name' => 'စောင့်နေချိန်ရှည်',
                'ja_name' => '待機時間が長い',
                'ko_name' => '긴 대기 시간',
                'tr_name' => 'bekleme süresi uzun',
            ]],
            ['name' => [
                'en_name' => 'Bad taste',
                'zh_name' => '口味不好',
                'zh_tw_name' => '口味不好',
                'th_name' => 'รสชาติไม่ดี',
                'my_name' => 'အရသာမကောင်းဘူး',
                'ja_name' => '味が悪い',
                'ko_name' => '나쁜 맛',
                'tr_name' => 'tadı kötü',
            ]],
            ['name' => [
                'en_name' => 'Increase the quantity',
                'zh_name' => '多点',
                'zh_tw_name' => '多點',
                'th_name' => 'เพิ่มจำนวน',
                'my_name' => 'အရေအတွက်များစွာဖြစ်လိုက်သည်',
                'ja_name' => '数量を増やす',
                'ko_name' => '추가 수량',
                'tr_name' => 'miktarı artırmak',
            ]],
        ];
        foreach ($list as $item) {
            $resaon = $db->name('return_food_reason')->where('name', $item['name']['zh_name'])->find();
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
            $db->name('return_food_reason')->insert([
                'uuid' => createUuid(),
                'name' => $item['name']['zh_name'],
                'multi_language_name_uuid' => $nameUuid,
                'create_time' => time(),
                'update_time' => time(),
            ]);
        }
    }
}
