<?php

use think\migration\Migrator;

/**
 * 创建物品分类可见性配置表
 *
 * 用于配置物品类别对门店角色的可见性规则
 * - 总店创建配置，同步到各子店
 * - 支持全部门店角色或部分门店角色
 * - 白名单模式：仅显示配置中的物品类别
 */
class CreateMaterialCategoryVisibility extends Migrator
{
    // 迁移目标：所有商户数据库和 saas 主库
    const TARGET = 'all';

    /**
     * Change Method.
     */
    public function change()
    {
        // 检查表是否已存在
        if ($this->hasTable('material_category_visibility')) {
            return;
        }

        // 创建物品分类可见性配置表
        $table = $this->table('material_category_visibility', ['comment' => '物品分类可见性配置表']);
        $table->addColumn('uuid', 'biginteger', ['default' => 0, 'comment' => '唯一ID'])
            ->addColumn('name', 'text', ['null' => true, 'comment' => '配置名称（多语言JSON）'])
            ->addColumn('is_all_roles', 'integer', ['limit' => 1, 'default' => 0, 'comment' => '是否全部门店角色：1-全部，0-部分'])
            ->addColumn('category_uuids', 'text', ['null' => true, 'comment' => '物品类别UUID列表，JSON数组格式'])
            ->addColumn('role_configs', 'text', ['null' => true, 'comment' => '角色配置，JSON数组格式 [{role_uuid, company_uuid}]'])
            ->addColumn('headquarter_uuid', 'biginteger', ['default' => 0, 'comment' => '总部UUID，0=当前店创建，>0=来自总部'])
            ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
            ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
            ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
            ->addIndex(['uuid'], ['unique' => true])
            ->addIndex(['headquarter_uuid'])
            ->create();
    }
}
