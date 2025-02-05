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

namespace help;

class ImageHelp
{

    /**
     * 确保是白底黑字
     * @param \SplFileInfo|string $file
     * @return Image
     */
    public static function whiteBackgroundWithBlackText($imgPath, $imgSavePath)
    {
        $image = new \Imagick($imgPath);

        // 处理透明背景
        if ($image->getImageAlphaChannel()) {
            // 创建白色背景并合并
            $white = new \Imagick();
            $white->newImage($image->getImageWidth(), $image->getImageHeight(), 'white');
            $white->compositeImage($image, \Imagick::COMPOSITE_OVER, 0, 0);
            $image = $white;
        }
       
        // 转换为灰度图像
        $image->setImageType(\Imagick::IMGTYPE_GRAYSCALE);
       
        // 增强对比度
        $image->contrastImage(1);
       
        // 去除噪点（使用中值滤波）
        $image->adaptiveBlurImage(3, 0);
       
        // 计算平均亮度
        $mean = $image->getImageChannelMean(\Imagick::CHANNEL_GRAY);
        $threshold = $mean['mean'] * 0.95;
       
        // 二值化处理
        $image->thresholdImage($threshold);
       
        // 确保是白底黑字
        $newMean = $image->getImageChannelMean(\Imagick::CHANNEL_GRAY);
        if ($newMean['mean'] < ($image->getQuantumRange()['quantumRangeLong'] / 2)) {
            $image->negateImage(false);
        }
       
        // 最终的清理（再次使用中值滤波）
        $image->adaptiveBlurImage(3, 0);
   
        // 保存文件
        $image->setImageFormat('png');
        $image->setImageCompressionQuality(95);
        
        // 判断目录不存在则添加
        if (!file_exists(dirname($imgSavePath))) {
            mkdir(dirname($imgSavePath), 0777, true);
        }
        
        $image->writeImage($imgSavePath);
       
        // 清理资源
        $image->clear();
        $image->destroy();
    }

}
