<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class InitProductCategory extends Migrator
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
        $category = $db->name('product_category')->where('category_key', 'all')->find();
        if (!$category) {
            $nameUuid = createUuid();
            $db->name('multi_language_name')->insert([
                'uuid' => $nameUuid,
                'en_name' => 'All',
                'zh_name' => '全部',
                'zh_tw_name' => '全部',
                'th_name' => 'ทั้งหมด',
                'my_name' => 'အားလုံး',
                'ja_name' => 'すべて',
                'ko_name' => '전부',
                'tr_name' => 'Tümü',
                'create_time' => time(),
                'update_time' => time(),
            ]);
            $db->name('product_category')->insert([
                'uuid' => createUuid(),
                'name' => json_encode([
                    'th' => 'Tümü',
                    'zh' => '全部',
                    'zhtw' => '全部',
                    'en' => 'All',
                    'ja' => 'すべて',
                    'ko' => '전부',
                    'my' => 'အားလုံး',
                    'tr' => 'ทั้งหมด',
                ], JSON_UNESCAPED_UNICODE),
                'multi_language_name_uuid' => $nameUuid,
                'status' => 1,
                'parent_uuid' => 0,
                'category_key' => 'all',
                'create_time' => time(),
                'update_time' => time(),
            ]);
        }
    }
}
