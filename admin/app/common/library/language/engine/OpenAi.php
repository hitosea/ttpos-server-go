<?php

namespace app\common\library\language\engine;

use think\facade\Log;

class OpenAi
{

    // const URL = "http://103.63.139.229:8088/api/translate";
    const URL = "https://aitrans.ttpos.com/translate";
    const CURL_TIMEOUT = 10000;
    const CONNECT_TIMEOUT = 10;

    private string $APP_KEY = '';
    private string $SEC_KEY = '';

    /**
     * @param $APP_KEY
     * @param $SEC_KEY
     */
    public function __construct($APP_KEY = '', $SEC_KEY = '')
    {
        $this->APP_KEY = $APP_KEY;
        $this->SEC_KEY = $SEC_KEY;
    }

    /**
     * 翻译
     * @param array $q
     * @param $from
     * @return array
     * @throws \Exception
     */
    public function translate($q, $from = 'zh')
    {
        try {
            $data = ['data' => []];
            foreach ($q as $key => $value) {
                $data['data'][] = [
                    'lang' => is_string($key) ? $key : $from,
                    'content' => $value
                ];
            }

            $jsonData = json_encode($data);
            $ret = $this->request(self::URL, 'post', $jsonData, true);
            $res = json_decode($ret, true);
            if ($res['code'] == 200 && !empty($res['data'])) {
                $dataArray = $res['data'];
                $res['data'] = is_array($dataArray) && !empty($dataArray) ? $dataArray : [];
                return $res['data'];
            }
            return [];
        } catch (\Exception $e) {
            throw new \Exception("OpenAi Translation failed: " . $e->getMessage());
        }
    }

    /**
     * 转发请求
     */
    public function forward($data)
    {
        try {
            $jsonData = json_encode($data, JSON_UNESCAPED_UNICODE);
            $ret = $this->request(self::URL, 'post', $jsonData, true);
            $res = json_decode($ret, true);
            return $res;
        } catch (\Exception $e) {
            throw new \Exception("Forward OpenAi Translation failed: " . $e->getMessage());
        }
    }

    /**
     * 请求
     *
     * @param [type] $url
     * @param string $method
     * @param array $data
     * @param boolean $header
     * @param [type] $timeout
     * @return bool|string
     */
    private function request($url, $method = 'get', $data = array(), $header = false, $timeout = self::CURL_TIMEOUT)
    {
        $curl = curl_init();
        $method = strtoupper($method);
        curl_setopt($curl, CURLOPT_URL, $url);
        curl_setopt($curl, CURLOPT_CUSTOMREQUEST, $method);
        if ($method == 'POST') curl_setopt($curl, CURLOPT_POSTFIELDS, $data);
        curl_setopt($curl, CURLOPT_TIMEOUT, $timeout);
        curl_setopt($curl, CURLOPT_CONNECTTIMEOUT, self::CONNECT_TIMEOUT);
        if ($header === 'form') {
            curl_setopt($curl, CURLOPT_HTTPHEADER, array(
                'Content-Type: application/x-www-form-urlencoded;',
                'Cache-Control: no-cache',
                'Pragma: no-cache'
            ));
        } else if (is_array($header)) {
            curl_setopt($curl, CURLOPT_HTTPHEADER, $header);
        } else if ($header !== false) {
            curl_setopt($curl, CURLOPT_HTTPHEADER, array(
                'Content-Type: application/json; charset=utf-8',
                'Cache-Control: no-cache',
                'Pragma: no-cache'
            ));
        }
        curl_setopt($curl, CURLOPT_FAILONERROR, false);
        curl_setopt($curl, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($curl, CURLOPT_HEADER, true);
        curl_setopt($curl, CURLINFO_HEADER_OUT, true);
        if (1 == strpos("$" . $url, "https://")) {
            curl_setopt($curl, CURLOPT_SSL_VERIFYPEER, false);
            curl_setopt($curl, CURLOPT_SSL_VERIFYHOST, false);
            curl_setopt($curl, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_1_1); 
        }
        list($content, $status) = [curl_exec($curl), curl_getinfo($curl), curl_close($curl)];
        if ($content === false) {
            Log::error('Curl error: ' . curl_error($curl));
            Log::error('Curl error number: ' . curl_errno($curl));
            Log::error('Curl info: ' . print_r(curl_getinfo($curl), true));
        }
        $content = trim(substr($content, $status['header_size']));
        return (intval($status["http_code"]) === 200) ? $content : false;
    }
}
