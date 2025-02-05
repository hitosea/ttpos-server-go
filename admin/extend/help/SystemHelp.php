<?php

namespace help;

use function exec;
use think\Exception;
use think\facade\Log;
use think\facade\Request;

// 执行本地cmd 			SystemHelp::cmd($cmd);
// 执行远程cmd命令 		 SystemHelp::sshCmd($host,$user,$password,$cmd);

class SystemHelp
{
    /**
     * @param $command
     * @return array
     * @throws Exception
     */
    public static function exec($command): array
    {
        exec($command, $output, $ret);
        if ($ret) {
            $output = implode("\n", $output);
            throw new Exception("System call failed.\nCommand: {$command}\n\n{$output}");
        }
        return $output;
    }

    /**
     * @param $command
     * @param int $retryNum
     * @return string|null
     */
    public static function shot($command, $retryNum = 5): ?string
    {
        try {
            $ret = self::exec($command);
        } catch (Exception $errA) {
            // 服务重启，循环5次，每次等待5秒后重新获取
            if ($retryNum > 0) {
                for ($i = 1; $i < $retryNum; $i++) {
                    sleep(1);
                    try {
                        $ret = self::exec($command);
                        break;
                    } catch (Exception $errB) {
                        if ($i == 4) {
                            $ret = [$errB->getMessage()];
                            Log::info($errB->getMessage());
                        }
                    }
                }
            } else {
                $ret = [$errA->getMessage()];
                Log::info($errA->getMessage());
            }
        }
        return $ret[0] ?? null;
    }

    /**
     * @param $command
     * @param $appendLog
     * @return string|null
     */
    public static function cmd($command, $appendLog = false): ?string
    {
        if ($appendLog === true) {
            Log::info($command);
        }
        return self::shot($command, 1);
    }

    /**
     * 执行远程cmd命令
     *
     * @param [type] $host
     * @param [type] $user
     * @param [type] $password
     * @param [type] $cmd
     * @return void
     */
    public static function sshCmd($host, $user, $password, $cmd)
    {
        $ret = '';
        try {
            self::exec("sshpass -p $password ssh -o StrictHostKeyChecking=no $user@$host '$cmd'");
        } catch (Exception $errA) {
            $ret = 'system call failed';
        }
        return $ret;
    }

    /**
     * 判断浏览器名称和版本
     */
    public static function getClientBrowser()
    {
        if (empty($_SERVER['HTTP_USER_AGENT'])) {
            return 'robot！';
        }
        if ((false == strpos($_SERVER['HTTP_USER_AGENT'], 'MSIE')) && (strpos($_SERVER['HTTP_USER_AGENT'], 'Trident') !== FALSE)) {
            return 'Internet Explorer 11.0';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'MSIE 10.0')) {
            return 'Internet Explorer 10.0';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'MSIE 9.0')) {
            return 'Internet Explorer 9.0';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'MSIE 8.0')) {
            return 'Internet Explorer 8.0';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'MSIE 7.0')) {
            return 'Internet Explorer 7.0';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'MSIE 6.0')) {
            return 'Internet Explorer 6.0';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'Edge')) {
            return 'Edge';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'Firefox')) {
            return 'Firefox';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'Chrome')) {
            return 'Chrome';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'Safari')) {
            return 'Safari';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'Opera')) {
            return 'Opera';
        }
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], '360SE')) {
            return '360SE';
        }
        //微信浏览器
        if (false !== strpos($_SERVER['HTTP_USER_AGENT'], 'MicroMessage')) {
            return 'MicroMessage';
        }
        return '';
    }

    /**
     * 判断是否为App端访问
     */
    public static function getPlatform()
    {
        $userAgent = Request::header('User-Agent');
        if (strpos($userAgent, 'Mobile') !== false) {
            return 'Mobile';
        }
        if (strpos($userAgent, 'Android') !== false) {
            return 'Android';
        }
        if (strpos($userAgent, 'iPhone') !== false) {
            return 'iPhone';
        }
        $platform = Request::header('platform');
        return empty($platform) ? 'Web' : $platform;
    }
}
