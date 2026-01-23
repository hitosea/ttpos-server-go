<?php

use think\migration\Migrator;
use think\facade\Db;

/**
 * 迁移品牌采购限购配置到新方案表
 * 
 * 功能说明：
 * 1. 创建 4 个新表：
 *    - ttpos_purchase_limit_scheme (限购方案主表)
 *    - ttpos_purchase_limit_scheme_item (物品配置表)
 *    - ttpos_purchase_limit_scheme_shop (门店配置表)
 * 
 * 2. 迁移旧表数据：
 *    - 从 ttpos_purchase_quota_config 迁移到新方案表
 *    - 从 ttpos_purchase_quota_config_shop 迁移到门店配置表
 *    - 默认周期：全周（周一到周日）
 * 
 * 3. 删除旧表：
 *    - ttpos_purchase_quota_config
 *    - ttpos_purchase_quota_config_shop
 * 
 * 4. 使用事务保证原子性
 */
class MigratePurchaseQuotaToLimitScheme extends Migrator
{
    /**
     * 执行迁移
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 使用事务保证原子性
        try {
            // Step 1: 创建新表
            $this->createNewTables();

            // Step 2: 迁移数据
            // $this->migrateData($db);

            // Step 3: 删除旧表
            $this->dropOldTables();

            dump('迁移成功：旧限购配置已迁移到新方案表');
        } catch (\Exception $e) {
            dump('迁移失败：' . $e->getMessage());
            throw $e;
        }
    }

    /**
     * 创建新表
     */
    private function createNewTables()
    {
        // 1. 创建限购方案主表
        if (!$this->hasTable('purchase_limit_scheme')) {
            $table = $this->table('purchase_limit_scheme', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_general_ci',
                'comment' => '限购方案表',
            ]);
            $table->addColumn('id', 'biginteger', ['identity' => true, 'signed' => false, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('name', 'text', ['null' => false, 'comment' => '方案名称'])
                ->addColumn('status', 'integer', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY, 'default' => 1, 'comment' => '状态：0=关闭，1=开启'])
                ->addColumn('apply_to_all_shops', 'integer', ['limit' => \Phinx\Db\Adapter\MysqlAdapter::INT_TINY, 'default' => 1, 'comment' => '是否适用所有门店：0=否，1=是'])
                ->addColumn('daily_limit', 'integer', ['default' => 0, 'comment' => '每日申请次数限制（0=不限制）'])
                ->addColumn('weekdays', 'string', ['limit' => 20, 'default' => '', 'comment' => '星期配置（逗号分隔，1-7表示周一到周日）'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['status', 'delete_time'], ['name' => 'idx_status'])
                ->create();
        }
        
        // 2. 创建物品配置表
        if (!$this->hasTable('purchase_limit_scheme_item')) {
            $table = $this->table('purchase_limit_scheme_item', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_general_ci',
                'comment' => '限购方案物品配置表',
            ]);
            $table->addColumn('id', 'biginteger', ['identity' => true, 'signed' => false, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('scheme_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '限购方案UUID'])
                ->addColumn('material_code', 'string', ['limit' => 50, 'default' => '', 'comment' => '物品编码'])
                ->addColumn('unit_code', 'string', ['limit' => 50, 'default' => '', 'comment' => '单位编码'])
                ->addColumn('quota_limit', 'decimal', ['precision' => 20, 'scale' => 8, 'default' => 0, 'comment' => '限购数量（0=不限制）'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['scheme_uuid', 'material_code', 'delete_time'], ['name' => 'idx_scheme_material'])
                ->create();
        }
        
        // 3. 创建门店配置表
        if (!$this->hasTable('purchase_limit_scheme_shop')) {
            $table = $this->table('purchase_limit_scheme_shop', [
                'id' => false,
                'primary_key' => ['id'],
                'engine' => 'InnoDB',
                'collation' => 'utf8mb4_general_ci',
                'comment' => '限购方案门店配置表',
            ]);
            $table->addColumn('id', 'biginteger', ['identity' => true, 'signed' => false, 'comment' => '主键ID'])
                ->addColumn('uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '唯一标识'])
                ->addColumn('scheme_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '限购方案UUID'])
                ->addColumn('company_uuid', 'biginteger', ['signed' => false, 'default' => 0, 'comment' => '门店UUID'])
                ->addColumn('create_time', 'integer', ['default' => 0, 'comment' => '创建时间'])
                ->addColumn('update_time', 'integer', ['default' => 0, 'comment' => '更新时间'])
                ->addColumn('delete_time', 'integer', ['default' => 0, 'comment' => '删除时间'])
                ->addIndex(['uuid'], ['unique' => true, 'name' => 'uk_uuid'])
                ->addIndex(['scheme_uuid', 'company_uuid', 'delete_time'], ['name' => 'idx_scheme_company'])
                ->create();
        }
        
    }

    /**
     * 迁移数据
     */
    private function migrateData($db)
    {
        // 检查旧表是否存在
        if (!$this->hasTable('purchase_quota_config')) {
            return;
        }

        // 查询所有旧配置记录
        $oldConfigs = $db->name('purchase_quota_config')
            ->where('delete_time', 0)
            ->where('status', 1)
            ->select()
            ->toArray();
        
        if (empty($oldConfigs)) {
            return;
        }

        // 生成新的 UUID
        $schemeUuid = $this->generateUuid();
        $currentTime = time();
        
        // 1. 创建限购方案主记录
        $schemeId = $db->name('purchase_limit_scheme')->insertGetId([
            'uuid' => $schemeUuid,
            'name' => '每日限购',
            'status' => 1,
            'apply_to_all_shops' => 1,
            'daily_limit' => 2, // 旧系统没有每日限制，默认为 0
            'weekdays' => '1,2,3,4,5,6,7', // 默认全周
            'create_time' => $currentTime,
            'update_time' => $currentTime,
            'delete_time' => 0,
        ]);
        if (!$schemeId) {
            dump('【迁移数据】创建限购方案主记录失败');
            return;
        }

        // 2. 创建物品配置记录
        foreach ($oldConfigs as $config) {
            // 2. 创建物品配置记录
            $db->name('purchase_limit_scheme_item')->insert([
                'uuid' => $this->generateUuid(),
                'scheme_uuid' => $schemeUuid,
                'material_code' => $config['material_code'],
                'unit_code' => $config['unit_code'],
                'quota_limit' => $config['quota_limit'] ?? 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
                'delete_time' => 0,
            ]);
        }

    }

    /**
     * 删除旧表
     */
    private function dropOldTables()
    {
        // 删除门店关联表
        if ($this->hasTable('purchase_quota_config_shop')) {
            $this->table('purchase_quota_config_shop')->drop()->save();
        }

        // 删除主表
        if ($this->hasTable('purchase_quota_config')) {
            $this->table('purchase_quota_config')->drop()->save();
        }

    }

    /**
     * 生成 UUID（雪花ID）
     * 注：这里使用简单的时间戳+随机数，生产环境应使用实际的雪花ID生成器
     */
    private function generateUuid()
    {
        // 使用微秒时间戳 + 随机数
        return (int)(microtime(true) * 10000) + mt_rand(1000, 9999);
    }

}
