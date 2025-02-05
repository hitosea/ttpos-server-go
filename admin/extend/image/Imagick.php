<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK IT ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2015 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: yunwuxin <448901948@qq.com>
// +----------------------------------------------------------------------

namespace image;

class Imagick
{
    /**
     * 压缩图片
     */
    public static function thumb($path, $w, $h, $savePath)
    {
        // 创建一个 ImageMagick 对象
        $image = new \Imagick($path);
        // 调整图像大小
        $image->resizeImage($w, $h, \Imagick::FILTER_LANCZOS, 1, true);
        // 设置图像压缩类型
        $image->setImageCompression(\Imagick::COMPRESSION_JPEG);
        $image->setImageCompressionQuality(100); // 设置为100表示无损压缩
        // 保存无损压缩后的图像
        $image->writeImage($savePath);
        // 释放内存
        $image->clear();
        $image->destroy();
    }
}
