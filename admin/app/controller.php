<?php

declare(strict_types=1);

namespace app;

use think\App;
use think\Validate;

/**
 * 控制器基础类
 */
abstract class controller
{
    /**
     * Request实例
     */
    protected $request;

    /**
     * 应用实例
     */
    protected $app;

    /**
     * 是否批量验证
     */
    protected $batchValidate = false;

    /**
     * 控制器中间件
     */
    protected $middleware = [];

    /**
     * 构造方法
     */
    public function __construct(App $app)
    {
        $this->app = $app;
        $this->request = $this->app->request;

        // 控制器初始化
        $this->initialize();
    }

    // 初始化
    protected function initialize() {}

    /**
     * 验证数据
     */
    protected function validate(array $data, $validate, array $message = [], bool $batch = false)
    {
        if (is_array($validate)) {
            $v = new Validate();
            $v->rule($validate);
        } else {
            if (strpos($validate, '.')) {
                // 支持场景
                list($validate, $scene) = explode('.', $validate);
            }
            $class = false !== strpos($validate, '\\') ? $validate : $this->app->parseClass('validate', $validate);
            $v = new $class();
            if (!empty($scene)) {
                $v->scene($scene);
            }
        }

        $v->message($message);

        // 是否批量验证
        if ($batch || $this->batchValidate) {
            $v->batch(true);
        }

        return $v->failException(true)->check($data);
    }

    /**
     * 返回封装后的 API 数据到客户端
     */
    protected function renderJson($code = 1, $msg = '', $data = [])
    {
        return compact('code', 'msg', 'data');
    }

    /**
     * 返回操作成功json
     */
    protected function renderSuccess($msg = 'success', $data = [])
    {
        return json($this->renderJson(1, __($msg), $data));
    }

    /**
     * 返回操作成功json - 直接输出 - 响应完成, 关闭连接 - 但进程还在
     */
    protected function renderEchoSuccess($msg = 'success', $data = [])
    {
        $res = json($this->renderJson(1, __($msg), $data));
        header('Access-Control-Allow-Credentials: true');
        header('Access-Control-Max-Age: 1800');
        header('Access-Control-Allow-Methods: GET, POST, PATCH, PUT, DELETE, OPTIONS');
        header('Access-Control-Allow-Headers: *');
        foreach ($res->getHeader() as $name => $val) {
            header($name . (!is_null($val) ? ':' . $val : ''));
        }
        echo json_encode($res->getData());
        fastcgi_finish_request(); /* 响应完成, 关闭连接 */
    }

    /**
     * 返回操作失败json
     */
    protected function renderError($msg = 'error', $data = [], $code = 0)
    {
        return json($this->renderJson($code, __($msg), $data));
    }

    /**
     * 获取post数据 (数组)
     */
    protected function postData($key = null)
    {
        return $this->request->param(is_null($key) ? '' : $key . '/a');
    }

    /**
     * 获取post数据 (数组)
     */
    protected function getData($key = null)
    {
        return $this->request->get(is_null($key) ? '' : $key);
    }
}
