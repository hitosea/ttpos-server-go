<?php
use think\migration\Migrator;

class SeedOrderSourceAndNationalityData extends Migrator
{
    /**
     * 预置外卖来源和国籍默认数据
     */
    public function change()
    {
        // 预置外卖来源数据
        $this->seedOrderSources();
        
        // 预置国籍数据
        $this->seedNationalities();
    }

    /**
     * 预置外卖来源数据
     */
    private function seedOrderSources()
    {
        $orderSources = [
            [
                'name_zh' => 'Grab',
                'name_en' => 'Grab',
                'name_th' => 'Grab',
                'name_zh_tw' => 'Grab',
                'name_my' => 'Grab',
                'name_ja' => 'Grab',
                'name_ko' => 'Grab',
                'name_tr' => 'Grab',
                'name_sv' => 'Grab',
                'sort' => 1,
            ],
            [
                'name_zh' => 'LINE MAN',
                'name_en' => 'LINE MAN',
                'name_th' => 'LINE MAN',
                'name_zh_tw' => 'LINE MAN',
                'name_my' => 'LINE MAN',
                'name_ja' => 'LINE MAN',
                'name_ko' => 'LINE MAN',
                'name_tr' => 'LINE MAN',
                'name_sv' => 'LINE MAN',
                'sort' => 2,
            ],
            [
                'name_zh' => '悟空外卖',
                'name_en' => 'Wukong Delivery',
                'name_th' => 'Wukong Delivery',
                'name_zh_tw' => '悟空外賣',
                'name_my' => 'Wukong Delivery',
                'name_ja' => 'Wukong Delivery',
                'name_ko' => 'Wukong Delivery',
                'name_tr' => 'Wukong Delivery',
                'name_sv' => 'Wukong Delivery',
                'sort' => 3,
            ],
            [
                'name_zh' => 'Foodpanda',
                'name_en' => 'Foodpanda',
                'name_th' => 'Foodpanda',
                'name_zh_tw' => 'Foodpanda',
                'name_my' => 'Foodpanda',
                'name_ja' => 'Foodpanda',
                'name_ko' => 'Foodpanda',
                'name_tr' => 'Foodpanda',
                'name_sv' => 'Foodpanda',
                'sort' => 4,
            ],
        ];

        foreach ($orderSources as $source) {
            // 检查是否已存在
            $exists = $this->fetchRow("SELECT id FROM ttpos_multi_language_name WHERE zh_name = '{$source['name_zh']}' AND en_name = '{$source['name_en']}' AND delete_time = 0");
            
            if (!$exists) {
                // 生成 UUID
                $multiLangUuid = createUuid();
                $orderSourceUuid = createUuid();
                $currentTime = time();
                
                // 插入多语言名称
                $this->execute("INSERT INTO ttpos_multi_language_name (uuid, zh_name, en_name, zh_tw_name, th_name, my_name, ja_name, ko_name, tr_name, sv_name, create_time, update_time, delete_time) VALUES ({$multiLangUuid}, '{$source['name_zh']}', '{$source['name_en']}', '{$source['name_zh_tw']}', '{$source['name_th']}', '{$source['name_my']}', '{$source['name_ja']}', '{$source['name_ko']}', '{$source['name_tr']}', '{$source['name_sv']}', {$currentTime}, {$currentTime}, 0)");
                
                // 插入外卖来源
                $this->execute("INSERT INTO ttpos_order_source (uuid, multi_language_name_uuid, sort, status, create_time, update_time, delete_time) VALUES ({$orderSourceUuid}, {$multiLangUuid}, {$source['sort']}, 1, {$currentTime}, {$currentTime}, 0)");
            }
        }
    }

    /**
     * 预置国籍数据
     */
    private function seedNationalities()
    {
        $nationalities = [
            [
                'name_zh' => '泰国',
                'name_en' => 'Thailand',
                'name_th' => 'ไทย',
                'name_zh_tw' => '泰國',
                'name_my' => 'ထိုင်း',
                'name_ja' => 'タイ',
                'name_ko' => '태국',
                'name_tr' => 'Tayland',
                'name_sv' => 'Thailand',
                'sort' => 1,
            ],
            [
                'name_zh' => '中国',
                'name_en' => 'China',
                'name_th' => 'จีน',
                'name_zh_tw' => '中國',
                'name_my' => 'တရုတ်',
                'name_ja' => '中国',
                'name_ko' => '중국',
                'name_tr' => 'Çin',
                'name_sv' => 'Kina',
                'sort' => 2,
            ],
            [
                'name_zh' => '美国',
                'name_en' => 'United States',
                'name_th' => 'สหรัฐอเมริกา',
                'name_zh_tw' => '美國',
                'name_my' => 'အမေရိကန်',
                'name_ja' => 'アメリカ',
                'name_ko' => '미국',
                'name_tr' => 'Amerika Birleşik Devletleri',
                'name_sv' => 'USA',
                'sort' => 3,
            ],
            [
                'name_zh' => '日本',
                'name_en' => 'Japan',
                'name_th' => 'ญี่ปุ่น',
                'name_zh_tw' => '日本',
                'name_my' => 'ဂျပန်',
                'name_ja' => '日本',
                'name_ko' => '일본',
                'name_tr' => 'Japonya',
                'name_sv' => 'Japan',
                'sort' => 4,
            ],
            [
                'name_zh' => '韩国',
                'name_en' => 'South Korea',
                'name_th' => 'เกาหลีใต้',
                'name_zh_tw' => '韓國',
                'name_my' => 'တောင်ကိုရီးယား',
                'name_ja' => '韓国',
                'name_ko' => '한국',
                'name_tr' => 'Güney Kore',
                'name_sv' => 'Sydkorea',
                'sort' => 5,
            ],
            [
                'name_zh' => '英国',
                'name_en' => 'United Kingdom',
                'name_th' => 'สหราชอาณาจักร',
                'name_zh_tw' => '英國',
                'name_my' => 'ယူနိုက်တက်ကင်းဒမ်း',
                'name_ja' => 'イギリス',
                'name_ko' => '영국',
                'name_tr' => 'Birleşik Krallık',
                'name_sv' => 'Storbritannien',
                'sort' => 6,
            ],
            [
                'name_zh' => '法国',
                'name_en' => 'France',
                'name_th' => 'ฝรั่งเศส',
                'name_zh_tw' => '法國',
                'name_my' => 'ပြင်သစ်',
                'name_ja' => 'フランス',
                'name_ko' => '프랑스',
                'name_tr' => 'Fransa',
                'name_sv' => 'Frankrike',
                'sort' => 7,
            ],
            [
                'name_zh' => '俄罗斯',
                'name_en' => 'Russia',
                'name_th' => 'รัสเซีย',
                'name_zh_tw' => '俄羅斯',
                'name_my' => 'ရုရှား',
                'name_ja' => 'ロシア',
                'name_ko' => '러시아',
                'name_tr' => 'Rusya',
                'name_sv' => 'Ryssland',
                'sort' => 8,
            ],
        ];

        foreach ($nationalities as $nationality) {
            // 检查是否已存在
            $exists = $this->fetchRow("SELECT id FROM ttpos_multi_language_name WHERE zh_name = '{$nationality['name_zh']}' AND en_name = '{$nationality['name_en']}' AND delete_time = 0");
            
            if (!$exists) {
                // 生成 UUID
                $multiLangUuid = createUuid();
                $nationalityUuid = createUuid();
                $currentTime = time();
                
                // 插入多语言名称
                $this->execute("INSERT INTO ttpos_multi_language_name (uuid, zh_name, en_name, zh_tw_name, th_name, my_name, ja_name, ko_name, tr_name, sv_name, create_time, update_time, delete_time) VALUES ({$multiLangUuid}, '{$nationality['name_zh']}', '{$nationality['name_en']}', '{$nationality['name_zh_tw']}', '{$nationality['name_th']}', '{$nationality['name_my']}', '{$nationality['name_ja']}', '{$nationality['name_ko']}', '{$nationality['name_tr']}', '{$nationality['name_sv']}', {$currentTime}, {$currentTime}, 0)");
                
                // 插入国籍
                $this->execute("INSERT INTO ttpos_nationality (uuid, multi_language_name_uuid, sort, status, create_time, update_time, delete_time) VALUES ({$nationalityUuid}, {$multiLangUuid}, {$nationality['sort']}, 1, {$currentTime}, {$currentTime}, 0)");
            }
        }
    }

}

