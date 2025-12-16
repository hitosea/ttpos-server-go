<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class DropUuidUniqueIndexFromCompanyStaff extends Migrator
{
    // 迁移目标
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        // 连接 saas 数据库
        // 注意：saas 数据库就是默认数据库（database 配置为 'saas'）
        // 如果 saas 连接配置存在，使用 saas；否则使用 default
        try {
            $saasConfig = Db::getConfig('saas');
            if ($saasConfig) {
                $saasDb = Db::connect($saasConfig, true);
            } else {
                $saasDb = Db::connect(Db::getConfig('default'), true);
            }
        } catch (\Exception $e) {
            // 如果 saas 配置不存在，使用默认连接
            $saasDb = Db::connect(Db::getConfig('default'), true);
        }
        
        // 检查表是否存在
        $tableExists = $saasDb->query("SHOW TABLES LIKE 'ttpos_company_staff'");
        if (empty($tableExists)) {
            echo "表 ttpos_company_staff 不存在，跳过迁移\n";
            return;
        }
        
        // 检查索引是否存在
        $indexExists = $saasDb->query("SHOW INDEX FROM ttpos_company_staff WHERE Key_name = 'unique_uuid'");
        if (empty($indexExists)) {
            echo "索引 unique_uuid 不存在，跳过迁移\n";
            return;
        }
        
        // 删除 uuid 的唯一索引
        try {
            $saasDb->execute("ALTER TABLE `ttpos_company_staff` DROP INDEX `unique_uuid`");
            echo "成功删除 ttpos_company_staff 表的 uuid 唯一索引\n";
            
            // 添加普通索引（用于查询优化）
            $saasDb->execute("ALTER TABLE `ttpos_company_staff` ADD INDEX `idx_uuid` (`uuid`)");
            echo "成功添加 ttpos_company_staff 表的 uuid 普通索引\n";
        } catch (\Exception $e) {
            echo "删除索引失败: " . $e->getMessage() . "\n";
            throw $e;
        }
    }
}
