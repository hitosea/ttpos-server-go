<?php
/**
 * 补录旧数据的 hq_field_override 记录
 *
 * 背景：hq_field_override 表是后引入的功能，用于追踪子店对总部字段的覆盖。
 * 旧数据中，子店可能已修改了商品的上下架/价格，但没有对应的 override 记录。
 * 同步逻辑中"分开控制"模式下，无 override 记录 = 使用 HQ 值，会导致子店旧数据被覆盖。
 *
 * 方案：遍历每个子店数据库，跨库对比 HQ 值，差异处补录 override 记录。
 * 涉及字段：dine_shelf（店内上下架）、takeout_shelf（外卖上下架）、takeout_price（外卖价格）、
 *          safety_stock（安全库存）、negative_stock（负库存）
 */

use think\migration\Migrator;
use think\facade\Config;

class SeedHqFieldOverrideLegacyData extends Migrator
{
    // 无 TARGET：在每个商户数据库上执行

    /** @var array 待批量插入的 override 记录 [entity_type, entity_uuid, field_type] */
    private $pendingOverrides = [];

    public function change()
    {
        // 检查 hq_field_override 表是否存在
        if (!$this->hasTable('hq_field_override')) {
            throw new \RuntimeException('hq_field_override 表不存在，请先执行建表迁移');
        }

        // 从当前商户数据库获取 headquarter_uuid
        $companySetting = $this->fetchRow("SELECT headquarter_uuid FROM ttpos_company_setting WHERE delete_time = 0 LIMIT 1");
        if (empty($companySetting) || empty($companySetting['headquarter_uuid']) || $companySetting['headquarter_uuid'] == 0) {
            // 不是子店（没有总部），跳过
            return;
        }

        $headquarterUuid = $companySetting['headquarter_uuid'];
        $hqDbName = 'shop' . $headquarterUuid;

        // 连接总部数据库
        $hqPdo = $this->getShopDbConnection($hqDbName);

        // 1. dine_shelf：对比 ttpos_product_package.status
        $this->collectDineShelfOverrides($hqPdo, $hqDbName);

        // 2. takeout_shelf + takeout_price：对比 ttpos_product_package_takeout
        $this->collectTakeoutOverrides($hqPdo, $hqDbName);

        // 3. safety_stock + negative_stock：对比 ttpos_material
        $this->collectMaterialOverrides($hqPdo, $hqDbName);

        // 批量插入所有收集到的 override 记录
        $insertCount = $this->flushOverrides();

        if ($insertCount > 0) {
            echo "补录 {$insertCount} 条 override 记录 (HQ: {$hqDbName})\n";
        }
    }

    /**
     * 收集 dine_shelf override 记录
     * 对比子店和总部的 ttpos_product_package.status
     */
    private function collectDineShelfOverrides($hqPdo, $hqDbName)
    {
        $subProducts = $this->fetchAll(
            "SELECT uuid, status FROM ttpos_product_package WHERE headquarter_uuid > 0 AND delete_time = 0"
        );
        if (empty($subProducts)) {
            return;
        }

        $uuids = array_column($subProducts, 'uuid');
        $subMap = [];
        foreach ($subProducts as $row) {
            $subMap[$row['uuid']] = $row;
        }

        $existingOverrides = $this->getExistingOverrides($uuids, 'dine_shelf');

        $hqProducts = $this->fetchFromHq($hqPdo, $hqDbName, 'product_package', ['uuid', 'status'], $uuids);
        $hqMap = [];
        foreach ($hqProducts as $row) {
            $hqMap[$row['uuid']] = $row;
        }

        foreach ($subMap as $uuid => $subRow) {
            if (isset($existingOverrides[$uuid])) {
                continue;
            }
            if (!isset($hqMap[$uuid])) {
                continue;
            }
            if ($subRow['status'] == $hqMap[$uuid]['status']) {
                continue;
            }
            $this->pendingOverrides[] = ['product', $uuid, 'dine_shelf'];
        }
    }

    /**
     * 收集 takeout_shelf 和 takeout_price override 记录
     *
     * 注意：外卖商品在子店创建时会生成新的 UUID（不复用总部 UUID），
     * 匹配关系为 product_package_uuid + takeout_type
     */
    private function collectTakeoutOverrides($hqPdo, $hqDbName)
    {
        $subTakeouts = $this->fetchAll(
            "SELECT uuid, product_package_uuid, takeout_type, status, price FROM ttpos_product_package_takeout WHERE headquarter_uuid > 0 AND delete_time = 0"
        );
        if (empty($subTakeouts)) {
            return;
        }

        // 子店：按 product_package_uuid + takeout_type 建索引
        $subMap = [];
        foreach ($subTakeouts as $row) {
            $key = $row['product_package_uuid'] . '-' . $row['takeout_type'];
            $subMap[$key] = $row;
        }

        $subUuids = array_column($subTakeouts, 'uuid');
        $existingShelfOverrides = $this->getExistingOverrides($subUuids, 'takeout_shelf');
        $existingPriceOverrides = $this->getExistingOverrides($subUuids, 'takeout_price');

        $productPackageUuids = array_unique(array_column($subTakeouts, 'product_package_uuid'));

        $hqTakeouts = $this->fetchFromHq($hqPdo, $hqDbName, 'product_package_takeout',
            ['uuid', 'product_package_uuid', 'takeout_type', 'status', 'price'], $productPackageUuids,
            'product_package_uuid'
        );
        $hqMap = [];
        foreach ($hqTakeouts as $row) {
            $key = $row['product_package_uuid'] . '-' . $row['takeout_type'];
            $hqMap[$key] = $row;
        }

        $bomDiffSubUuids = $this->getBomDiffTakeoutUuids($hqPdo, $hqDbName, $subMap, $hqMap);

        foreach ($subMap as $key => $subRow) {
            if (!isset($hqMap[$key])) {
                continue;
            }
            $hqRow = $hqMap[$key];
            $subUuid = $subRow['uuid'];

            // takeout_shelf
            if (!isset($existingShelfOverrides[$subUuid]) && $subRow['status'] != $hqRow['status']) {
                $this->pendingOverrides[] = ['product_takeout', $subUuid, 'takeout_shelf'];
            }

            // takeout_price：主表价格或任一规格有差异
            if (!isset($existingPriceOverrides[$subUuid])) {
                if ($subRow['price'] != $hqRow['price'] || isset($bomDiffSubUuids[$subUuid])) {
                    $this->pendingOverrides[] = ['product_takeout', $subUuid, 'takeout_price'];
                }
            }
        }
    }

    /**
     * 对比子店和总部的外卖规格（bom_takeout），返回有差异的子店外卖商品 UUID 集合
     * 差异判定：规格数量不同、规格组合不同（product_bom_uuid 不一致）、或价格不同
     *
     * @param array $subMap  子店 map[pkg_uuid-takeout_type => row]
     * @param array $hqMap   总部 map[pkg_uuid-takeout_type => row]
     * @return array map[sub_takeout_uuid => true]
     */
    private function getBomDiffTakeoutUuids($hqPdo, $hqDbName, array $subMap, array $hqMap)
    {
        // 收集需要对比的配对（子店 takeout UUID => 总部 takeout UUID）
        $subTakeoutUuids = [];
        $hqTakeoutUuids = [];
        $subToHqMap = [];
        foreach ($subMap as $key => $subRow) {
            if (isset($hqMap[$key])) {
                $subTakeoutUuids[] = $subRow['uuid'];
                $hqTakeoutUuids[] = $hqMap[$key]['uuid'];
                $subToHqMap[$subRow['uuid']] = $hqMap[$key]['uuid'];
            }
        }

        if (empty($subTakeoutUuids)) {
            return [];
        }

        // 查询子店 bom_takeout，按子店 takeout UUID 分组
        $placeholders = implode(',', array_fill(0, count($subTakeoutUuids), '?'));
        $stmt = $this->query(
            "SELECT product_package_takeout_uuid, product_bom_uuid, price FROM ttpos_product_bom_takeout WHERE product_package_takeout_uuid IN ({$placeholders}) AND delete_time = 0",
            array_values($subTakeoutUuids)
        );
        $subBoms = $stmt ? $stmt->fetchAll() : [];

        $subGrouped = [];
        foreach ($subBoms as $row) {
            $subGrouped[$row['product_package_takeout_uuid']][$row['product_bom_uuid']] = $row['price'];
        }

        // 查询总部 bom_takeout，按总部 takeout UUID 分组
        $hqPlaceholders = implode(',', array_fill(0, count($hqTakeoutUuids), '?'));
        $hqStmt = $hqPdo->prepare(
            "SELECT product_package_takeout_uuid, product_bom_uuid, price FROM {$hqDbName}.ttpos_product_bom_takeout WHERE product_package_takeout_uuid IN ({$hqPlaceholders}) AND delete_time = 0"
        );
        $hqStmt->execute(array_values($hqTakeoutUuids));
        $hqBoms = $hqStmt->fetchAll(\PDO::FETCH_ASSOC);

        $hqGrouped = [];
        foreach ($hqBoms as $row) {
            $hqGrouped[$row['product_package_takeout_uuid']][$row['product_bom_uuid']] = $row['price'];
        }

        // 对比每对子店-总部的规格集合
        $diffSubUuids = [];
        foreach ($subToHqMap as $subUuid => $hqUuid) {
            $subBomSet = $subGrouped[$subUuid] ?? [];
            $hqBomSet = $hqGrouped[$hqUuid] ?? [];

            if (count($subBomSet) != count($hqBomSet)) {
                $diffSubUuids[$subUuid] = true;
                continue;
            }

            foreach ($hqBomSet as $bomUuid => $hqPrice) {
                if (!isset($subBomSet[$bomUuid]) || $subBomSet[$bomUuid] != $hqPrice) {
                    $diffSubUuids[$subUuid] = true;
                    break;
                }
            }
        }

        return $diffSubUuids;
    }

    /**
     * 收集 safety_stock 和 negative_stock override 记录
     * 对比子店和总部的 ttpos_material.safety_stock 和 allow_negative_stock
     */
    private function collectMaterialOverrides($hqPdo, $hqDbName)
    {
        // 检查 ttpos_material 表中可选字段是否存在（部分商户库可能缺少 safety_stock）
        $materialTable = $this->table('material');
        $hasSafetyStock = $materialTable->hasColumn('safety_stock');
        $hasNegativeStock = $materialTable->hasColumn('allow_negative_stock');

        if (!$hasSafetyStock && !$hasNegativeStock) {
            return;
        }

        $columns = ['uuid'];
        if ($hasSafetyStock) {
            $columns[] = 'safety_stock';
        }
        if ($hasNegativeStock) {
            $columns[] = 'allow_negative_stock';
        }

        $columnStr = implode(', ', $columns);
        $subMaterials = $this->fetchAll(
            "SELECT {$columnStr} FROM ttpos_material WHERE headquarter_uuid > 0 AND delete_time = 0"
        );
        if (empty($subMaterials)) {
            return;
        }

        $uuids = array_column($subMaterials, 'uuid');
        $subMap = [];
        foreach ($subMaterials as $row) {
            $subMap[$row['uuid']] = $row;
        }

        $existingSafetyOverrides = $hasSafetyStock ? $this->getExistingOverrides($uuids, 'safety_stock') : [];
        $existingNegativeOverrides = $hasNegativeStock ? $this->getExistingOverrides($uuids, 'negative_stock') : [];

        $hqMaterials = $this->fetchFromHq($hqPdo, $hqDbName, 'material', $columns, $uuids);
        $hqMap = [];
        foreach ($hqMaterials as $row) {
            $hqMap[$row['uuid']] = $row;
        }

        foreach ($subMap as $uuid => $subRow) {
            if (!isset($hqMap[$uuid])) {
                continue;
            }
            $hqRow = $hqMap[$uuid];

            if ($hasSafetyStock && !isset($existingSafetyOverrides[$uuid]) && $subRow['safety_stock'] != $hqRow['safety_stock']) {
                $this->pendingOverrides[] = ['material', $uuid, 'safety_stock'];
            }

            if ($hasNegativeStock && !isset($existingNegativeOverrides[$uuid]) && $subRow['allow_negative_stock'] != $hqRow['allow_negative_stock']) {
                $this->pendingOverrides[] = ['material', $uuid, 'negative_stock'];
            }
        }
    }

    /**
     * 批量插入所有收集到的 override 记录
     * 使用 INSERT IGNORE 配合唯一索引 uk_entity_field(entity_uuid, field_type, delete_time) 排重
     *
     * @return int 实际插入行数
     */
    private function flushOverrides()
    {
        if (empty($this->pendingOverrides)) {
            return 0;
        }

        $now = time();
        $batchSize = 500;
        $totalInserted = 0;

        foreach (array_chunk($this->pendingOverrides, $batchSize) as $chunk) {
            $valuePlaceholders = [];
            $params = [];
            foreach ($chunk as $row) {
                $uuid = createUuid();
                $valuePlaceholders[] = '(?, ?, ?, ?, ?, ?, 0)';
                $params[] = $uuid;
                $params[] = $row[0]; // entity_type
                $params[] = $row[1]; // entity_uuid
                $params[] = $row[2]; // field_type
                $params[] = $now;    // create_time
                $params[] = $now;    // update_time
            }

            $sql = "INSERT IGNORE INTO ttpos_hq_field_override (uuid, entity_type, entity_uuid, field_type, create_time, update_time, delete_time) VALUES "
                . implode(', ', $valuePlaceholders);

            $totalInserted += $this->execute($sql, $params);
        }

        $this->pendingOverrides = [];
        return $totalInserted;
    }

    /**
     * 查询已有的 override 记录，返回 map[entity_uuid] => true
     */
    private function getExistingOverrides(array $uuids, string $fieldType)
    {
        if (empty($uuids)) {
            return [];
        }

        $placeholders = implode(',', array_fill(0, count($uuids), '?'));
        $params = array_values($uuids);
        $params[] = $fieldType;
        $stmt = $this->query(
            "SELECT entity_uuid FROM ttpos_hq_field_override WHERE entity_uuid IN ({$placeholders}) AND field_type = ? AND delete_time = 0",
            $params
        );

        $map = [];
        if ($stmt) {
            foreach ($stmt->fetchAll() as $row) {
                $map[$row['entity_uuid']] = true;
            }
        }
        return $map;
    }

    /**
     * 从总部数据库查询数据
     *
     * @param string $matchColumn IN 匹配的字段名，默认 uuid
     */
    private function fetchFromHq($hqPdo, $hqDbName, $tableName, array $columns, array $values, $matchColumn = 'uuid')
    {
        if (empty($values)) {
            return [];
        }

        $columnStr = implode(', ', $columns);
        $placeholders = implode(',', array_fill(0, count($values), '?'));
        $sql = "SELECT {$columnStr} FROM {$hqDbName}.ttpos_{$tableName} WHERE {$matchColumn} IN ({$placeholders}) AND delete_time = 0";

        $stmt = $hqPdo->prepare($sql);
        $stmt->execute(array_values($values));
        return $stmt->fetchAll(\PDO::FETCH_ASSOC);
    }

    /**
     * 连接指定商户数据库
     *
     * @throws \PDOException
     */
    private function getShopDbConnection($dbName)
    {
        $config = Config::get('database.connections.mysql');

        $dsn = sprintf(
            'mysql:host=%s;port=%s;dbname=%s;charset=%s',
            $config['hostname'] ?? '127.0.0.1',
            $config['hostport'] ?? 3306,
            $dbName,
            $config['charset'] ?? 'utf8mb4'
        );

        return new \PDO(
            $dsn,
            $config['username'] ?? 'root',
            $config['password'] ?? '',
            [
                \PDO::ATTR_ERRMODE => \PDO::ERRMODE_EXCEPTION,
                \PDO::ATTR_DEFAULT_FETCH_MODE => \PDO::FETCH_ASSOC,
                \PDO::ATTR_EMULATE_PREPARES => false,
            ]
        );
    }
}
