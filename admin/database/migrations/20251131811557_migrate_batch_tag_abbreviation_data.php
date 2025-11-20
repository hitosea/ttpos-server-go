<?php

use think\migration\Migrator;
use think\migration\db\Column;

class MigrateBatchTagAbbreviationData extends Migrator
{
    /**
     * Change Method.
     *
     * 为现有的分批类型设置默认缩写
     * 优先使用英文名称（en_name），如果没有英文名称则使用中文名称（zh_name），如果都没有则使用名称的前几个字符
     */
    public function change()
    {
        // 获取所有分批类型及其多语言名称
        $sql = "
            UPDATE ttpos_batch_tag bt
            INNER JOIN ttpos_multi_language_name mln ON bt.multi_language_name_uuid = mln.uuid
            SET bt.abbreviation = CASE
                WHEN mln.en_name != '' AND LENGTH(mln.en_name) <= 255 THEN mln.en_name
                WHEN mln.zh_name != '' AND LENGTH(mln.zh_name) <= 255 THEN mln.zh_name
                WHEN mln.en_name != '' THEN LEFT(mln.en_name, 255)
                WHEN mln.zh_name != '' THEN LEFT(mln.zh_name, 255)
                WHEN bt.name != '' THEN LEFT(bt.name, 255)
                ELSE ''
            END
            WHERE bt.abbreviation = '' OR bt.abbreviation IS NULL
        ";

        $this->execute($sql);
    }
}

