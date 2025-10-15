<?php

namespace app\admin\model\supplier;

use PDO;
use app\admin\model\CompanyStaff;
use app\common\model\shop\User  as ShopStaffModel;
use app\common\model\supplier\Supplier as SupplierModel;

/**
 * 门店模型
 */
class Supplier extends SupplierModel
{
    /**
     * 添加
     */
    public function add($data, $app_id)
    {
        //添加门店
        $data['company_uuid'] = $data['uuid'] ?? 0;
        $data['is_open_h5'] = $data['is_open_scan'] ?? 0;
        $data['is_open_h5_order'] = $data['is_accept_scan_order'] ?? 0;
        $data['real_name'] = $data['user_name'] ?? '';
        $data['name'] = $data['name'] ?? '';
        $data['expire_time'] = $data['expire_time'] ?? 0;
        $data['app_id'] = $app_id;
        $data['is_main'] = ($data['parent_id'] ?? 0) > 0 ? 0 : 1;
        $data['delivery_set'] = ["10", "20"];
        $data['store_set'] = ["30", "40"];
        $data['delivery_time'] = json_encode(["00:00", "23:59"]);
        $data['pick_time'] = json_encode(["00:00", "23:59"]);
        $data['store_time'] = json_encode(["00:00", "23:59"]);
        $data['parent_id'] = ($data['parent_id'] ?? 0) ?: 0;
        $this->save($data);

        // 新增商家用户信息
        $shopUser = new CompanyStaff;
        $data['phone'] = $data['link_phone'] ?? '';
        if (!$shopUser->add($this->company_uuid, $data)) {
            $this->error = $shopUser->error;
            return false;
        }

        // 新增商家数据库
        return $this->addShopDatabase($data);
    }

    /**
     * 添加 数据库
     */
    private function addShopDatabase($data)
    {
        // 初始化数据库
        $filePath = realpath(root_path() . '/database/seeds/shop_01.sql');
        if (!file_exists($filePath)) {
            throw new \Exception('init sql file not exists');
        }

        //
        $host = env('DB_HOST');
        $port = env('DB_PORT');
        $pdo = new PDO("mysql:host={$host};port={$port}", env('DB_USERNAME'), env('DB_PASSWORD'));

        // 检测数据库
        $databaseName = $username = 'shop' . $this->uuid;
        $dbExists = $pdo->query("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '{$databaseName}'")->fetchColumn();
        if (!$dbExists) {
            $pdo->exec("CREATE DATABASE {$databaseName} CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci");
            // $pdo->exec("CREATE USER '{$username}'@'{$host}' IDENTIFIED BY '{$password}'");
            // $pdo->exec("GRANT ALL PRIVILEGES ON {$databaseName}.* TO '{$username}'@'{$host}'");
            // $pdo->exec("FLUSH PRIVILEGES");
            $pdo->exec("use {$databaseName}");
            $pdo->exec(file_get_contents($filePath));
        }

        // 初始化数据
        try {
            return $this->initShopBaseData($data);
        } catch (\Throwable $th) {
            $pdo = new PDO("mysql:host=" . env('DB_HOST') . ";port=" . env('DB_PORT'), env('DB_USERNAME'), env('DB_PASSWORD'));
            if ($pdo->query("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '{$databaseName}'")->fetchColumn()) {
                $pdo->exec("DROP DATABASE IF EXISTS {$databaseName}");
            }
            //
            $this->error = $th->getMessage();
            return false;
        }
    }

    /**
     * 初始化基础数据
     */
    private function initShopBaseData($data)
    {
        $companyUuid = $this->company_uuid;
        $shopUser = CompanyStaff::where('company_uuid', $companyUuid)->find();
        //
        $host = env('DB_HOST');
        $port = env('DB_PORT');
        $prefix = env('DB_PREFIX');
        $pdo = new PDO("mysql:host={$host};port={$port}", env('DB_USERNAME'), env('DB_PASSWORD'));
        // 检测数据库
        $databaseName = 'shop' . $companyUuid;
        $pdo->exec("use {$databaseName}");

        // app
        $datas = $shopUser->app?->append([])->toArray();
        $datas['auth_start_time'] = $shopUser->app->getData('auth_start_time');
        $datas['create_time'] = $shopUser->app->getData('create_time');
        $datas['update_time'] = $shopUser->app->getData('update_time');
        $pdo->exec($this->getInsertSql($prefix . 'company', $datas, [
            'uuid',
            'name',
            'logo',
            'expire_time',
            'auth_day',
            'auth_start_time',
            'status',
            'create_time',
            'update_time'
        ]));

        // shop_user
        $datas = $shopUser->toArray();
        $datas['is_super'] = 1; // 超级管理员
        $datas['password'] = salt_hash($data['password']);
        $datas['real_name'] = $shopUser->getData('username');
        $datas['create_time'] = $shopUser->getData('create_time');
        $datas['update_time'] = $shopUser->getData('update_time');
        $subShopUserFields = (new ShopStaffModel([], $this->company_uuid))->getFields();
        $pdo->exec($this->getInsertSql($prefix . 'staff', $datas, array_keys($subShopUserFields)));
        //
        $fileDataPath = realpath(root_path() . '/database/seeds/shop_init_data.sql');
        if (!file_exists($fileDataPath)) {
            throw new \Exception('init sql file not exists');
        }
        $pdo->exec(file_get_contents($fileDataPath));

        // 执行迁移文件
        exec("php /var/www/think migrate:run --db=" . $databaseName, $output, $result_code);
        if ($result_code != 0) {
            if ($pdo->query("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '{$databaseName}'")->fetchColumn()) {
                $pdo->exec("DROP DATABASE IF EXISTS {$databaseName}");
            }
            if ($output) {
                trace($output, 'error');
            }
            $this->error = "执行迁移文件错误: " . ($output[count($output) - 1] ?? 'runtime 目录 无写入权限');
            return false;
        }

        //  supplier
        $datas = $this->toArray();
        $datas['create_time'] = $this->getData('create_time');
        $datas['update_time'] = $this->getData('update_time');
        $pdo->exec($this->getInsertSql($prefix . 'company_setting', $datas, array_keys($this->getFields())));

        // 初始化支付方式
        $this->initPaymentMethod($pdo, $prefix, $datas['is_open_member']);

        // 同步设置
        $this->synchronousSetting($this, 'initShopBaseData');

        // 新建默认仓库
        $this->createDefaultWarehouse($pdo, $prefix);

        //
        return true;
    }

    /**
     * 创建默认仓库
     */
    private function createDefaultWarehouse($pdo, $prefix)
    {
        // 判断是否存在默认仓库和在途仓
        $defaultWarehouse = $pdo->query("SELECT uuid FROM {$prefix}warehouse WHERE type = 'normal' AND is_default = 1 AND delete_time = 0 LIMIT 1")->fetch(\PDO::FETCH_ASSOC);
        $transitWarehouse = $pdo->query("SELECT uuid FROM {$prefix}warehouse WHERE type = 'transit' AND delete_time = 0 LIMIT 1")->fetch(\PDO::FETCH_ASSOC);
        if ($defaultWarehouse && $transitWarehouse) {
            return;
        }

        // 定义默认仓和在途仓的多语言名称
        $defaultWarehouseNames = [
            'en' => 'Default',
            'zh' => 'Default',
            'zhtw' => 'Default',
            'th' => 'Default',
            'my' => 'Default',
            'ja' => 'Default',
            'ko' => 'Default',
            'tr' => 'Default',
            'sv' => 'Default',
        ];
        $transitWarehouseNames = [
            'en' => 'Transit',
            'zh' => 'Transit',
            'zhtw' => 'Transit',
            'th' => 'Transit',
            'my' => 'Transit',
            'ja' => 'Transit',
            'ko' => 'Transit',
            'tr' => 'Transit',
            'sv' => 'Transit',
        ];

        // 生成多语言UUID
        $defaultNameUuid = createUuid();
        $transitNameUuid = createUuid();

        $multiLanguageRecords = [];
        if (!$defaultWarehouse) {
            $multiLanguageRecords[] = [
                'uuid' => $defaultNameUuid,
                'en_name' => $defaultWarehouseNames['en'],
                'zh_name' => $defaultWarehouseNames['zh'],
                'zh_tw_name' => $defaultWarehouseNames['zhtw'],
                'th_name' => $defaultWarehouseNames['th'],
                'my_name' => $defaultWarehouseNames['my'],
                'ja_name' => $defaultWarehouseNames['ja'],
                'ko_name' => $defaultWarehouseNames['ko'],
                'tr_name' => $defaultWarehouseNames['tr'],
                'sv_name' => $defaultWarehouseNames['sv'],
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ];
        }
        if (!$transitWarehouse) {
            $multiLanguageRecords[] = [
                'uuid' => $transitNameUuid,
                'en_name' => $transitWarehouseNames['en'],
                'zh_name' => $transitWarehouseNames['zh'],
                'zh_tw_name' => $transitWarehouseNames['zhtw'],
                'th_name' => $transitWarehouseNames['th'],
                'my_name' => $transitWarehouseNames['my'],
                'ja_name' => $transitWarehouseNames['ja'],
                'ko_name' => $transitWarehouseNames['ko'],
                'tr_name' => $transitWarehouseNames['tr'],
                'sv_name' => $transitWarehouseNames['sv'],
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ];
        }

        // 插入多语言记录
        foreach ($multiLanguageRecords as $multiLanguageRecord) {
            $pdo->exec($this->getInsertSql($prefix . 'multi_language_name', $multiLanguageRecord, [
                'uuid',
                'en_name',
                'zh_name',
                'zh_tw_name',
                'th_name',
                'my_name',
                'ja_name',
                'ko_name',
                'tr_name',
                'sv_name',
                'create_time',
                'update_time',
                'delete_time'
            ]));
        }

        $warehouses = [];
        if (!$defaultWarehouse) {
            $warehouses[] = [
                'uuid' => createUuid(),
                'name' => json_encode($defaultWarehouseNames),
                'multi_language_name_uuid' => $defaultNameUuid,
                'type' => 'normal',
                'code' => 'WH01',
                'status' => 1,
                'contact' => '',
                'phone' => '',
                'address' => '',
                'is_default' => 1,
                'erp_code' => '',
                'headquarter_uuid' => 0,
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ];
        }

        if (!$transitWarehouse) {
            $warehouses[] = [
                'uuid' => createUuid(),
                'name' => json_encode($transitWarehouseNames),
                'multi_language_name_uuid' => $transitNameUuid,
                'type' => 'transit',
                'code' => 'WH02',
                'status' => 1,
                'contact' => '',
                'phone' => '',
                'address' => '',
                'is_default' => 0,
                'erp_code' => '',
                'headquarter_uuid' => 0,
                'create_time' => time(),
                'update_time' => time(),
                'delete_time' => 0,
            ];
        }


        // 插入仓库数据
        foreach ($warehouses as $warehouse) {
            $pdo->exec($this->getInsertSql($prefix . 'warehouse', $warehouse, [
                'uuid',
                'name',
                'multi_language_name_uuid',
                'type',
                'code',
                'status',
                'contact',
                'phone',
                'address',
                'is_default',
                'erp_code',
                'headquarter_uuid',
                'create_time',
                'update_time',
                'delete_time'
            ]));
        }
    }

    /**
     * 获取sql
     */
    private function getInsertSql($name, $datas, $fields = []): string
    {
        $filteredData = $fields ? array_filter($datas, function ($key) use ($fields) {
            return in_array($key, $fields);
        }, ARRAY_FILTER_USE_KEY) : $datas;
        $columns = implode(", ", array_keys($filteredData));
        $values = implode(", ", array_map(function ($value) {
            if (is_array($value)) {
                $value = str_replace("\\", "\\\\", $value);
                $value = str_replace("'", "\'", $value);
                return "'" . json_encode($value) . "'";
            } elseif (is_string($value)) {
                $value = str_replace("\\", "\\\\", $value);
                $value = str_replace("'", "\'", $value);
                return "'" . $value . "'";
            } elseif ($value === null) {
                return "'" . $value . "'";
            } else {
                $value = str_replace("'", "\'", $value);
                return $value;
            }
        }, array_values($filteredData)));
        //
        return "INSERT INTO $name ($columns) VALUES ($values);";
    }

    /**
     * 初始化支付方式
     */
    private function initPaymentMethod($pdo, $prefix, $isOpenMember)
    {
        // payment_method Cash-现金 Balance Payment-余额
        $paymentMethodList = [
            [
                'uuid' => createUuid(),
                'name' => 'Cash',
                'code' => 40,
                'payment_name' => 'Cash',
                'source' => 0,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 1,
                'status' => 1,
                'sort' => 1,
                'default_img' => '/image/pay/ja_pay.png',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'uuid' => createUuid(),
                'name' => 'Balance Payment',
                'code' => 10,
                'payment_name' => 'Balance Payment',
                'source' => 0,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
                'status' => $isOpenMember,
                'sort' => 1,
                'default_img' => '/image/pay/ja_pay.png',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'uuid' => createUuid(),
                'name' => 'WeChat Pay',
                'payment_name' => 'WeChat Pay',
                'source' => 2,
                'code' => 90111,
                'status' => 1,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
                'sort' => 0,
                'default_img' => '',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'uuid' => createUuid(),
                'name' => 'Alipay',
                'payment_name' => 'Alipay',
                'source' => 2,
                'code' => 90222,
                'status' => 1,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
                'sort' => 0,
                'default_img' => '',
                'create_time' => time(),
                'update_time' => time(),
            ],
            [
                'uuid' => createUuid(),
                'name' => 'QR PromptPay',
                'payment_name' => 'QR PromptPay',
                'source' => 2,
                'code' => 90333,
                'status' => 1,
                'is_show_cashier' => 1,
                'is_show_assistant' => 1,
                'is_show_member_recharge' => 0,
                'sort' => 0,
                'default_img' => '',
                'create_time' => time(),
                'update_time' => time(),
            ],
        ];
        foreach ($paymentMethodList as $paymentMethodItem) {
            $pdo->exec($this->getInsertSql($prefix . 'payment_method', $paymentMethodItem, [
                'uuid',
                'name',
                'code',
                'payment_name',
                'source',
                'is_show_cashier',
                'is_show_assistant',
                'is_show_member_recharge',
                'status',
                'sort',
                'default_img',
                'create_time',
                'update_time',
            ]));
        }
    }
}
