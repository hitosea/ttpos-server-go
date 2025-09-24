<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddWarehouseUuidToMaterialTable extends Migrator
{
   
    public function change()
    {
        $table = $this->table('material');
        if (!$table->hasColumn('warehouse_uuid')) {
            $table->addColumn('warehouse_uuid',  'biginteger', ['signed' => false, 'default' => 0, 'comment' => '默认仓库Uuid，表示该原料的来自哪个仓库', 'after' => 'headquarter_uuid'])
                  ->update();
        }

        $table = $this->table('warehouse_out_form_item');
        if (!$table->hasColumn('warehouse_uuid')) {
            $table->addColumn('warehouse_uuid',  'biginteger', ['signed' => false, 'default' => 0, 'comment' => '仓库uuid，出库的仓库', 'after' => 'warehouse_out_form_uuid'])
                  ->update();
        }
    }
}
