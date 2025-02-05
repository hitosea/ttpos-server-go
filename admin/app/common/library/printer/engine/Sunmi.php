<?php

namespace app\common\library\printer\engine;

use help\HttpHelp;
use app\common\library\printer\party\SunmiCloudPrinter;

/**
 * 商米
 */
class Sunmi extends Basics
{
    // 接口路径
    const PATH = '/cgi-bin/print.cgi';

    /**
     * 执行订单打印
     */
    public function printTicket($content, $printMethod = 1)
    {
        $config = json_decode($this->config, true);
        // 云打印
        if ($this->printer->printer_type['value'] == 'SUNMI_CLOUD') {
            $printer = new SunmiCloudPrinter(567, $config['APP_ID'], $config['APP_KEY']);
            $printer->orderData = $content;
            $res = $printer->pushContent($config['SN'], time() . substr(implode('', array_map('ord', str_split(substr(uniqid(), 7, 13), 1))), 0, 8));
            if (!$res) {
                $this->error = "请求错误";
                return false;
            }
            $code = $res['code'] ?? 0;
            if ($code == 1 || $code == 10000) {
                return true;
            }
            switch ($code) {
                case '20000':
                    $this->error = "网关校验缺少必要参数";
                    break;
                case '20001':
                    $this->error = "请求超过有效期";
                    break;
                case '30000':
                    $this->error = "开发者身份验证失败（APPID无效；访问IP不在配置的IP白名单列表）";
                    break;
                case '30001':
                    $this->error = "用户缺少相关权限（能力未关联）";
                    break;
                case '40000':
                    $this->error = "签名验证失败";
                    break;
                case '50000':
                    $this->error = "服务器异常";
                    break;
                case '50001':
                    $this->error = "网关异常";
                    break;
                case '10071400':
                    $this->error = "参数错误，请注意参数类型是否错误";
                    break;
                case '10071500':
                    $this->error = "服务器错误";
                    break;
                case '10071701':
                    $this->error = "未知设备，需要检查SN号是否错误";
                    break;
                case '10071702':
                    $this->error = "设备已绑定";
                    break;
                case '10071703':
                    $this->error = "设备绑定异常";
                    break;
                case '10071704':
                    $this->error = "该SN号设备不属于本渠道，需要联系销售绑定到贵司渠道下";
                    break;
                case '10071705':
                    $this->error = "订单已推送，需要确认订单号是否重复";
                    break;
                case '10071706':
                    $this->error = "订单未知";
                    break;
                case '10071707':
                    $this->error = "没有设备或不是此通道";
                    break;
                default:
                    $this->error = $code . ' : ' . ($res['msg'] ?? '异常错误');
                    break;
            }
            return false;
        }
        // 局域网打印
        $url = 'http://' . $config['IP'] . self::PATH . "?sn={$config['SN']}&copies=1";
        $res = HttpHelp::postRequest($url, $content, ['Content-Type: text/plain; charset=uft-8'], 2);
        if (!$res) {
            return false;
        }
        return $res;
    }
}
