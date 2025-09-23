<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddHeadquarterUuidToMaterialTable extends Migrator
{
   
    public function change()
    {
        $table = $this->table('material');

        // 检查字段是否已存在
        if (!$table->hasColumn('headquarter_uuid')) {
            $table->addColumn('headquarter_uuid',  'biginteger', ['signed' => false, 'default' => 0, 'comment' => '总部Uuid', 'after' => 'actual_sale_num'])
                  ->update();
        }
    }
}
