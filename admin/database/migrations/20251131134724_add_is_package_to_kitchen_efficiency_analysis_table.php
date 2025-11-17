<?php

use think\migration\Migrator;

class AddIsPackageToKitchenEfficiencyAnalysisTable extends Migrator
{
    /**
     * 添加is_package字段到后厨效率分析表
     */
    public function change()
    {
        $table = $this->table('kitchen_efficiency_analysis');

        // 添加is_package字段
        if (!$table->hasColumn('is_package')) {
            $table->addColumn('is_package', 'integer', ['null' => false, 'default' => 0, 'comment' => '是否是套餐: 0-否, 1-是', 'after' => 'product_package_uuid'])
                ->update();
        }
    }
}
