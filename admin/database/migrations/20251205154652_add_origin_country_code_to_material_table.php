<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddOriginCountryCodeToMaterialTable extends Migrator
{
    /**
     * 添加 origin_country_code 字段到 ttpos_material 表
     */
    public function change()
    {
        $table = $this->table('material');

        // 检查字段是否已存在
        if (!$table->hasColumn('origin_country_code')) {
            $table->addColumn('origin_country_code', 'string', [
                'limit' => 10,
                'null' => false,
                'default' => '',
                'comment' => '原产地国家编码（ISO 3166-1 alpha-2，如：CN, US, TH）',
                'after' => 'allow_substore_visible'
            ])->update();
        }
    }
}

