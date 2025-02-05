<?php

namespace help;


class FileHelp
{

    /**
     * 获取目录下的文件列表
     * @param $dir
     * @return array|string
     */
    public static function getFilesInDir($dir): array|string
    {
        $files = [];
        $dh = opendir($dir);
        while (($file = readdir($dh)) !== false) {
            if ($file != '.' && $file != '..') {
                $filePath = $dir . '/' . $file;
                if (is_dir($filePath)) {
                    $files = array_merge($files, self::getFilesInDir($filePath));
                } else {
                    $files[] = $filePath;
                }
            }
        }
        closedir($dh);
        return $files;
    }

}
