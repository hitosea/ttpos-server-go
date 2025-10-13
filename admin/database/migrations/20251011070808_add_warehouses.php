<?php

use think\facade\Env;
use think\migration\Migrator;
use think\migration\db\Column;
use PDO;

class AddWarehouses extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
        $this->initWarehousesForAllShops();
    }

    /**
     * 为所有shop数据库初始化仓库
     */
    private function initWarehousesForAllShops()
    {
        $host = Env::get('DB_HOST');
        $port = Env::get('DB_PORT');
        $username = Env::get('DB_USERNAME');
        $password = Env::get('DB_PASSWORD');
        $saasDatabase = Env::get('DB_DATABASE');
        $prefix = Env::get('DB_PREFIX');

        // 连接数据库
        $pdo = new PDO("mysql:host={$host};port={$port}", $username, $password);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

        // 查询saas库中所有有效的公司
        $pdo->exec("use {$saasDatabase}");
        $sql = "SELECT uuid FROM {$prefix}company WHERE delete_time = 0";
        $companies = $pdo->query($sql)->fetchAll(PDO::FETCH_ASSOC);

        echo "找到 " . count($companies) . " 个公司，开始检查仓库...\n";

        foreach ($companies as $company) {
            $companyUuid = $company['uuid'];
            $shopDatabase = 'shop' . $companyUuid;

            try {
                // 检查shop数据库是否存在
                $dbExists = $pdo->query("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '{$shopDatabase}'")->fetchColumn();
                if (!$dbExists) {
                    echo "数据库 {$shopDatabase} 不存在，跳过\n";
                    continue;
                }

                // 切换到shop数据库
                $pdo->exec("use {$shopDatabase}");

                // 检查warehouse表是否存在
                $tableExists = $pdo->query("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '{$shopDatabase}' AND table_name = '{$prefix}warehouse'")->fetchColumn();
                if (!$tableExists) {
                    echo "数据库 {$shopDatabase} 中不存在warehouse表，跳过\n";
                    continue;
                }

                // 检查是否存在默认仓库和在途仓库
                $defaultWarehouse = $pdo->query("SELECT uuid FROM {$prefix}warehouse WHERE type = 'normal' AND is_default = 1 AND delete_time = 0 LIMIT 1")->fetch(PDO::FETCH_ASSOC);
                $transitWarehouse = $pdo->query("SELECT uuid FROM {$prefix}warehouse WHERE type = 'transit' AND delete_time = 0 LIMIT 1")->fetch(PDO::FETCH_ASSOC);

                // 如果没有默认仓库，创建
                if (!$defaultWarehouse) {
                    echo "为 {$shopDatabase} 创建默认仓库...\n";
                    $defaultWarehouse = $this->createWarehouse($pdo, $prefix, 'normal', 'Default', 'WH01', 1);
                } else {
                    echo "数据库 {$shopDatabase} 已存在默认仓库\n";
                }

                // 如果没有在途仓库，创建
                if (!$transitWarehouse) {
                    echo "为 {$shopDatabase} 创建在途仓库...\n";
                    $this->createWarehouse($pdo, $prefix, 'transit', 'Transit', 'WH02', 0);
                } else {
                    echo "数据库 {$shopDatabase} 已存在在途仓库\n";
                }

                // 如果有默认仓库，将material表的数据关联到warehouse_item
                if ($defaultWarehouse) {
                    $this->syncMaterialToWarehouseItem($pdo, $prefix, $defaultWarehouse['uuid'], $shopDatabase);
                }

            } catch (\Exception $e) {
                echo "处理数据库 {$shopDatabase} 时出错: " . $e->getMessage() . "\n";
                continue;
            }
        }

        echo "仓库初始化完成\n";
    }

    /**
     * 创建仓库
     */
    private function createWarehouse($pdo, $prefix, $type, $name, $code, $isDefault)
    {
        // 定义多语言名称
        $warehouseNames = [
            'en' => $name,
            'zh' => $name,
            'zhtw' => $name,
            'th' => $name,
            'my' => $name,
            'ja' => $name,
            'ko' => $name,
            'tr' => $name,
            'sv' => $name,
        ];

        // 生成多语言UUID
        $nameUuid = $this->createUuid();

        // 创建多语言记录
        $multiLanguageRecord = [
            'uuid' => $nameUuid,
            'en_name' => $warehouseNames['en'],
            'zh_name' => $warehouseNames['zh'],
            'zh_tw_name' => $warehouseNames['zhtw'],
            'th_name' => $warehouseNames['th'],
            'my_name' => $warehouseNames['my'],
            'ja_name' => $warehouseNames['ja'],
            'ko_name' => $warehouseNames['ko'],
            'tr_name' => $warehouseNames['tr'],
            'sv_name' => $warehouseNames['sv'],
            'create_time' => time(),
            'update_time' => time(),
            'delete_time' => 0,
        ];

        $pdo->exec($this->getInsertSql($prefix . 'multi_language_name', $multiLanguageRecord));

        // 创建仓库记录
        $warehouseUuid = $this->createUuid();
        $warehouse = [
            'uuid' => $warehouseUuid,
            'name' => json_encode($warehouseNames),
            'multi_language_name_uuid' => $nameUuid,
            'type' => $type,
            'code' => $code,
            'status' => 1,
            'contact' => '',
            'phone' => '',
            'address' => '',
            'is_default' => $isDefault,
            'erp_code' => '',
            'headquarter_uuid' => 0,
            'create_time' => time(),
            'update_time' => time(),
            'delete_time' => 0,
        ];

        $pdo->exec($this->getInsertSql($prefix . 'warehouse', $warehouse));

        return ['uuid' => $warehouseUuid];
    }

    /**
     * 同步material表数据到warehouse_item表
     */
    private function syncMaterialToWarehouseItem($pdo, $prefix, $defaultWarehouseUuid, $shopDatabase)
    {
        // 检查warehouse_item表是否存在
        $tableExists = $pdo->query("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '{$shopDatabase}' AND table_name = '{$prefix}warehouse_item'")->fetchColumn();
        if (!$tableExists) {
            echo "数据库 {$shopDatabase} 中不存在warehouse_item表，跳过同步\n";
            return;
        }

        // 查询所有有效的material
        $materials = $pdo->query("SELECT uuid, stock_num, code FROM {$prefix}material WHERE delete_time = 0")->fetchAll(PDO::FETCH_ASSOC);

        if (empty($materials)) {
            echo "数据库 {$shopDatabase} 中没有material数据，跳过同步\n";
            return;
        }

        echo "为 {$shopDatabase} 同步 " . count($materials) . " 个物料到warehouse_item表...\n";

        $createCount = 0;
        $updateCount = 0;

        foreach ($materials as $material) {
            // 检查该物料是否已经存在于warehouse_item中
            $existingItem = $pdo->query("SELECT uuid FROM {$prefix}warehouse_item WHERE material_uuid = {$material['uuid']} AND warehouse_uuid = {$defaultWarehouseUuid} AND delete_time = 0 LIMIT 1")->fetch(PDO::FETCH_ASSOC);

            if (!$existingItem) {
                // 创建warehouse_item记录
                $warehouseItem = [
                    'uuid' => $this->createUuid(),
                    'warehouse_uuid' => $defaultWarehouseUuid,
                    'material_uuid' => $material['uuid'],
                    'material_code' => $material['code'] ?? '',
                    'stock' => $material['stock_num'] ?? 0,
                    'reserved_stock' => 0,
                    'create_time' => time(),
                    'update_time' => time(),
                    'delete_time' => 0,
                ];

                $pdo->exec($this->getInsertSql($prefix . 'warehouse_item', $warehouseItem));
                $createCount++;
            } else {
                // 更新库存数量
                $updateSql = "UPDATE {$prefix}warehouse_item SET stock = {$material['stock_num']}, update_time = " . time() . " WHERE uuid = {$existingItem['uuid']}";
                $pdo->exec($updateSql);
                $updateCount++;
            }
        }

        echo "数据库 {$shopDatabase} 物料同步完成（新增: {$createCount}, 更新: {$updateCount}）\n";
    }

    /**
     * 生成UUID
     */
    private function createUuid()
    {
        return createUuid();
    }

    /**
     * 生成INSERT SQL语句
     */
    private function getInsertSql($tableName, $data, $fields = [])
    {
        if (empty($fields)) {
            $fields = array_keys($data);
        }

        $filteredData = array_filter($data, function ($key) use ($fields) {
            return in_array($key, $fields);
        }, ARRAY_FILTER_USE_KEY);

        $columns = implode(", ", array_keys($filteredData));
        $values = implode(", ", array_map(function ($value) {
            if (is_array($value)) {
                $value = str_replace("\\", "\\\\", json_encode($value));
                $value = str_replace("'", "\'", $value);
                return "'" . $value . "'";
            } elseif (is_string($value)) {
                $value = str_replace("\\", "\\\\", $value);
                $value = str_replace("'", "\'", $value);
                return "'" . $value . "'";
            } elseif ($value === null) {
                return "NULL";
            } else {
                return $value;
            }
        }, array_values($filteredData)));

        return "INSERT INTO $tableName ($columns) VALUES ($values);";
    }
}
