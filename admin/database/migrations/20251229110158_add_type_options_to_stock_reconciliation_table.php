<?php
declare(strict_types=1);

use think\migration\Migrator;

/**
 * 盘点单类型字段扩展迁移
 * 
 * 为 ttpos_stock_reconciliation 表的 type 字段新增日盘、周盘、月盘类型
 * 
 * 变更内容：
 * - 修改 type 字段注释，新增类型说明：3-日盘、4-周盘、5-月盘
 * - 保持字段类型、默认值、现有数据不变
 * 
 * 类型定义：
 * 1 - 指定物品盘点
 * 2 - 全部物品盘点
 * 3 - 日盘（新增）
 * 4 - 周盘（新增）
 * 5 - 月盘（新增）
 */
final class AddTypeOptionsToStockReconciliationTable extends Migrator
{
    /**
     * 迁移变更
     */
    public function change(): void
    {
        $table = $this->table('stock_reconciliation');
        
        // 检查 type 字段是否存在
        if ($table->hasColumn('type')) {
            // 仅修改字段注释，不修改字段类型、默认值和现有数据
            $table->changeColumn('type', 'integer', [
                'null' => false,
                'default' => 1,
                'comment' => '盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘',
                'signed' => true,
            ])->update();
        }
    }
}

