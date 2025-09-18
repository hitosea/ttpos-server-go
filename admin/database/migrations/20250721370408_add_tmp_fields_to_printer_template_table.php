<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddTmpFieldsToPrinterTemplateTable extends Migrator
{
    /**
     * 添加 tmp_uuid 和 tmp_data 字段到 ttpos_printer_template 表
     */
    public function change()
    {
        $table = $this->table('printer_template');

        // 检查 tmp_uuid 字段是否已存在
        if (!$table->hasColumn('tmp_uuid')) {
            $table->addColumn('tmp_uuid', 'biginteger', ['default' => 0, 'comment' => '临时模板UUID', 'after' => 'is_show_sku'])->update();
        }

        // 检查 tmp_data 字段是否已存在
        if (!$table->hasColumn('tmp_data')) {
            $table->addColumn('tmp_data', 'text', ['comment' => '临时模板数据', 'after' => 'tmp_uuid'])->update();
        }
    }
}
