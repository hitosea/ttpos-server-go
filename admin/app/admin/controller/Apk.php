<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use Google\Cloud\Storage\StorageClient;

/**
 * 安装包资源
 * @Apidoc\Group("base")
 * @Apidoc\Sort(1)
 */
class Apk extends Controller
{
    /**
     * @Apidoc\Title("下载最新包")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/apk/download")
     * @Apidoc\Param("p", type="string", require=false, desc="平台 如：TTPOS")
     * @Apidoc\Param("e", type="string", require=false, desc="环境 如：Prod Demo Test")
     * @Apidoc\Param("v", type="string", require=false, desc="版本号 如：1.0.8")
     * @Apidoc\Param("t", type="string", require=false, desc="设备 Assistant-助手 Cashier-收银 Kitchen-厨显 Menu-平板 Shop-商家后台")
     */
    public function download()
    {
        $param = $this->getData();
        $platform = $param['p'] ?? '';
        $env = $param['e'] ?? '';
        $version = $param['v'] ?? '';
        $terminal = $param['t'] ?? '';
        //
        $credential_path = __DIR__ .'/../../../runtime/storage/' . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        if (!file_exists($credential_path)) {
            $credential_path = '/var/certificate/' . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
            if (!file_exists($credential_path)) {
                echo "Not configured.";
                exit;
            }
        }
        //
        putenv("GOOGLE_APPLICATION_CREDENTIALS=" . $credential_path);
        $bucket_name = env('GOOGLE_APPLICATION_BUCKET_NAME');
        // 流包装
        $storage = new StorageClient();
        $storage->registerStreamWrapper();
        if ($env == 'Demo') {
            $file_name = "{$platform}-{$env}-{$terminal}-V{$version}.apk";
        } else {
            $file_name = "{$platform}-{$terminal}-V{$version}.apk";
        }
        $gs_path = "gs://{$bucket_name}/{$platform}/{$env}/{$terminal}/{$file_name}";
        // 检查文件是否存在
        if (file_exists($gs_path)) {
            // 设置HTTP头部信息
            header('Content-Description: File Transfer');
            header('Content-Type: application/octet-stream');
            header('Content-Disposition: attachment; filename="'.basename($file_name).'"');
            header('Expires: 0');
            header('Cache-Control: must-revalidate');
            header('Pragma: public');
            header('Content-Length: ' . filesize($gs_path));
            // 清除缓冲区并关闭输出缓冲
            ob_clean();
            flush();
            // 读取文件内容并输出
            readfile($gs_path);
            exit;
        } else {
            echo "File not found.";
        }
    }

    /**
     * @Apidoc\Title("预下载检查安装包是否存在")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/apk/preDownload")
     * @Apidoc\Param("p", type="string", require=false, desc="平台 如：TTPOS")
     * @Apidoc\Param("e", type="string", require=false, desc="环境 如：Prod Demo Test")
     * @Apidoc\Param("v", type="string", require=false, desc="版本号 如：1.0.8")
     * @Apidoc\Param("t", type="string", require=false, desc="设备 Assistant-助手 Cashier-收银 Kitchen-厨显 Menu-平板 Shop-商家后台")
     */
    public function preDownload()
    {
        $param = $this->getData();
        $platform = $param['p'] ?? '';
        $env = $param['e'] ?? '';
        $version = $param['v'] ?? '';
        $terminal = $param['t'] ?? '';
        //
        $credential_path = __DIR__ .'/../../../runtime/storage/' . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        if (!file_exists($credential_path)) {
            $credential_path = '/var/certificate/' . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
            if (!file_exists($credential_path)) {
                echo "Not configured.";
                exit;
            }
        }
        //
        putenv("GOOGLE_APPLICATION_CREDENTIALS=" . $credential_path);
        $bucket_name = env('GOOGLE_APPLICATION_BUCKET_NAME');
        // 流包装
        $storage = new StorageClient();
        $storage->registerStreamWrapper();
        if ($env == 'Demo') {
            $file_name = "{$platform}-{$env}-{$terminal}-V{$version}.apk";
        } else {
            $file_name = "{$platform}-{$terminal}-V{$version}.apk";
        }
        $gs_path = "gs://{$bucket_name}/{$platform}/{$env}/{$terminal}/{$file_name}";
        // 检查文件是否存在
        if (file_exists($gs_path)) {
            return $this->renderSuccess();
        } else {
            return $this->renderError();
        }
    }
}
