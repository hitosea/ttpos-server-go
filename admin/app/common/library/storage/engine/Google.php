<?php

namespace app\common\library\storage\engine;

use image\Image;
use think\Exception;
use Google\Cloud\Storage\StorageClient;

/**
 * 谷歌云存储引擎
 */
class Google extends Server
{
    private $config;

    private $credentialsPath;

    private $urlParam;

    /**
     * 构造方法
     */
    public function __construct($config)
    {
        parent::__construct();
        $this->config = $config;
        // 
        $this->credentialsPath = '/var/certificate/' . $config['credentials_file'];
        if (!file_exists($this->credentialsPath)) {
            throw new Exception('storage下未找到云存储配置文件');
        }
        //
        putenv("GOOGLE_APPLICATION_CREDENTIALS=" . $this->credentialsPath);
    }

    /**
     * 执行上传
     */
    public function upload($thumb=0)
    {
        // 要上传图片的本地路径
        $realPath = $this->getRealPath();
        // 创建存储客户端
        $storage = new StorageClient(['keyFilePath' => $this->credentialsPath]);
        // 获取存储桶
        $bucket = $storage->bucket($this->config['bucket']);
        // 上传目录
        $path = (request()->appId ? 'shop' . request()->appId : 'saas') . '/' . date('Ymd');
        // 上传文件
        $saveName = $path .'/'. $this->fileName;
        // 图片上传
        $fileMimeType = $this->file->getMime();
        if ($thumb && strpos($fileMimeType, 'image') !== false) {
            /** @var \GdImage $image */
            $image = Image::open($this->file)->thumb($thumb, $thumb, Image::THUMB_SCALING)->getImg();
            ob_start();
            imagepng($image);
            $imageData = ob_get_contents();
            ob_end_clean();
            imagedestroy($image);
        } else {
            $imageData = fopen($realPath, 'r');
        }
        // 上传本地文件到 Google Cloud Storage
        $object = $bucket->upload($imageData, [
            'name' => $this->config['uploads_catalogue'] . '/' . $saveName
        ]);
        // 检查上传结果
        if (!$object) {
            $this->error = '上传失败';
            return false;
        }
        // 
        $this->urlParam = explode('?', $object->signedUrl(new \DateTime('+100 years')))[1];
        // 
        return $saveName;
    }

    /**
     * 删除文件
     */
    public function delete($fileName)
    {
        // 创建存储客户端
        $storage = new StorageClient(['keyFilePath' => $this->credentialsPath]);
        // 获取存储桶
        $bucket = $storage->bucket($this->config['bucket']);
        // 
        $bucket = $storage->bucket($bucket);
        $object = $bucket->object($this->config['uploads_catalogue'] . '/' . $fileName);
        $object->delete();
        // 
        return true;
    }

    /**
     * 返回文件路径
     */
    public function getFileName()
    {
        return $this->fileName;
    }

    /**
     * 返回参数
     */
    public function getUrlParam()
    {
        return $this->urlParam;
    }
}
