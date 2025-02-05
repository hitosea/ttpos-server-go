<?php

namespace app\common\library\storage;

use think\Exception;

/**
 * 存储模块驱动
 */
class Driver
{
    private $config;    // upload 配置
    private $engine;    // 当前存储引擎类

    /**
     * 构造方法
     */
    public function __construct($config, $storage = null)
    {
        if ($config['default'] == 'google') {
            if (!($config['engine']['google']['credentials_file'] ?? '') || !($config['engine']['google']['bucket'] ?? '') || !($config['engine']['google']['uploads_catalogue'] ?? '')) {
                throw new Exception('未找到云存储环境变量配置');
            }
        }
        // 保存配置
        $this->config = $config;
        // 实例化当前存储引擎
        $this->engine = $this->getEngineClass($storage);
    }

    /**
     * 设置上传的文件信息
     */
    public function setUploadFile($name = 'iFile')
    {
        return $this->engine->setUploadFile($name);
    }

    /**
     * 设置服务器本地的文件信息
     */
    public function setLocalFile($file)
    {
        return $this->engine->setLocalFile($file);
    }

    /**
     * 设置上传的文件信息
     */
    public function setUploadFileByReal($filePath)
    {
        return $this->engine->setUploadFileByReal($filePath);
    }

    /**
     * 执行文件上传
     */
    public function upload($thumb = 0)
    {
        return $this->engine->upload($thumb);
    }

    /**
     * 执行文件删除
     */
    public function delete($fileName)
    {
        return $this->engine->delete($fileName);
    }

    /**
     * 获取错误信息
     */
    public function getError()
    {
        return $this->engine->getError();
    }

    /**
     * 获取文件路径
     */
    public function getFileName()
    {
        return $this->engine->getFileName();
    }

    /**
     * 获取访问参数
     */
    public function getUrlParam()
    {
        return $this->engine->getUrlParam();
    }

    /**
     * 返回文件信息
     */
    public function getFileInfo()
    {
        return $this->engine->getFileInfo();
    }

    /**
     * 获取当前的存储引擎
     */
    private function getEngineClass($storage = null)
    { 
        $engineName = is_null($storage) ? ($this->config['default'] ?? 'local') : $storage;
        $classSpace = __NAMESPACE__ . '\\engine\\' . ucfirst($engineName);
        if (!class_exists($classSpace)) {
            throw new Exception('未找到存储引擎类: ' . $engineName);
        }
        return new $classSpace($this->config['engine'][$engineName]);
    }

}
