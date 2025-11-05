<?php

use think\migration\Migrator;
use Phinx\Db\Adapter\MysqlAdapter;

class ChangeTransferOrderWarehouseNameToText extends Migrator
{
    /**
     * 修改 ttpos_transfer_order 表的 out_warehouse_name 和 in_warehouse_name 字段类型为 text
     */
    public function change()
    {
        $table = $this->table('transfer_order');
        
        // 检查表是否存在
        if (!$table->exists()) {
            return;
        }
        
        // 检查字段是否存在，如果存在则修改类型
        if ($table->hasColumn('out_warehouse_name')) {
            $table->changeColumn('out_warehouse_name', 'text', ['comment' => '出库仓库名称'])->update();
        }
        
        if ($table->hasColumn('in_warehouse_name')) {
            $table->changeColumn('in_warehouse_name', 'text', ['comment' => '入库仓库名称'])->update();
        }
    }
}

