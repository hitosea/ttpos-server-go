<?php

use think\facade\Env;
use think\migration\Migrator;
use think\migration\db\Column;
use think\facade\Db;

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
        $this->initWarehousesAndItem();
    }

    /**
     * 为所有shop数据库初始化仓库
     */
    private function initWarehousesAndItem()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        $company = $db->name("company")->where("delete_time", 0)->find();

        $host = Env::get('DB_HOST');
        $port = Env::get('DB_PORT');
        $username = Env::get('DB_USERNAME');
        $password = Env::get('DB_PASSWORD');
        $prefix = Env::get('DB_PREFIX');

        // 连接数据库
        $pdo = new \PDO("mysql:host={$host};port={$port}", $username, $password);
        $pdo->setAttribute(\PDO::ATTR_ERRMODE, \PDO::ERRMODE_EXCEPTION);
        $pdo->exec("use shop{$company['uuid']}");

        // 检查是否存在默认仓库和在途仓库
        $defaultWarehouse = $pdo->query("SELECT uuid FROM {$prefix}warehouse WHERE type = 'normal' AND is_default = 1 AND delete_time = 0 LIMIT 1")->fetch(\PDO::FETCH_ASSOC);
        $transitWarehouse = $pdo->query("SELECT uuid FROM {$prefix}warehouse WHERE type = 'transit' AND delete_time = 0 LIMIT 1")->fetch(\PDO::FETCH_ASSOC);

        // 如果没有默认仓库，创建
        if (!$defaultWarehouse) {
            $warehouseNames = [
                'en' => "Default warehouse",
                'zh' => "默认仓库",
                'zhtw' => "默認倉庫",
                'th' => "คลังสินค้าเบื้องต้น",
                'my' => "ပုံမှန်ဂိုဒေါင်",
                'ja' => "デフォルトの倉庫",
                'ko' => "기본 창고",
                'tr' => "Varsayılan depo",
                'sv' => "Standardlager",
            ];
            $defaultWarehouse = $this->createWarehouse($pdo, $prefix, 'normal', $warehouseNames, 'WH01', 1);
        }

        // 如果没有在途仓库，创建
        if (!$transitWarehouse) {
            $warehouseNames = [
                'en' => "Transit Warehouse",
                'zh' => "在途仓库",
                'zhtw' => "在途倉庫",
                'th' => "คลังสินค้าระหว่างทาง",
                'my' => "လမ်းခရီးရှိဂိုဒေါင်",
                'ja' => "輸送中の倉庫",
                'ko' => "운송 중 창고",
                'tr' => "Yoldaki depo",
                'sv' => "Transitlager",
            ]; 
            $this->createWarehouse($pdo, $prefix, 'transit', $warehouseNames, 'WH02', 0);
        }

        // 如果有默认仓库，将material表的数据关联到warehouse_item
        if ($defaultWarehouse) {
            $this->syncMaterialToWarehouseItem($pdo, $prefix, $defaultWarehouse['uuid']);
        }
    }

    /**
     * 创建仓库
     */
    private function createWarehouse($pdo, $prefix, $type, $warehouseNames, $code, $isDefault)
    {
        // 生成多语言UUID
        $nameUuid = createUuid();

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
        $warehouseUuid = createUuid();
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
    private function syncMaterialToWarehouseItem($pdo, $prefix, $defaultWarehouseUuid)
    {
        // 查询所有有效的material
        $materials = $pdo->query("SELECT uuid, stock_num, code FROM {$prefix}material WHERE delete_time = 0")->fetchAll(\PDO::FETCH_ASSOC);

        if (empty($materials)) {
            return;
        }

        $createCount = 0;
        $updateCount = 0;

        foreach ($materials as $material) {
            // 检查该物料是否已经存在于warehouse_item中
            $existingItem = $pdo->query("SELECT uuid FROM {$prefix}warehouse_item WHERE material_uuid = {$material['uuid']} AND warehouse_uuid = {$defaultWarehouseUuid} AND delete_time = 0 LIMIT 1")->fetch(\PDO::FETCH_ASSOC);

            if (!$existingItem) {
                // 创建warehouse_item记录
                $warehouseItem = [
                    'uuid' => createUuid(),
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
