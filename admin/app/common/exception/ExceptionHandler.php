<?php

namespace app\common\exception;

use app\common\enum\order\OrderErrorEnum;
use Throwable;
use think\Response;
use think\facade\Log;
use think\exception\Handle;
use think\db\exception\PDOException;

/**
 * 重写Handle的render方法，实现自定义异常消息
 */
class ExceptionHandler extends Handle
{
    private $code;
    private $message;
    private $data;

    /**
     * 输出异常信息
     * @param \think\Request $request
     * @param Throwable $e
     * @return Response
     */
    public function render($request, Throwable $e): Response
    {
        if ($e instanceof PDOException) {
            if (strstr($e->getMessage(), 'String data, right truncated: 1406 Data too long for column')) {
                $this->code = 0;
                $this->message = "数据长度超过限制";
            } else if (strstr($e->getMessage(), 'Numeric value out of range: 1264 Out of range value')) {
                $this->code = OrderErrorEnum::ORDER_OVERAGE;
                $this->message = "订单已达最大金额，请重新调整订单金额";
            } else if (strstr($e->getMessage(), '1305 SAVEPOINT trans2 does not exist')) {
                $this->code = 0;
                $this->message = "网络繁忙";
            } else if (strstr($e->getMessage(), 'Duplicate entry')) {
                $this->code = 0;
                $this->message = "主键重复";
            } else {
                $this->code = 0;
                $this->message = $e->getMessage();
            }
        } else if ($e instanceof BaseException) {
            $this->code = $e->code;
            $this->message = $e->message;
            $this->data = $e->data;
        } else if ($e instanceof \hg\apidoc\exception\HttpException) {
            $this->code = $e->getCode();
            $this->message = $e->getMessage();
        } else {
            $this->code = 0;
            $this->message = $e->getMessage() ?: '很抱歉，服务器内部错误';
        }
        //
        $this->recordErrorLog($request, $e);
        //
        $data = [
            'msg' => __($this->message),
            'code' => $this->code,
            'data' => $this->data,
        ];
        if (env('APP_DEBUG') || env('SERVER_MODE') == 'debug') {
            $data['line'] = $e->getline() ?: '';
            $data['file'] = $e->getFile() ?: '';
        }
        //
        return json($data);
    }

    /**
     * 将异常写入日志
     */
    private function recordErrorLog($request, Throwable $e)
    {
        try {
            $url = $request->domain() . $request->url();
            Log::write($url . ' : ' . $e->getMessage() ?: '服务器内部错误', 'error');
            Log::write($e->getTraceAsString(), 'error');
        } catch (\Throwable $th) {
            //throw $th;
        }
    }
}
