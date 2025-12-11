<?php

use think\migration\Migrator;
use think\facade\Db;
use think\facade\Config;

class MigrateExistingUsersToSaasStaff extends Migrator
{
    const TARGET = 'main';
    /**
     * 将现有用户数据迁移到 saas 数据库的统一账号表
     */
    public function change()
    {
        // 检查目标表是否存在
        $saasDb = Db::connect(Db::getConfig('saas'), true);
        $tables = $saasDb->query("SHOW TABLES LIKE 'ttpos_staff'");
        if (empty($tables)) {
            throw new \Exception('目标表 ttpos_staff 不存在，请先执行创建表的迁移');
        }

        // 检查是否已经迁移过（如果 saas.ttpos_staff 表已有数据，跳过迁移）
        $existingCount = $saasDb->name('staff')->where('delete_time', 0)->count();
        if ($existingCount > 0) {
            echo "数据已迁移，跳过本次迁移\n";
            return;
        }

        // 1. 从 saas.ttpos_company_staff 获取所有员工基础信息
        $companyStaffList = $saasDb->name('company_staff')
            ->where('delete_time', 0)
            ->order('create_time', 'asc')
            ->select()
            ->toArray();

        // 转换为数组格式（如果返回的是对象）
        if (!empty($companyStaffList) && is_object($companyStaffList[0])) {
            $companyStaffList = array_map(function($item) {
                return (array)$item;
            }, $companyStaffList);
        }

        if (empty($companyStaffList)) {
            echo "没有需要迁移的数据\n";
            return;
        }

        // 2. 按 uuid 分组，获取每个员工关联的第一个门店（用于设置 last_company_uuid）
        $staffFirstCompany = [];
        foreach ($companyStaffList as $cs) {
            $uuid = $cs['uuid'];
            if (!isset($staffFirstCompany[$uuid])) {
                $staffFirstCompany[$uuid] = $cs['company_uuid'];
            }
        }

        // 3. 收集所有门店UUID（用于遍历门店数据库）
        $companyUuids = array_unique(array_column($companyStaffList, 'company_uuid'));

        // 4. 从各个门店数据库获取员工详细信息
        $staffDetailsMap = []; // uuid => staff details
        $processedUuids = []; // 已处理的 uuid（用于去重）

        foreach ($companyUuids as $companyUuid) {
            try {
                // 连接门店数据库（数据库名格式：shop{company_uuid}）
                $shopDbName = 'shop' . $companyUuid;
                $pdo = $this->getShopDbConnection($shopDbName);

                if (!$pdo) {
                    continue;
                }

                // 使用 PDO 查询该门店数据库的 ttpos_staff 表
                $sql = "SELECT uuid, real_name, password, password_change_count, password_change_time, is_disable 
                        FROM ttpos_staff 
                        WHERE delete_time = 0";
                $stmt = $pdo->prepare($sql);
                $stmt->execute();
                $shopStaffList = $stmt->fetchAll(\PDO::FETCH_ASSOC);

                foreach ($shopStaffList as $shopStaff) {
                    $uuid = $shopStaff['uuid'] ?? 0;
                    if ($uuid == 0) {
                        continue;
                    }
                    // 只取第一条记录（处理重复数据）
                    if (!isset($processedUuids[$uuid])) {
                        $staffDetailsMap[$uuid] = [
                            'real_name' => $shopStaff['real_name'] ?? '',
                            'password' => $shopStaff['password'] ?? '',
                            'password_change_count' => $shopStaff['password_change_count'] ?? 0,
                            'password_change_time' => $shopStaff['password_change_time'] ?? 0,
                            'is_disable' => $shopStaff['is_disable'] ?? 0,
                        ];
                        $processedUuids[$uuid] = true;
                    }
                }
            } catch (\Exception $e) {
                // 如果门店数据库不存在或连接失败，跳过
                echo "警告：门店 {$companyUuid} 数据库连接失败: " . $e->getMessage() . "\n";
                continue;
            }
        }

        // 5. 按 uuid 分组 company_staff，处理邮箱重复的情况
        $staffCompanyMap = []; // uuid => company_staff records
        foreach ($companyStaffList as $cs) {
            $uuid = $cs['uuid'];
            if (!isset($staffCompanyMap[$uuid])) {
                $staffCompanyMap[$uuid] = [];
            }
            $staffCompanyMap[$uuid][] = $cs;
        }

        // 6. 构建插入数据
        $insertData = [];
        $emailMap = []; // 用于处理邮箱重复
        $currentTime = time();

        foreach ($staffCompanyMap as $uuid => $companyStaffRecords) {
            // 取第一条记录作为基础数据
            $baseRecord = $companyStaffRecords[0];
            
            // 转换为数组格式（如果返回的是对象）
            if (is_object($baseRecord)) {
                $baseRecord = (array)$baseRecord;
            }

            // 获取邮箱（来自 username 字段）
            $email = trim($baseRecord['username'] ?? '');
            if (empty($email)) {
                echo "警告:员工 {$uuid} 的 username 为空，跳过\n";
                continue;
            }

            // 处理邮箱重复：如果邮箱已存在，添加后缀
            $originalEmail = $email;
            $emailSuffix = 1;
            while (isset($emailMap[$email])) {
                $email = $originalEmail . '_' . $emailSuffix;
                $emailSuffix++;
            }
            $emailMap[$email] = true;

            // 获取手机号
            $phone = trim($baseRecord['phone'] ?? '');
            // 处理手机号空字符串的情况（保持为空字符串）

            // 获取门店数据库中的详细信息
            $staffDetails = $staffDetailsMap[$uuid] ?? [];

            // 获取 last_company_uuid（第一个门店）
            $lastCompanyUuid = $staffFirstCompany[$uuid] ?? 0;

            $insertData[] = [
                'uuid' => $uuid,
                'email' => $email,
                'phone' => $phone,
                'real_name' => $staffDetails['real_name'] ?? '',
                'password' => $staffDetails['password'] ?? '',
                'password_change_count' => $staffDetails['password_change_count'] ?? 0,
                'password_change_time' => $staffDetails['password_change_time'] ?? 0,
                'is_disable' => $staffDetails['is_disable'] ?? 0,
                'last_company_uuid' => $lastCompanyUuid,
                'create_time' => $baseRecord['create_time'] ?? $currentTime,
                'update_time' => $baseRecord['update_time'] ?? $currentTime,
                'delete_time' => $baseRecord['delete_time'] ?? 0,
            ];
        }

        // 7. 批量插入数据
        if (!empty($insertData)) {
            // 分批插入，每批 100 条
            $batchSize = 100;
            $batches = array_chunk($insertData, $batchSize);

            foreach ($batches as $batch) {
                try {
                    $saasDb->name('staff')->insertAll($batch);
                    echo "已插入 " . count($batch) . " 条记录\n";
                } catch (\Exception $e) {
                    echo "插入数据失败: " . $e->getMessage() . "\n";
                    // 如果批量插入失败，尝试逐条插入
                    foreach ($batch as $item) {
                        try {
                            $saasDb->name('staff')->insert($item);
                        } catch (\Exception $e2) {
                            echo "插入员工 {$item['uuid']} 失败: " . $e2->getMessage() . "\n";
                        }
                    }
                }
            }

            echo "数据迁移完成，共迁移 " . count($insertData) . " 条记录\n";
        } else {
            echo "没有需要迁移的数据\n";
        }
    }

    /**
     * 获取门店数据库连接（使用 PDO）
     */
    private function getShopDbConnection($dbName)
    {
        try {
            // 获取数据库配置
            $config = Config::get('database.connections.mysql');
            
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
