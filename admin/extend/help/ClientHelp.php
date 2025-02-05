<?php

namespace help;

class ClientHelp
{
    /**
     * 验证客户端版本
     * @return mixed
     */
    public static function verifyClientVersion($ver, $d = '<')
    {
        $version = request()->header('Version-Name') ?: '0.0.0';
        if ($version && version_compare($version, $ver, $d)) {
            return true;
        } else {
            return false; 
        }
    }

}
