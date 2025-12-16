<?php

use think\facade\Db;
use think\migration\Migrator;

class AddIsDisableToCompanyStaff extends Migrator
{
    // 迁移目标
    const TARGET = 'main';

    /**
     * Change Method.
     *
     * 为 saas.ttpos_company_staff 表添加 is_disable 字段，并从商家数据库同步数据
     */
    public function change()
    {
        $table = $this->table('company_staff'); 
        // 添加 sellable_quantity 字段
        if (!$table->hasColumn('is_disable')) {
           $table->addColumn('is_disable', 'integer', [ 
               'null' => false,
               'default' => 0,
               'comment' => '是否禁用1禁用,0未禁用',
               'after' => 'is_super'
           ])->update();
       }

        $saasDb = Db::connect(Db::getConfig('default'), true);
        // 从商家数据库同步 is_disable 数据
        $this->syncIsDisableFromShopDatabase($saasDb);
    }

    /**
     * 从商家数据库同步 is_disable 数据
     */
    private function syncIsDisableFromShopDatabase($saasDb)
    {
        // 获取所有门店的 company_staff 记录
        $companyStaffList = $saasDb->table('ttpos_company_staff')
            ->where('delete_time', 0)
            ->select()
            ->toArray();

        if (empty($companyStaffList)) {
            echo "没有需要同步的数据\n";
            return;
        }

        // 按门店分组
        $companyStaffByCompany = [];
        foreach ($companyStaffList as $cs) {
            $companyUuid = $cs['company_uuid'];
            if (!isset($companyStaffByCompany[$companyUuid])) {
                $companyStaffByCompany[$companyUuid] = [];
            }
            $companyStaffByCompany[$companyUuid][] = $cs;
        }

        $syncedCount = 0;
        $skippedCount = 0;

        // 遍历每个门店，从对应的商家数据库同步数据
        foreach ($companyStaffByCompany as $companyUuid => $staffList) {
            try {
                // 获取门店数据库连接
                $shopDb = $this->getShopDbConnection($companyUuid);
                if (!$shopDb) {
                    echo "门店 {$companyUuid} 的数据库连接失败，跳过\n";
                    $skippedCount += count($staffList);
                    continue;
                }

                // 查询该门店下所有员工的 is_disable 状态
                $stmt = $shopDb->prepare("SELECT uuid, is_disable FROM ttpos_staff WHERE delete_time = 0");
                $stmt->execute();
                $shopStaffList = [];
                while ($row = $stmt->fetch(\PDO::FETCH_ASSOC)) {
                    $shopStaffList[$row['uuid']] = (int)$row['is_disable'];
                }

                // 更新 company_staff 表的 is_disable 字段
                foreach ($staffList as $cs) {
                    $staffUuid = $cs['uuid'];
                    $isDisable = isset($shopStaffList[$staffUuid]) ? (int)$shopStaffList[$staffUuid] : 0;

                    $saasDb->table('ttpos_company_staff')
                        ->where('uuid', $staffUuid)
                        ->where('company_uuid', $companyUuid)
                        ->update(['is_disable' => $isDisable]);

                    $syncedCount++;
                }
            } catch (\Exception $e) {
                echo "同步门店 {$companyUuid} 的数据失败: " . $e->getMessage() . "\n";
                $skippedCount += count($staffList);
            }
        }

        echo "同步完成: 成功 {$syncedCount} 条，跳过 {$skippedCount} 条\n";
    }

    /**
     * 获取门店数据库连接
     */
    private function getShopDbConnection($companyUuid)
    {
        try {
            // 门店数据库连接格式：shop{company_uuid}
            $dbName = 'shop' . $companyUuid;

            // 获取数据库配置
            $config = \think\facade\Config::get('database.connections.mysql');

            // 构建 PDO DSN
            $dsn = sprintf(
                'mysql:host=%s;port=%s;dbname=%s;charset=%s',
                $config['hostname'] ?? '127.0.0.1',
                $config['hostport'] ?? 3306,
                $dbName,
                $config['charset'] ?? 'utf8mb4'
            );

            // 创建 PDO 连接
            $pdo = new \PDO(
                $dsn,
                $config['username'] ?? 'root',
                $config['password'] ?? '',
                [
                    \PDO::ATTR_ERRMODE => \PDO::ERRMODE_EXCEPTION,
                    \PDO::ATTR_DEFAULT_FETCH_MODE => \PDO::FETCH_ASSOC,
                    \PDO::ATTR_EMULATE_PREPARES => false,
                ]
            );

            return $pdo;
        } catch (\Exception $e) {
            return null;
        }
    }
}
