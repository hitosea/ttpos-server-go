<?php

namespace help;

// 磁盘的帮助			     
class DiskHelp
{
    /**
    * 服务器存储空间信息
    *
    * @return array
    */
    public static function getDiskSpaceInfo()
    {
        $diskTotalSpace = disk_total_space("/") / 1073741824;
        $diskFreeSpace = disk_free_space("/") / 1073741824;
        $diskUsedSpace = $diskTotalSpace - $diskFreeSpace;
        $diskUsedPercentage = ($diskUsedSpace / $diskTotalSpace) * 100;

        return [
            'total_space' => round($diskTotalSpace, 1), // 总大小
            'free_space' => round($diskFreeSpace, 1), // 可用大小
            'used_space' => round($diskUsedSpace, 1), // 已用大小
            'used_percentage' => round($diskUsedPercentage, 0), // 已用百分比
        ];
    }

    /**
     * 获取当前设备机器码
     */
    public static function getMachineCode()
    {
        $machineCode = cache('machine_code');
        if (!$machineCode) {
            $machineCode = StringHelp::getGuidV4();
            cache('machine_code', $machineCode, 86400);
        }
        return $machineCode;
    }

}
