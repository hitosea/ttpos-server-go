<?php

namespace help;

use Exception;
use think\facade\Cache;
use Endroid\QrCode\QrCode;
use Endroid\QrCode\Color\Color;
use Endroid\QrCode\Writer\PngWriter;
use Endroid\QrCode\Encoding\Encoding;
use Endroid\QrCode\RoundBlockSizeMode\RoundBlockSizeModeMargin;
use Endroid\QrCode\ErrorCorrectionLevel\ErrorCorrectionLevelHigh;

class ImgHelp
{

     /**
     * @param string $data 内容
     * @param int $size 尺寸
     * @param int $margin 外边距
     * @return string
     * @throws Exception
     */
    public static function qrcode($data, $size = 300, $margin = 0)
    {
        $writer = new PngWriter();
        $qrCode = QrCode::create($data)
            ->setEncoding(new Encoding('UTF-8'))
            ->setErrorCorrectionLevel(new ErrorCorrectionLevelHigh()) // 错误矫正：高级
            ->setSize($size)
            ->setMargin($margin)
            ->setRoundBlockSizeMode(new RoundBlockSizeModeMargin())
            ->setForegroundColor(new Color(0, 0, 0)) // 前景黑色
            ->setBackgroundColor(new Color(255, 255, 255)); // 背景白色
        return $writer->write($qrCode)->getDataUri(); // 返回base64图片
    }

    /**
     * 把图片绝对路径域名替换
     * @throws Exception
     */
    public static function addImageDomain($imageUrl, $is = true)
    {
        if (!$imageUrl) {
            return '';
        }
        if (strstr($imageUrl,'http://qn-cdn.jjjshop.net')) {
            return $imageUrl;
        }
        $urlComponents = parse_url($imageUrl);
        if (isset($urlComponents['path']) && is_string($urlComponents['path']) && $urlComponents['path'] !== '' && $urlComponents['path'][0] !== '/') {
            $urlComponents['host'] = base_url();
        } else {
            $urlComponents['host'] = rtrim(base_url(), '/');
        }
        $newUrl = $is ? $urlComponents['host'] : '';
        if (isset($urlComponents['path'])) {
            $newUrl .= $urlComponents['path'];
        }
        if (isset($urlComponents['query'])) {
            $newUrl .= '?' . $urlComponents['query'];
        }
        if (isset($urlComponents['fragment'])) {
            $newUrl .= '#' . $urlComponents['fragment'];
        }
        return $newUrl;
    }

    /**
     * 把图片路径域名去掉，只保留相对路径
     * @throws Exception
     */
    public static function removeImageDomain($imageUrl)
    {
        $urlComponents = parse_url($imageUrl);
        // 重新构建URL
        $newUrl = '';
        if (isset($urlComponents['path'])) {
            $newUrl .= $urlComponents['path'];
        }
        if (isset($urlComponents['query'])) {
            $newUrl .= '?' . $urlComponents['query'];
        }
        if (isset($urlComponents['fragment'])) {
            $newUrl .= '#' . $urlComponents['fragment'];
        }
        return $newUrl;
    }

    /**
     * 下载图片到本地
     *
     * @param string $url 图片的URL
     * @param string $path 图片保存的本地路径，默认为'./uploads/cloud/'
     * @return string|bool
     */
    public static function downloadCloudImage($url, $path = './uploads/cloud/')
    {
        $pathInfo = pathinfo($url);
        $filename = $pathInfo['basename'];
        $filename = explode('?', $filename)[0] ?? '';
        if (!$filename) {
            return '';
        }
        //
        $filenameCacheKey = 'sync_downloadeds_' . md5($filename);
        $downloadedFilenamePath = Cache::get($filenameCacheKey);
        if ($downloadedFilenamePath) {
            return $downloadedFilenamePath;
        }
        try {
            $imageData = file_get_contents($url);
            if ($imageData !== false) {
                if (!file_exists($path)) {
                    mkdir($path, 0777, true);
                }
                $localFilePath = $path . $filename;
                //
                if (file_put_contents($localFilePath, $imageData) !== false) {
                    $relativePath = str_replace('public', '', $localFilePath);
                    $relativePath = preg_replace('/\/+/', '/', ltrim($relativePath, '.'));
                    Cache::set($filenameCacheKey, $relativePath);
                    return $relativePath;
                } else {
                    return '';
                }
            } else {
                return '';
            }
        } catch (\Throwable $th) {
            return '';
        }
    }
}
