<?php

namespace app\admin\model\supplier;

use PDO;
use app\admin\model\Shop as ShopUser;
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
        $shopUser = new ShopUser;
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
        $pdo = new PDO("mysql:host={$host};port={$port}", 'root', env('DB_ROOT_PASSWORD'));

        // 检测数据库
        $databaseName = $username = 'shop' . $this->uuid;
        $dbExists = $pdo->query("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '{$databaseName}'")->fetchColumn();
        if (!$dbExists) {
            $pdo->exec("CREATE DATABASE {$databaseName}");
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
            $pdo = new PDO("mysql:host=" . env('DB_HOST') . ";port=" . env('DB_PORT'), 'root', env('DB_ROOT_PASSWORD'));
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
        $shopUser = ShopUser::where('company_uuid', $companyUuid)->find();
        //
        $host = env('DB_HOST');
        $port = env('DB_PORT');
        $prefix = env('DB_PREFIX');
        $pdo = new PDO("mysql:host={$host};port={$port}", 'root', env('DB_ROOT_PASSWORD'));
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
        $datas['create_time'] = $shopUser->getData('create_time');
        $datas['update_time'] = $shopUser->getData('update_time');
        $subShopUserFields = (new ShopUser([], $this->app_id))->getFields();
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
        $pdo->exec($this->getInsertSql($prefix . 'supplier', $datas, array_keys($this->getFields())));

        // 同步设置
        $this->synchronousSetting($this, 'initShopBaseData');

        //
        return true;
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
}
