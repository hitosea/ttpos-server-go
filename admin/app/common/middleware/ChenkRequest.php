<?php

namespace app\common\middleware;

use help\DooHelp;
use help\JsEncrypt;
use app\common\exception\BaseException;

/**
 * token验证中间件
 * Class AuthTokenMiddleware
 * @package app\http\middleware
 */
class ChenkRequest
{
    //自执行中间件方法
    public function handle($request, \Closure $next)
    {
        //检测是否有木马文件上传
        if ($files = $request->file()) {
            foreach ($files as $key => $file) {
                if (strstr(strtolower($file->getOriginalName()), ".php")) {
                    throw new BaseException(['msg' => '监听到木马文件，禁止上传！']);
                }
            }
        }

        // 解密内容
        $headerEncrypt = $request->header('encrypt');
        $encrypt = $headerEncrypt ? JsEncrypt::pgpParseStr($headerEncrypt) : '';
        if ($encrypt && $request->isPost() && $content = $request->param('encrypted')) {
            // 新版本解密提交的内容
            if ($encrypt['encrypt_type'] === 'jsencrypt' && $encrypt['client_type'] === 'jsencrypt') {
                $content = JsEncrypt::decryptApi($content, $encrypt['encrypt_id']);
                if ($content) {
                    $request->withInput($content);
                }
            } else if ($encrypt['encrypt_type'] === 'pgp') {
                $encrypt = DooHelp::pgpParseStr($headerEncrypt);
                $content = DooHelp::pgpDecryptApi($content, $encrypt['encrypt_id']);
                if ($content) {
                    $request->withInput(json_encode($content));
                }
            }
        }

        $response = $next($request);

        // 加密返回内容
        if ($encrypt && $content = $response->getContent()) {
            if ($encrypt['client_type'] === 'pgp') {
                $encrypted = DooHelp::pgpEncryptApi($content, $encrypt['client_key']);
            }
            if ($encrypt['client_type'] === 'jsencrypt') {
                $encrypted = JsEncrypt::encrypt($content, $encrypt['client_key']);
            }
            $response->content(json_encode([
                'encrypted' => $encrypted
            ]));
        }

        return $response;
    }

    //<!--  = 检测是否存在sql注入 =  -->
    private function is_exist_sql_inject()
    {
        $pattern = '/\binsert\s|\bselect\s|\bselect\*\s|\bselect\*from\s|\bselect\*from\(|\bupdate\s|\bdelete\s|\bcreate\s|\balter\s|\bdrop\s|(\bexec|\bexecute)[\s\(]/';
        $param =  strtolower(json_encode($_REQUEST));
        return preg_match($pattern, $param) ? true : false;
    }
}
