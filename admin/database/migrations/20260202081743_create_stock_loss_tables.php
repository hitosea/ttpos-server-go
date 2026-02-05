<?php

use think\migration\Migrator;

/**
 * 创建报损单相关表
 *
 * 包含以下表：
 * - stock_loss: 报损单主表
 * - stock_loss_item: 报损单明细表
 * - stock_loss_annotation: 报损单批注表
 * - stock_loss_file: 报损单附件关联表
 */
class CreateStockLossTables extends Migrator
{
    /**
     * Change Method.
     */
    public function change()
    {
        $this->createStockLossTable();
        $this->createStockLossItemTable();
        $this->createStockLossItemUnitTable();
        $this->createStockLossAnnotationTable();
        $this->createStockLossFileTable();
    }

    /**
     * 创建报损单主表
     */
    private function createStockLossTable()
    {
        if ($this->hasTable('stock_loss')) {
            return;
        }

        $table = $this->table('stock_loss', ['comment' => '报损单主表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('code', 'string', ['limit' => 50, 'default' => '', 'comment' => '单据编号 SL+时间戳+序列号'])
            ->addColumn('erp_code', 'string', ['limit' => 50, 'default' => '', 'comment' => 'ERP单据编号'])
            ->addColumn('loss_type', 'integer', ['limit' => 1, 'default' => 1, 'comment' => '报损类型 1:物品损坏 2:物品报废 3:物品过期'])
            ->addColumn('warehouse_uuid', 'biginteger', ['default' => 0, 'comment' => '报损仓库UUID'])
            ->addColumn('reason', 'text', ['null' => true, 'comment' => '报损原因'])
            ->addColumn('status', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '状态 0:已保存 1:已提交 2:已审核通过 3:已驳回'])
            ->addColumn('submit_time', 'integer', ['default' => 0, 'comment' => '提交时间'])
            ->addColumn('approve_time', 'integer', ['default' => 0, 'comment' => '审核通过时间'])
            ->addColumn('reject_time', 'integer', ['default' => 0, 'comment' => '驳回时间'])
            ->addColumn('submitter_uuid', 'biginteger', ['default' => 0, 'comment' => '提交人UUID'])
            ->addColumn('approver_uuid', 'biginteger', ['default' => 0, 'comment' => '审核人UUID'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['code'])
            ->addIndex(['erp_code'])
            ->addIndex(['warehouse_uuid'])
            ->addIndex(['status'])
            ->addIndex(['delete_time'])
            ->create();
    }

    /**
     * 创建报损单明细表
     */
    private function createStockLossItemTable()
    {
        if ($this->hasTable('stock_loss_item')) {
            return;
        }

        $table = $this->table('stock_loss_item', ['comment' => '报损单明细表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('stock_loss_uuid', 'biginteger', ['default' => 0, 'comment' => '报损单UUID'])
            ->addColumn('material_uuid', 'biginteger', ['default' => 0, 'comment' => '物料UUID'])
            ->addColumn('material_name', 'text', ['null' => true, 'comment' => '物料名称'])
            ->addColumn('base_quantity', 'decimal', ['precision' => 14,'scale' => 4,'default' => '0.0000','comment' => '基准单位数量（所有单位转换后的总量）','after' => 'material_unit_name'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['stock_loss_uuid'])
            ->addIndex(['material_uuid'])
            ->create();
    }

    /**
     * 创建报损单物品单位明细表
     */
    protected function createStockLossItemUnitTable()
    {
        if ($this->hasTable('stock_loss_item_unit')) {
            return;
        }

        $table = $this->table('stock_loss_item_unit', ['comment' => '报损单物品单位明细表']);

        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '报损单物品单位明细UUID'])
            ->addColumn('stock_loss_item_uuid', 'biginteger', ['default' => 0, 'comment' => '报损单物品明细UUID'])
            ->addColumn('material_unit_uuid', 'biginteger', ['default' => 0, 'comment' => '物料单位UUID'])
            ->addColumn('material_unit_name', 'text', ['null' => true, 'comment' => '物料单位名称（多语言JSON备份）'])
            ->addColumn('quantity', 'decimal', ['precision' => 22, 'scale' => 4, 'null' => true, 'comment' => '单位数量'])
            ->addColumn('create_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['signed' => false, 'default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true, 'name' => 'unique_uuid'])
            ->addIndex(['stock_loss_item_uuid'], ['name' => 'idx_stock_loss_item_uuid'])
            ->addIndex(['material_unit_uuid'], ['name' => 'idx_material_unit_uuid'])
            ->create();
    }

    /**
     * 创建报损单批注表
     */
    private function createStockLossAnnotationTable()
    {
        if ($this->hasTable('stock_loss_annotation')) {
            return;
        }

        $table = $this->table('stock_loss_annotation', ['comment' => '报损单批注表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('stock_loss_uuid', 'biginteger', ['default' => 0, 'comment' => '报损单UUID'])
            ->addColumn('action', 'string', ['limit' => 20, 'default' => '', 'comment' => '操作类型 submit/resubmit/approve/reject'])
            ->addColumn('content', 'text', ['null' => true, 'comment' => '批注内容'])
            ->addColumn('operator_uuid', 'biginteger', ['default' => 0, 'comment' => '操作人UUID'])
            ->addColumn('operator_name', 'string', ['limit' => 100, 'default' => '', 'comment' => '操作人姓名'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['stock_loss_uuid'])
            ->create();
    }

    /**
     * 创建报损单附件关联表
     */
    private function createStockLossFileTable()
    {
        if ($this->hasTable('stock_loss_file')) {
            return;
        }

        $table = $this->table('stock_loss_file', ['comment' => '报损单附件关联表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('stock_loss_uuid', 'biginteger', ['default' => 0, 'comment' => '报损单UUID'])
            ->addColumn('file_uuid', 'biginteger', ['default' => 0, 'comment' => '文件UUID'])
            ->addColumn('sort_order', 'integer', ['default' => 0, 'comment' => '排序顺序'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['stock_loss_uuid'])
            ->addIndex(['file_uuid'])
            ->create();
    }
}
