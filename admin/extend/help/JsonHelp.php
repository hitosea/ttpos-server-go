<?php

namespace help;

// json帮助
class JsonHelp
{
    /**
     * json 转换true,false
     */
    public static function jsonRecursive(&$array)
    {
        foreach ($array as $key => $value) {
            if (is_array($value)) {
                self::jsonRecursive($array[$key]);
            } else {
                if ($value === 'true') {
                    $array[$key] = true;
                } else if ($value === 'false') {
                    $array[$key] = false;
                }
            }
        }
    }

}
