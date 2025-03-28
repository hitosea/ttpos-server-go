<?php

namespace app\common\service\qrcode;

use help\StringHelp;
use Endroid\QrCode\QrCode;
use Endroid\QrCode\Writer\PngWriter;
use app\common\service\qrcode\AuthService;
use app\common\model\store\Table as TableModel;

/**
 * 二维码
 */
class TableService extends Base
{
    private $id;
    private $source;
    private $action;
    private $base_url;

    /**
     * 构造方法
     */
    public function __construct($id, $source, $action = 'get')
    {
        parent::__construct();
        $this->id = $id;
        $this->source = $source;
        $this->action = $action;
        $this->base_url = env('H5_BASE_URL') ?: base_url();
    }

    /**
     * 桌台二维码
     */
    public function getImage()
    {
        $companyUuid = request()->appId;
        $table = TableModel::detail($this->id);
        $qrCodeValue = $this->action == 'update' ? StringHelp::generatePassword(6, 1) : $table['qrcode_token'];
        $arr = ['a' => $companyUuid, 't' => $this->id, 'q' => $qrCodeValue];
        $auth = (new AuthService);
        $token = $auth->generateToken($arr);
        if ($this->source == 'wx') {
            $qrCode = new QrCode($this->base_url . "pages/product/share-login/#/?token={$token}");
        } else if ($this->source == 'mp' || $this->source == 'h5') {
            $qrCode = new QrCode($this->base_url . "/#/home?token={$token}");
        }
        $qrCode->setSize(300);
        $qrCode->setMargin(10);
        if ($this->action == 'update') {
            $table->save(['qrcode_token' => $qrCodeValue]);
        }
        //
        return (new PngWriter())->write($qrCode)->getDataUri();
    }

    /**
     * 批量生成桌台二维码
     *
     * @param array $tableIds 桌台 ID 数组
     * @return array 返回包含每个桌台 ID、名称和二维码的数组
     */
    public function generateBatchQrCodes(array $tableIds)
    {
        $qrCodeData = [];
        $auth = new AuthService;
        $companyUuid = request()->appId;
        //
        $tables = TableModel::where('table_id', 'in', $tableIds)->select()->toArray();
        foreach ($tables as $table) {
            $id = $table['table_id'];
            $qrCodeValue = $this->action == 'update' ? StringHelp::generatePassword(6, 1) : $table['qrcode_token'];
            $arr = ['a' => $companyUuid, 't' => $id, 'q' => $qrCodeValue];
            $token = $auth->generateToken($arr);
            if ($this->source == 'wx') {
                $qrCode = new QrCode($this->base_url . "pages/product/share-login/#/?token={$token}");
            } else if ($this->source == 'mp' || $this->source == 'h5') {
                $qrCode = new QrCode($this->base_url . "scan/#/?token={$token}");
            }
            $qrCode->setSize(300);
            $qrCode->setMargin(10);
            if ($this->action == 'update') {
                $table->save(['qrcode_token' => $qrCodeValue]);
            }
            $qrCodeData[] = [
                'table_id' => $id,
                'table_no' => $table['table_no'],
                'qrcode' => (new PngWriter())->write($qrCode)->getDataUri()
            ];
        }
        return $qrCodeData;
    }

    /**
     * 下载桌台二维码
     */
    public function getDownload()
    {
        $table = TableModel::detail($this->id);
        $companyUuid = request()->appId;
        // 保存目录
        $savePath = $this->getPosterPath($table['app_id']);

        // 删除目录下的文件
        if (!$this->is_empty_dir($savePath)) {
            $this->deleteDir(substr($savePath, 0, -1));
        }

        // 创建目录
        if (!is_dir($savePath)) {
            mkdir($savePath, 0755, true);
        }

        // 下载二维码
        if ($this->source == 'wx') {
            $this->saveQrcodeToDir($table['app_id'], 'pages/product/share-login', $savePath, $this->id, $table['shop_supplier_id']);
        } else if ($this->source == 'mp' || $this->source == 'h5') {
            $arr = ['a' => $companyUuid, 't' => $this->id, 'q' => $table['qrcode_token']];
            $auth = (new AuthService);
            $token = $auth->generateToken($arr);
            $this->saveMpQrcodeToDir("scan/#/?token={$token}", $savePath);
        }

        // 打开或创建压缩文件
        $zipNameUrl = $this->getZipPath($table['app_id']);
        $zip = new \ZipArchive();
        if ($zip->open($zipNameUrl, \ZipArchive::OVERWRITE) !== TRUE) {
            //OVERWRITE 参数会覆写压缩包的文件 文件必须已经存在
            if ($zip->open($zipNameUrl, \ZipArchive::CREATE) !== true) {
                // 文件不存在则生成一个新的文件 用CREATE打开文件会追加内容至zip
                return '下载失败，文件夹不存在';
            }
        }

        $this->addFileToZip($savePath, $zip); //调用方法，对要打包的根目录进行操作，并将ZipArchive的对象传递给方法
        $zip->close(); //关闭处理的zip文件

        // 检查文件是否存在
        if (!file_exists($zipNameUrl)) {
            return '下载失败，压缩文件不存在';
        }

        // 设置文件头信息并输出文件
        $zipName = $this->source . '.zip';
        header('Content-Type: application/zip');
        header('Content-disposition: attachment; filename=' . $zipName);
        header('Content-Length: ' . filesize($zipNameUrl));
        ob_clean();
        flush();
        readfile($zipNameUrl);
        exit;
    }


    /**
     * 二维码文件路径
     */
    private function getPosterPath($app_id)
    {
        // 保存路径
        $tempPath = root_path('public') . 'temp' . '/' . $app_id . '/table-' . $this->id . '/' . $this->source . '/';
        return $tempPath;
    }

    /**
     * 二维码文件路径
     */
    private function getZipPath($app_id)
    {
        // 保存路径
        $tempPath = root_path('public') . 'temp' . '/' . $app_id . '/table-' . $this->id . '/' . $this->source . '.zip';
        return $tempPath;
    }

    /**
     * 删除当前目录及其目录下的所有目录和文件
     * @param string $path 待删除的目录
     * @note  $path路径结尾不要有斜杠/(例如:正确[$path='./static/image'],错误[$path='./static/image/'])
     */
    private function deleteDir($path)
    {
        if (is_dir($path)) {
            //扫描一个目录内的所有目录和文件并返回数组
            $dirs = scandir($path);
            foreach ($dirs as $dir) {
                //排除目录中的当前目录(.)和上一级目录(..)
                if ($dir != '.' && $dir != '..') {
                    //如果是目录则递归子目录，继续操作
                    $sonDir = $path . '/' . $dir;
                    if (is_dir($sonDir)) {
                        //递归删除
                        $this->deleteDir($sonDir);
                        //目录内的子目录和文件删除后删除空目录
                        @rmdir($sonDir);
                    } else {
                        //如果是文件直接删除
                        @unlink($sonDir);
                    }
                }
            }
            @rmdir($path);
        }
    }

    /**
     * 打包文件夹
     */
    private function addFileToZip($path, $zip)
    {
        $handler = opendir($path);
        while (($filename = readdir($handler)) !== false) {
            if ($filename != "." && $filename != "..") {
                if (is_dir($path . "/" . $filename)) {
                    $this->addFileToZip($path . "/" . $filename, $zip);
                } else { //将文件加入zip对象
                    $zip->addFile($path . "/" . $filename);
                    $zip->renameName($path . "/" . $filename, $filename);
                }
            }
        }
        @closedir($handler);
    }

    private function is_empty_dir($fp)
    {
        if (!file_exists($fp)) {
            return false;
        }
        $H = @opendir($fp);
        $i = 0;
        while ($_file = readdir($H)) {
            $i++;
        }
        closedir($H);
        if ($i > 2) {
            return false;
        } else {
            return true;
        }
    }
}
