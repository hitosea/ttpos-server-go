<?php

use think\migration\Migrator;

class AlterMultiLanguageNameFieldsToVarchar1000 extends Migrator
{
    /**
     * 将 multi_language_name 表的所有语言字段从 VARCHAR(255) 改为 VARCHAR(1000)
     * 修复 Bug: bug-251128-001 - 编写卖点内容过长，保存报错
     */
    public function change()
    {
        $table = $this->table('multi_language_name');

        // 修改所有语言字段长度为 1000
        $fields = [
            'en_name'     => '英文名称',
            'zh_name'     => '中文名称',
            'zh_tw_name'  => '繁体中文名称',
            'th_name'     => '泰语名称',
            'my_name'     => '缅甸语名称',
            'ja_name'     => '日语名称',
            'ko_name'     => '韩语名称',
            'tr_name'     => '土耳其语名称',
            'sv_name'     => '瑞典语名称',
        ];

        foreach ($fields as $fieldName => $comment) {
            if ($table->hasColumn($fieldName)) {
                $table->changeColumn($fieldName, 'string', [
                    'limit'   => 1000,
                    'null'    => false,
                    'default' => '',
                    'comment' => $comment,
                ]);
            }
        }

        $table->update();
    }
}

