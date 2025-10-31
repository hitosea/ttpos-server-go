<?php

use think\facade\Db;
use think\migration\Migrator;

class UpdateSupplierCode extends Migrator
{
    /**
     * Change Method.
     *
     * 更新 ttpos_supplier 表中 erp_code='Headquarters - Supplier' 的数据的 code 字段为 HSP00001
     */
    public function change()
    {
        // 获取数据库连接
        $db = Db::connect(Db::getConfig('default'), true);
        // 更新 code 字段为 HSP00001
        $db->name("supplier")
            ->where("erp_code", "=", "Headquarters - Supplier")
            ->update(["code" => "HSP00001"]);
    }
}
