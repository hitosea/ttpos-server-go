<?php

use think\migration\Migrator;
use think\migration\db\Column;

class ModifyTmpDataFieldTypeInPrinterTemplateTable extends Migrator
{
    /**
     * 修改 ttpos_printer_template 表的 tmp_data 字段类型为 longtext
     */
    public function change()
    {
        $table = $this->table('printer_template');
        
        // 检查 tmp_data 字段是否存在
        if ($table->hasColumn('tmp_data')) {
            // 修改 tmp_data 字段类型为 longtext
            $table->changeColumn('tmp_data', 'text', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::TEXT_LONG, 'comment' => '临时模板数据', 'null' => true])
                  ->update();
        }
    }
}
