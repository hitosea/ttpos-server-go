<?php
/**
 * 请求服务
 * @author: xaboy<365615158@qq.com>
 * @day: 2017/11/23
 */
namespace help;

//use help\HttpService;
// HttpHelp::postRequest();
// HttpHelp::getRequest();

class HttpHelp {
    //错误信息
    private static $curlError;
    //header头信息
    private static $headerStr;
    //请求状态
    private static $status;

    /**
     * @return string
     */
    public static function getCurlError()
    {
        return self::$curlError;
    }

    public static function getStatus()
    {
        return self::$status;
    }

    public static function getRequest($url, $data = array(), $header = false, $timeout = 10)
    {
        if (!empty($data)) {
            $url .= (stripos($url, '?') === false ? '?' : '&');
            $url .= (is_array($data) ? http_build_query($data) : $data);
        }
        return self::request($url, 'get', array(), $header, $timeout);
    }
    
    public static function postRequest($url,  $data = array(), $header = false, $timeout = 10)
    {
        return self::request($url, 'post', $data, $header, $timeout);
    }

    public static function fileRequest($url, $filePath, $header = false, $timeout = 10)
    {
        self::$status = null;
        self::$curlError = null;
        self::$headerStr = null;
        // 初始化 cURL
        $curl = curl_init();
        // 设置 cURL 选项
        curl_setopt($curl, CURLOPT_URL, $url);
        curl_setopt($curl, CURLOPT_POST, 1);
        curl_setopt($curl, CURLOPT_TIMEOUT, $timeout);
        curl_setopt($curl, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($curl, CURLOPT_POSTFIELDS, [
            'file' => new \CURLFile($filePath)  // 添加需要上传的文件
        ]);
        if(is_array($header)){
            curl_setopt($curl, CURLOPT_HTTPHEADER, $header);
        }
        // 执行 cURL 请求
        self::$curlError = curl_error($curl);
        list($content, $status) = [curl_exec($curl), curl_getinfo($curl), curl_close($curl)];
        self::$status = $status;
        self::$headerStr = trim(substr($content, 0, $status['header_size']));
        // 
        return (intval($status["http_code"]) === 200) ? $content : false;
    }
    
    public static function request($url, $method = 'get', $data = array(), $header = false, $timeout = 15)
    {
        self::$status = null;
        self::$curlError = null;
        self::$headerStr = null;
        
        $curl = curl_init();
        $method = strtoupper($method);
        // 
        curl_setopt($curl, CURLOPT_URL, $url);
        //请求方式
        curl_setopt($curl, CURLOPT_CUSTOMREQUEST, $method);
        //post请求
        if ($method == 'POST') curl_setopt($curl, CURLOPT_POSTFIELDS, $data);
        //超时时间
        curl_setopt($curl, CURLOPT_TIMEOUT, $timeout);
        //设置header头
        if ( $header === 'form') {
            curl_setopt($curl, CURLOPT_HTTPHEADER,array(
                'Content-Type: application/x-www-form-urlencoded;',
                'Cache-Control: no-cache',
                'Pragma: no-cache'
            ));
        } else if(is_array($header)){
            curl_setopt($curl, CURLOPT_HTTPHEADER, $header);
        }
        else if ( $header !== false  ) {
        	curl_setopt($curl, CURLOPT_HTTPHEADER,array(
                'Content-Type: application/json; charset=utf-8',
                'Cache-Control: no-cache',
                'Pragma: no-cache'
        	));
		}
        curl_setopt($curl, CURLOPT_FAILONERROR, false);
        //返回抓取数据
        curl_setopt($curl, CURLOPT_RETURNTRANSFER, true);
        //输出header头信息
        curl_setopt($curl, CURLOPT_HEADER, true);
        //TRUE 时追踪句柄的请求字符串，从 PHP 5.1.3 开始可用。这个很关键，就是允许你查看请求header
        curl_setopt($curl, CURLINFO_HEADER_OUT, true);
        //https请求
        if (1 == strpos("$" . $url, "https://")) {
            curl_setopt($curl, CURLOPT_SSL_VERIFYPEER, false);
            curl_setopt($curl, CURLOPT_SSL_VERIFYHOST, false);
        }
        self::$curlError = curl_error($curl);
        list($content, $status) = [curl_exec($curl), curl_getinfo($curl), curl_close($curl)];
        self::$status = $status;
        self::$headerStr = trim(substr($content, 0, $status['header_size']));
        $content = trim(substr($content, $status['header_size']));
        return (intval($status["http_code"]) === 200) ? $content : false;
    }

    public static function getHeaderStr()
    {
        return self::$headerStr;
    }

    public static function getHeader()
    {
        $headArr = explode("\r\n", self::$headerStr);
        return $headArr;
    }

}