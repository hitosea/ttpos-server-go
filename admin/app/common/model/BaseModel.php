<?php

namespace app\common\model;

use think\Model;
use help\HttpHelp;
use think\facade\Config;

/**
 * 模型基类
 */
class BaseModel extends Model
{
    public static $app_id;

    protected $alias = '';

    protected $error = '';

    protected $errorData = [];

    protected int $errorCode = 0;

    /**
     * 构造方法
     */
    public function __construct(array|object $data = [], $appId = null)
    {
        // 后期静态绑定app_id
        if ($appId !== null) {
            self::$app_id = $appId;
        } else {
            self::bindAppId();
        }
        // 应用名称
        $appName = app('http')->getName();
        // 动态设置数据库连接配置
        if ((self::$app_id && $appName != 'admin' && $this->name != 'app') || ($appName == 'admin' && request()->appId) || $appId) {
            $name = 'shop' . self::$app_id;
            $config = config('database');
            if (!isset($config['connections'][$name])) {
                $mysql = $config['connections']['mysql'];
                $mysql['database'] = $name;
                $mysql['username'] = env('DB_USERNAME');
                $mysql['password'] = env('DB_PASSWORD');
                $config['connections'] = array_merge($config['connections'], [$name => $mysql]);
                Config::set($config, 'database');
            }
            $this->setConnection($name);
        } else {
            $this->setConnection('mysql');
        }
        // 
        if (strlen($this->connection) == 14 && $this->name == 'staff') {
            $this->name = 'shop_user';
        }
        //
        parent::__construct($data);
    }

    /**
     * 后期静态绑定类名称
     */
    private static function bindAppId()
    {
        if ($app = app('http')->getName()) {
            if ($app != 'admin' && $app != 'job') {
                if (request()->appId === 0) {
                    self::$app_id = 0;
                } else {
                    self::$app_id = request()->appId ?: 0;
                }
            }
        }
    }

    /**
     * 设置app_id (所有)
     */
    public function setAppId($appId = 0)
    {
        request()->appId = $appId;
        self::$app_id = $appId;
        return $this;
    }

    /**
     * 设置默认的检索数据
     */
    protected function setQueryDefaultValue(&$query, $default = [])
    {
        $data = array_merge($default, $query);
        foreach ($query as $key => $val) {
            if (empty($val) && isset($default[$key])) {
                $data[$key] = $default[$key];
            }
        }
        return $data;
    }

    /**
     * 设置基础查询条件（用于简化基础alias和field）
     */
    public function setBaseQuery($alias = '', $join = [])
    {
        // 设置别名
        $aliasValue = $alias ?: $this->alias;
        $model = $this->alias($aliasValue)->field("{$aliasValue}.*");
        // join条件
        if (!empty($join)) : foreach ($join as $item):
                $model->join($item[0], "{$item[0]}.{$item[1]}={$aliasValue}."
                    . (isset($item[2]) ? $item[2] : $item[1]));
            endforeach;
        endif;
        return $model;
    }

    /**
     * 批量更新数据(支持带where条件)
     */
    public function updateAll($data)
    {
        return $this->transaction(function () use ($data) {
            $result = [];
            foreach ($data as $key => $item) {
                $result[$key] = self::update($item['data'], $item['where']);
            }
            return $this->toCollection($result);
        });
    }

    // 新增前
    public static function onBeforeInsert(Model $model)
    {
        $fields = $model->getTable() ? $model->getFields() : [];
        if (in_array('uuid', array_keys($fields)) && empty($model->uuid)) {
            $model->uuid = createUuid();
        }
    }

    // 更新前
    public static function onBeforeUpdate(Model $model)
    {
        if ($model->createTime && $model[$model->createTime]) {
            unset($model[$model->createTime]);
        }
        if ($model->updateTime && $model[$model->updateTime]) {
            $model[$model->updateTime] = $model->autoWriteTimestamp($model->updateTime);
        }
    }

    // 新增后
    public static function onAfterInsert(Model $model)
    {
        $model->updateMainCache($model);
    }

    // 更新后
    public static function onAfterUpdate(Model $model)
    {
        $model->updateMainCache($model);
    }

    /**
     * 获取当前调用的模块名称
     */
    protected static function getCalledModule()
    {
        if (preg_match('/app\\\(\w+)/', get_called_class(), $class)) {
            return $class[1];
        }
        return false;
    }

    /**
     * 返回模型的错误信息
     */
    public function getError()
    {
        return $this->error;
    }

    /**
     * 返回模型的错误信息
     */
    public function getErrorData()
    {
        return $this->errorData;
    }

    /**
     * 返回模型的code
     */
    public function getErrorCode()
    {
        return $this->errorCode;
    }

    /**
     * 处理错误
     * @param $message
     * @param null $queue
     */
    public function handleError($message, $errorCode = 0, $queue = null, $queueAll = null)
    {
        if ($queue) {
            $queue->release();
        }
        if ($queueAll) {
            $queueAll->release();
        }
        if ($errorCode) {
            $this->errorCode = $errorCode;
        }
        //
        $this->error = $message;
        //
        return false;
    }

    /**
     * 返回模型的列表
     */
    public function lists()
    {
        return $this->select();
    }

    /**
     * 更新缓存
     * 异步调用 Go Main 项目的缓存失效接口
     * 
     * @param Model $model 模型实例
     * @return void
     */
    public function updateMainCache(Model $model)
    {
        $tableName = $model->getTable();
        
        // 跳过不需要更新缓存的表
        if ($tableName == 'ttpos_staff_operation_log') {
            return;
        }

        // 获取商户ID
        $appId = request()->appId ?? 0;
        if (empty($appId)) {
            return;
        }

        try {
            // 调用 Go Main 项目的缓存失效接口（内部接口）
            $url = 'http://nginx/api/v1/internal/shop/product/cache/invalidate';
            // 使用 HttpHelp 发送 POST 请求
            HttpHelp::postRequest($url,  json_encode([
                'company_uuid' => $appId,
            ]), [
                'Content-Type: application/json',
                'X-API-KEY: ' . env('JWT_SECRET'),
            ], 1);
        } catch (\Exception $e) {
            \think\facade\Log::error('缓存失效任务执行失败: ' . $e->getMessage(), [
                'table_name' => $tableName,
                'app_id' => $appId,
                'trace' => $e->getTraceAsString(),
            ]);
        }
    }
}
