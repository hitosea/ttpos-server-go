<?php

declare(strict_types=1);

namespace app\common\command;

use think\facade\Db;
use think\console\Input;
use think\console\Output;
use think\console\Command;
use app\common\model\file\UploadFile;

// 处理上传文件到谷歌云
// ./cmd think upload-file-to-google
class UploadFileToGoogle extends Command
{

    protected function configure()
    {
        // 指令配置
        $this->setName('upload-file-to-google')->setDescription('处理上传文件到谷歌云');
    }

    protected function execute(Input $input, Output $output)
    {
        $credentialsPath = runtime_path('storage') . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        $bucket = env('GOOGLE_APPLICATION_UPLOADS_BUCKET_NAME');
        $catalogue = env('GOOGLE_APPLICATION_UPLOADS_CATALOGUE_NAME');
        if (!file_exists($credentialsPath) || !$bucket || !$catalogue) {
            echo "Not configured.";
            exit;
        }
        //
        $storage = new \Google\Cloud\Storage\StorageClient(['keyFilePath' => $credentialsPath]);

        // admin
        dump("---------------------------------------------------");
        dump("------------------处理平台端-----------------------");
        dump("---------------------------------------------------");
        $uploadFiles = Db::name('upload_file')
            ->where('storage', 'local')
            ->where(function ($query) {
                $query->whereNull('url_param')->whereOr('url_param', '');
            })
            ->where('is_delete', 0)
            ->select();

        // 批量上传文件到谷歌云
        foreach ($uploadFiles as $file) {
            $imageSrc = root_path('public/uploads') . $file['save_name'];
            $objectPath = $catalogue . '/' . $file['save_name'];
            $object = $storage->bucket($bucket)->object($objectPath);
            $url = $object->signedUrl(new \DateTime('+100 years'));
            //
            if (!$object->exists()) {
                if (file_exists($imageSrc)) {
                    $storage->bucket($bucket)->upload(file_get_contents($imageSrc), [
                        'name' => $objectPath,
                        'predefinedAcl' => 'publicRead' // 设置为公开访问
                    ]);
                    // 更新数据库中的存储类型为google
                    Db::name('upload_file')->where('file_id', $file['file_id'])->update([
                        'url_param' =>  explode('?', $url)[1],
                        'file_url' => "https://storage.googleapis.com/$bucket/$catalogue",
                    ]);
                    //
                    dump("文件处理成功: " . $imageSrc);
                }
            } else {
                // 更新数据库中的存储类型为google
                Db::name('upload_file')->where('file_id', $file['file_id'])->update([
                    'url_param' =>  explode('?', $url)[1],
                    'file_url' => "https://storage.googleapis.com/$bucket/$catalogue",
                ]);
                //
                dump("文件处理成功: " . $imageSrc);
            }
        }

        // 处理商户
        $apps = Db::name('app')->where('is_delete', 0)->select();
        foreach ($apps as $app) {
            request()->appId = $appId = $app['app_id'];
            dump("---------------------------------------------------");
            dump("---------------处理商家$appId------------------");
            dump("---------------------------------------------------");
            //
            $uploadFiles = (new UploadFile([], $appId))->where('storage', 'local')
                ->where(function ($query) {
                    $query->whereNull('url_param')->whereOr('url_param', '');
                })
                ->where('is_delete', 0)
                ->select();
            // 批量上传文件到谷歌云
            foreach ($uploadFiles as $file) {
                $imageSrc = root_path('public/uploads') . $file['save_name'];
                $objectPath = $catalogue . '/' . $file['save_name'];
                $object = $storage->bucket($bucket)->object($objectPath);
                $url = $object->signedUrl(new \DateTime('+100 years'));
                //
                if (!$object->exists()) {
                    if (file_exists($imageSrc)) {
                        $storage->bucket($bucket)->upload(file_get_contents($imageSrc), [
                            'name' => $objectPath,
                            'predefinedAcl' => 'publicRead' // 设置为公开访问
                        ]);
                        // 更新数据库中的存储类型为google
                        $file->where('company_uuid', $appId)->where('file_id', $file['file_id'])->update([
                            'url_param' =>  explode('?', $url)[1],
                            'file_url' => "https://storage.googleapis.com/$bucket/$catalogue",
                        ]);
                        //
                        dump("文件处理成功: " . $imageSrc);
                    }
                } else {
                    // 更新数据库中的存储类型为google
                    $file->where('company_uuid', $appId)->where('file_id', $file['file_id'])->update([
                        'url_param' =>  explode('?', $url)[1],
                        'file_url' => "https://storage.googleapis.com/$bucket/$catalogue",
                    ]);
                    //
                    dump("文件处理成功: " . $imageSrc);
                }
            }
        }
    }
}
