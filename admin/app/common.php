<?php

use \Firebase\JWT\JWT;
use \Firebase\JWT\Key;
use help\PrintTexHelp;
use help\SnowflakeHelp;
use think\facade\Cache;
use think\facade\Config;
use think\facade\Request;
use app\common\model\shop\User;
use app\common\model_old\shop\User as UserOld;
use app\common\model\settings\Setting as SettingModel;
use app\common\model_old\settings\Setting as SettingOldModel;

// 应用公共文件

/**
 * 获取当前系统版本号
 * @return mixed|null
 * @throws Exception
 */
function get_version()
{
    try {
        $file = root_path() . '/version.json';
        $version = json_decode(file_get_contents($file), true);
        return $version['version'];
    } catch (\Exception $e) {
        return '';
    }
}

/**
 * 生成密码hash值
 * @param $password
 * @return string
 */
function salt_hash($password)
{
    return md5(md5($password) . 'jjjshop_salt_2020');
}

/**
 * 写入日志
 * @param $value
 * @param string $type
 */
function log_write($value, $channel = 'log')
{
    $msg = is_string($value) ? $value : var_export($value, true);
    trace($msg, $channel);
}

/**
 * 获取当前域名及根路径
 * @return string
 */
function base_url()
{
    static $baseUrl = '';
    if (empty($baseUrl)) {
        $request = Request::instance();
        $scheme = 'http';
        if (
            strstr($_SERVER['HTTP_ORIGIN'] ?? '', 'https:')
            || strstr($request->domain(), 'https:')
            || $request->scheme() == 'https'
            || (isset($_SERVER['SERVER_PORT']) && $_SERVER['SERVER_PORT'] == 443)
        ) {
            $scheme = 'https';
        }
        $baseUrl = $scheme . '://' . $request->host() . '/';
    }
    return $baseUrl;
}

// 生成验签
function signToken($uid, $source, $device_id = '', $pwd = '', $company_uuid = 0)
{
    $exp = ($source == 'admin' || $source == 'shop') ? 36000 : (30 * 86400 * 1200);
    $key = Config::get('app.salt');     //这里是自定义的一个随机字串，应该写在config文件中的，解密时也会用，相当    于加密中常用的 盐  salt
    $token = array(
        "iat" => time(),                    //签发时间
        "nbf" => time(),                    //在什么时候jwt开始生效  （这里表示生成100秒后才生效）
        "exp" => time() + $exp,             //token 过期时间改为1200个月 86400=24小时 36000=10小时
        'pwd' => $pwd,
        'source' => $source,
        'device_id' => $device_id . '',
        'staff_uuid' => $uid,
        'company_uuid' => $company_uuid,
    );
    $jwt = JWT::encode($token, $key, "HS256");  //根据参数生成了 token
    return $jwt;
}

// 验证token
/**
 * 验证JWT token的有效性和类型
 *
 * @param string $token JWT token字符串
 * @param string $type 需要验证的token类型
 * @return array 返回包含验证状态的数组
 */
function checkToken($token, $type)
{
    $key = Config::get('app.salt');
    $status = array("code" => -1);
    try {
        JWT::$leeway = 60; //当前时间减去60，把时间留点余地
        $decoded = JWT::decode($token, new Key($key, 'HS256')); //HS256方式，这里要和签发的时候对应
        $arr = json_decode(json_encode($decoded), 1);
        // 
        $res['code'] = 1;
        $res['data'] = [                             //记录的userid的信息，这里是自已添加上去的，如果有其它信息，可以再添加数组的键值对
            'uid' => $arr['staff_uuid'],
            'type' => $arr['source'],
            'pwd' => $arr['pwd'] ?? '',
            'device_id' => $arr['device_id'],
            'company_uuid' => $arr['company_uuid'],
        ];
        // 
        if ($res['data']['type'] != $type){
            $status['msg'] = "签名不正确";
            return $status;
        }
        return $res;
    } catch (\Firebase\JWT\SignatureInvalidException $e) { //签名不正确
        $status['msg'] = "签名不正确";
        return $status;
    } catch (\Firebase\JWT\BeforeValidException $e) { // 签名在某个时间点之后才能用
        $status['msg'] = "token失效";
        return $status;
    } catch (\Firebase\JWT\ExpiredException $e) { // token过期
        $status['msg'] = "登录过期";
        return $status;
    } catch (Exception $e) { //其他错误
        $status['msg'] = "未知错误";
        return $status;
    }
}

/**
 * 自动侦测设置获取语言选择
 * @return string
 */
function checkDetect(): string
{
    if ($langSet = request()->language) {
        if ($langSet == 'zh-tw') {
            $langSet = 'zhtw';
        }
        return $langSet;
    }

    // url中设置了语言变量
    if ($langSet = Request::get("language")) {
        return $langSet;
    }

    // Header中设置了语言变量
    if ($langSet = Request::header("language")) {
        return $langSet;
    }

    // Header中设置了语言变量
    if ($langSet = Request::header("Accept-Language")) {
        return $langSet;
    }

    // Cookie中设置了语言变量
    if ($langSet = Request::cookie("language")) {
        return $langSet;
    }

    // 自动侦测浏览器语言
    if ($langSet = Request::server('HTTP_ACCEPT_LANGUAGE')) {
        $langMap = [
            'zh-tw' => 'zhtw',
            'zh-cn' => 'zh',
            'en-us' => 'en',
        ];
        if (preg_match('/^([a-z\d\-]+)/i', $langSet, $matches)) {
            $langSet = strtolower($matches[1]);
            return $langMap[$langSet] ?? $langSet;
        }
        return '';
    }

    return 'en';
}


/**
 * 国际化（替换国际语言）
 * @param $val
 * @param string $language
 * @return mixed
 */
function __(string $val, string $language = 'zh'): string
{
    $language = checkDetect() ?: $language;
    // doto think-php 自带的语言体系
    // return lang('无', [], $language);
    //
    $data = [];
    $langpath = root_path() . 'lang/' . strtolower($language) . '/auto.php';
    if (file_exists($langpath)) {
        $data = include $langpath;
        if (!is_array($data)) {
            $data = [];
        }
    }
    if ($data && array_key_exists($val, $data)) {
        return $data[$val];
    }
    return $val;
}

/**
 * 打印文本
 * @return string
 */
function printText($leftText, $centerText = "", $rightText = "", $total = 32, $leftNum = 0, $centerNum = 0, $rightNum = 0, $intervalNum = 4)
{
    return PrintTexHelp::printText($leftText, $centerText, $rightText, $total, $leftNum, $centerNum, $rightNum, $intervalNum);
}

/**
 * 提取语言
 * @return string|array
 */
function extractLanguage($json)
{
    try {
        $lang = checkDetect();
        $texts = json_decode($json, true);
        if (!$texts) {
            return $json;
        }
        //
        $languages = getSettingLanguages(request()->shopSupplierId ?: 0);
        foreach ($languages as $language) {
            $name = $language['name'] ?? '';
            $name = $name == 'zhtw' ? 'zhtw' : $name;
            if ($name == $lang && array_key_exists($language['key'] ?? '', $texts)) {
                return $texts[$language['key']];
            }
        }
        //
        if (array_key_exists($lang, $texts)) {
            return $texts[$lang];
        }
        if ($languages && array_key_exists($languages[0]['key'] ?? '', $texts)) {
            return $texts[$languages[0]['key'] ?? ''];
        }
        if (array_key_exists("en", $texts)) {
            return $texts["en"];
        }
        if (array_key_exists("1", $texts)) {
            return $texts["1"];
        }
        //
        return  $json;
    } catch (\Throwable $th) {
        trace($th->getMessage());
        return $json;
    }
}

/**
 * 获取当前系统设置的语言
 */
function getSettingLanguages($companyUuid = 0)
{
    $connection = (new User())->getConnection();
    if (strlen($connection) == 14) {
        if ($companyUuid) {
            if (!($shopInfo = Cache::get('{common_get_settingLanguages}_' . 'common_shop_info' . $companyUuid)) || !Cache::get('{firstshop}_' . 'first_shop_info')) {
                $shopInfo = UserOld::getShopInfo('', true);
                Cache::tag('common_get_settingLanguages')->set('{common_get_settingLanguages}_' . 'common_shop_info' . $companyUuid, $shopInfo);
            }
        } else {
            $shopInfo = UserOld::getShopInfo('', true);
        }
        $companyUuid = $shopInfo['company_uuid'] ?? 0;
        if (!$languages = Cache::get('{common_get_settingLanguages}_' . 'common_setting_languages' . $companyUuid)) {
            $languages = SettingOldModel::getSupplierLanguage($companyUuid);
            Cache::tag('common_get_settingLanguages')->set('{common_get_settingLanguages}_' . 'common_setting_languages' . $companyUuid, $languages);
        }
        return $languages;
    } else {
        if ($companyUuid) {
            if (!($shopInfo = Cache::get('{common_get_settingLanguages}_' . 'common_shop_info' . $companyUuid)) || !Cache::get('{firstshop}_' . 'first_shop_info')) {
                $shopInfo = User::getShopInfo('', true);
                Cache::tag('common_get_settingLanguages')->set('{common_get_settingLanguages}_' . 'common_shop_info' . $companyUuid, $shopInfo);
            }
        } else {
            $shopInfo = User::getShopInfo('', true);
        }
        $companyUuid = $shopInfo['company_uuid'] ?? 0;
        if (!$languages = Cache::get('{common_get_settingLanguages}_' . 'common_setting_languages' . $companyUuid)) {
            $languages = SettingModel::getSupplierLanguage($companyUuid);
            Cache::tag('common_get_settingLanguages')->set('{common_get_settingLanguages}_' . 'common_setting_languages' . $companyUuid, $languages);
        }
        return $languages;
    }
}

/**
 * 生成雪花算法ID
 */
function createUuid()
{
    static $snowflake = null;
    if ($snowflake === null) {
        $snowflake = new SnowflakeHelp(1);
    }
    return $snowflake->next();
}

/**
 * 获取商家默认语言
 * @return string
 */
function getDefaultLanguage()
{
    static $defaultLanguage = null;
    if ($defaultLanguage !== null) {
        return $defaultLanguage;
    }
    $language = getSettingLanguages();
    $defaultLanguage = 'en'; // 默认值
    if (!empty($language)) {
        $keys = array_column($language, 'key');
        if (in_array('zh', $keys)) {
            $defaultLanguage = 'zh';
        } else {
            $defaultLanguage = $language[0]['key'];
        }
    }
    return $defaultLanguage;
}

/**
 * php-fpm 响应完成, 关闭连接
 */
if (!function_exists("fastcgi_finish_request")) {
    function fastcgi_finish_request() {}
}
