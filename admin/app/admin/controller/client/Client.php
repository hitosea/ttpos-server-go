<?php

namespace app\admin\controller\client;

use think\facade\Cache;
use Endroid\QrCode\QrCode;
use think\facade\Validate;
use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use Endroid\QrCode\Writer\PngWriter;
use app\admin\validate\ClientValidate;
use app\common\model\settings\Setting;
use Google\Cloud\Storage\StorageClient;
use app\common\enum\settings\SettingEnum;
use app\common\library\language\engine\OpenAi;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\storage\Driver as StorageDriver;
use app\common\model\client\ClientVersion as ClientVersionModel;

/**
 * 客户端管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(4)
 */
class Client extends Controller
{
    /**
     * @Apidoc\Title("AI翻译转发")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/client.client/aiTranslate")
     * @Apidoc\Param("data", type="array", require=true, default="", desc="数据")
     */
    public function aiTranslate()
    {
        $ai = new OpenAi();
        $res = $ai->forward(request()->post());
        if ($res) {
            $res['data'] = json_encode($res['data'], JSON_UNESCAPED_UNICODE);
        }
        return $this->renderSuccess('', $res);
    }

    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/api/admin/client.client/index")
     * @Apidoc\Param("type", type="string", require=true, default="1", desc="类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端,6自助点餐机")
     * @Apidoc\Param("version_number", type="string", require=true, default="", desc="版本号")
     * @Apidoc\Param("forced_update", type="string", require=true, default="", desc="强制更新 0全部 1是 2否")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\client\ClientVersion\lists", desc="列表", withoutField="package_url", children={
     *    @Apidoc\Returned("uuid", type="float", desc="uuid，用于下载"),
     * }))
     */
    public function index()
    {
        $param = $this->postData();
        $type = $param['type'] ?? 1;
        $versionNumber = $param['version_number'] ?? '';
        $forcedUpdate = $param['forced_update'] ?? '';
        $types = [1 => 'Cashier', 2 => 'Menu', 3=> 'Kitchen', 4 => 'Shop', 5 => 'Assistant', 6 => 'Kiosk'];
        //
        $credentialsPath = root_path('runtime/storage') . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        if (!file_exists($credentialsPath)) {
            $credentialsPath = '/var/certificate/' . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        }
        $bucket = env('GOOGLE_APPLICATION_BUCKET_NAME');
        $env = env('GOOGLE_APPLICATION_BUCKET_ENV');
        if (!file_exists($credentialsPath) || !$bucket || !$env) {
            return $this->renderError('Not configured.');
        }
        //
        $platform = Cache::get('RENEWALINFO-PLATFORM');
        if (!$platform) {
            $platform = Setting::where('key', SettingEnum::PLATFORM_BRAND)->value('describe') ?: 'ttpos';
        }
        if (!$platform) {
            return $this->renderError('无平台信息');
        }
        //
        $platform = strtoupper($platform);
        //
        $storage = new StorageClient(['keyFilePath' => $credentialsPath]);
        $terminal = $types[$type];
        // 获取bucket下面的文件列表
        $objects = $storage->bucket($bucket)->objects(['prefix' => "{$platform}/{$env}/{$terminal}/"]);
        // 对象数组
        $objectsDatas = [];
        //
        foreach ($objects as $object) {
            $name = $object->name();
            if (pathinfo($name, PATHINFO_EXTENSION) !== 'apk') {
                continue;
            }
            $info = $object->info();
            $objectsDatas[] = [
                'md5_hash' => $info['md5Hash'],
                'original_name' => $name,
            ];
            //
            if ($document = ClientVersionModel::where('original_name', $name)->find()) {
                if ($document->md5_hash != $info['md5Hash']) {
                    $document->md5_hash = $info['md5Hash'];
                    $document->create_time = strtotime($info['timeCreated']);
                    $document->update_time = strtotime($info['updated']);
                    $document->save();
                }
                continue;
            }
            //
            preg_match('/V([\d.]+)\.apk$/', $name, $matches);
            $version = $matches[1] ?? '0.0.0';
            //
            $brand = $platform == 'TTPOS' ? 1 : 2;
            $packageName = "com." . strtolower($platform) . '.' . strtolower($terminal);
            //
            $apkData['name'] = $packageName;
            $apkData['versionCode'] = $apkData['compileSdkVersion'] = version_compare('1.0.8', $version, '<') ? 300 : 0;
            $apkData['versionName'] = $version;
            $apkData['platformBuildVersionName'] = $version;
            //
            (new ClientVersionModel)->create([
                'name' => basename($name),
                'original_name' => $name,
                'type' => $type,
                'brand' => $brand,
                'version_number' => $apkData['versionCode'],
                'version_name' => $version,
                'apk_version_code' => $apkData['versionCode'],
                'forced_update' => 0,
                'is_publish' => 0,
                'update_log' => '',
                'apk_data' => json_encode($apkData, JSON_UNESCAPED_SLASHES),
                'package_url' => strtok($object->signedUrl(new \DateTime('+1 hour')), '?'),
                'md5_hash' => $info['md5Hash'],
                'create_time' => strtotime($info['timeCreated']),
                'update_time' => strtotime($info['updated']),
            ]);
        }

        // 清理被删除的文件
        $deleteData = (new ClientVersionModel)
            ->where('type', $type)
            ->whereNotIn('original_name', array_column($objectsDatas, 'original_name'))
            ->select();
        foreach ($deleteData as $data) {
            $object = $storage->bucket($bucket)->object($data->original_name);
            if (!$object->exists()) {
                $data->delete();
            }
        }

        //
        $param = $this->postData();
        $type = $param['type'] ?? 1;
        $versionNumber = $param['version_number'] ?? '';
        $forcedUpdate = $param['forced_update'] ?? '';
        //
        $list = ClientVersionModel::when($versionNumber, function ($q) use ($versionNumber) {
                $q->like('version_name', $versionNumber);
            })
            ->like('original_name', "{$platform}/{$env}/{$terminal}/")
            ->when($forcedUpdate, function ($q) use ($forcedUpdate) {
                $q->where('forced_update', $forcedUpdate == 1 ? 1 : 0);
            })
            ->field('id, type, brand, download_num, version_number, is_publish, version_name, forced_update, update_log, update_time as create_time')
            ->field('MD5(package_url) as uuid')
            ->where('type', $type)
            ->orderRaw('CAST(SUBSTRING_INDEX(version_name, ".", 1) AS UNSIGNED) DESC, 
                      CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(version_name, ".", 2), ".", -1) AS UNSIGNED) DESC,
                      CAST(SUBSTRING_INDEX(version_name, ".", -1) AS UNSIGNED) DESC')
            ->paginate($param);

        foreach ($list as $k => $item) {
            $list[$k]["language_update_log"] = $item['update_log'];
            $updateLogs = json_decode($item["update_log"], true);
            $language = checkDetect();
            if (isset($updateLogs[$language])) {
                $list[$k]["language_update_log"] = $updateLogs[$language];
            }
        }
        //
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("发布")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/client.client/publish")
     * @Apidoc\Param("id", type="string", require=true, default="", desc="ID")
     * @Apidoc\Param("update_log", type="string", require=true, default="", desc="更新日志")
     * @Apidoc\Param("forced_update", type="int", require=true, default="", desc="强制更新 0全部 1是 2否")
     * @Apidoc\Returned()
     */
    public function publish(ClientValidate $validate)
    {
        $param = $validate->goCheck('publish');
        //
        $res = (new ClientVersionModel)->where('id', $param['id'])->find();
        if (!$res) {
            return $this->renderError("数据不存在");
        }
        //
        $res->update_log = $param['update_log'];
        $res->forced_update = $param['forced_update'] ?? 0;
        $res->is_publish = 1;
        $res->save();
        //
        return $this->renderSuccess();
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/client.client/add")
     * @Apidoc\Param("type", type="int", require=true, default="", desc="类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端,6自助点餐机")
     * @Apidoc\Param(ref="app\common\model\client\ClientVersion\lists", desc="登录日志列表", withoutField="id,num,create_time,update_time")
     * @Apidoc\Returned("list", type="array", ref="app\common\model\admin\LoginLog\getList", desc="登录日志列表")
     */
    public function add(ClientValidate $validate)
    {
        $param = $validate->goCheck('add');
        //
        $filePath = public_path() . ($param['package_url'] ?? '');
        if (!file_exists($filePath)) {
            return $this->renderError("上传文件不存在或已丢失，请重新上传");
        }
        //
        $originalName = cache('client_upload_original_name_' . $param['package_url']) ?: '';
        //
        $res = (new ClientVersionModel)
            ->where('type', $param['type'])
            ->where('brand', $param['brand'])
            ->where('version_number', $param['version_number'])
            ->find();
        if ($res) {
            return $this->renderError("版本号已经存在");
        }
        //
        exec("aapt dump badging $filePath", $output, $returnVar);
        if ($returnVar !== 0) {
            return $this->renderError(__("aapt 命令执行失败，返回状态码：") . $returnVar);
        }
        if (!$output) {
            return $this->renderError("无法正常解析APK文件");
        }
        // 使用正则表达式匹配所有键值对
        preg_match_all("/(\w+)='([^']+)'/", $output[0], $matches);
        $parsedArray = [];
        foreach ($matches[0] as $match) {
            preg_match("/(\w+)='([^']+)'/", $match, $pair);
            if (count($pair) === 3) {
                $parsedArray[$pair[1]] = $pair[2];
            }
        }
        //
        $param['version_number'] = $parsedArray['versionCode'];
        $param['version_name'] = $parsedArray['versionName'];
        $param['apk_data'] = json_encode($parsedArray);
        $param['apk_version_code'] = $parsedArray['versionCode'];
        $param['original_name'] = $originalName;
        //
        $res = (new ClientVersionModel)->create($param);
        return $this->renderSuccess('', $res);
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/client.client/delete")
     * @Apidoc\Param("id", type="int", require=true, default="", desc="ID")
     */
    public function delete(ClientValidate $validate)
    {
        $param = $validate->goCheck('id');
        $version = (new ClientVersionModel)->where('id', $param['id'])->find();
        if (!$version) {
            return $this->renderError("数据不存在");
        }
        $version->delete();
        return $this->renderSuccess();
    }

    /**
     * @Apidoc\Title("二维码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/client.client/qrcode")
     * @Apidoc\Param("uuid", type="string", require=true, default="", desc="UUID")
     * @Apidoc\Returned("", type="string",  desc="base64 二维码数据")
     */
    public function qrcode(ClientValidate $validate)
    {
        $param = $validate->goCheck('uuid');
        $version = (new ClientVersionModel)->whereRaw("MD5(package_url) = '" . $param['uuid'] . "'")->find();
        if (!$version) {
            return $this->renderError("数据不存在");
        }
        //
        $qrCode = new QrCode($version->package_url);
        //
        return $this->renderSuccess('', (new PngWriter())->write($qrCode)->getDataUri());
    }

    /**
     * @Apidoc\Title("下载")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/client.client/download")
     * @Apidoc\Param("uuid", type="string", require=true, default="", desc="UUID")
     */
    public function download(ClientValidate $validate)
    {
        $param = $validate->goCheck('uuid');
        $version = ClientVersionModel::whereRaw("MD5(package_url) = '" . $param['uuid'] . "'")->find();
        if (!$version) {
            return $this->renderError("数据不存在");
        }
        //
        $version->download_num = $version->download_num + 1;
        $version->save();
        //
        return redirect($version->package_url);
    }

    /**
     * @Apidoc\Title("上传接口")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/client.client/upload")
     * @Apidoc\Param("file", type="file", default="", desc="文件")
     * @Apidoc\Returned("file_path", type="string", default="/20240327/8018f81f7285a7403937635c338e747c.zip", desc="使用这个字段去预览保存")
     * @Apidoc\Returned("version_number", type="string", default="", desc="版本号")
     * @Apidoc\Returned("version_name", type="string", default="", desc="版本名称")
     */
    public function upload()
    {
        $fileInfo = request()->file('file');
        $validate = Validate::rule([
            'image' => 'file|fileExt:apk|max:100000000',
        ])->message([
            'image.file' => '请上传文件',
            'image.fileMime' => '文件类型必须为apk格式',
            'image.max' => '文件大小不能超过100MB',
        ]);
        if (!$validate->check(['image' => $fileInfo])) {
            $errors = $validate->getError();
            return $this->renderError($errors);
        }
        // 实例化存储驱动
        $config = SettingModel::getItem('storage');
        $storageDriver = new StorageDriver($config);
        // 设置上传文件的信息
        $storageDriver->setUploadFile('file');
        // 上传
        $saveName = $storageDriver->upload();
        if ($saveName == '') {
            return $this->renderError('上传失败' . $storageDriver->getError());
        }
        $saveName = str_replace('\\', '/', $saveName);
        //
        $filePath =  'uploads/' . $saveName;
        //
        exec("aapt dump badging $filePath", $output, $returnVar);
        if ($returnVar !== 0) {
            return $this->renderError(__("aapt 命令执行失败，返回状态码：") . $returnVar);
        }
        if (!$output) {
            return $this->renderError("无法正常解析APK文件");
        }
        // 使用正则表达式匹配所有键值对
        preg_match_all("/(\w+)='([^']+)'/", $output[0], $matches);
        $parsedArray = [];
        foreach ($matches[0] as $match) {
            preg_match("/(\w+)='([^']+)'/", $match, $pair);
            if (count($pair) === 3) {
                $parsedArray[$pair[1]] = $pair[2];
            }
        }
        //
        $data['file_path'] = $filePath;
        $data['apk_data'] = $parsedArray;
        $data['version_number'] = $parsedArray['versionCode'];
        $data['version_name'] = $parsedArray['versionName'];
        $data['apk_version_code'] = $parsedArray['versionCode'];
        //
        cache('client_upload_original_name_' . $filePath, $fileInfo->getOriginalName(), 60 * 60 * 24);
        //
        return $this->renderSuccess("上传成功", $data);
    }

    /**
     * @Apidoc\Title("获取客户端最新版本")
     * @Apidoc\Method ("get")
     * @Apidoc\Url("/api/admin/client.client/getNewVersion")
     * @Apidoc\Param("type", type="string", default="product", desc="类型：1收银端,2平板端,3厨显端,4商家后台端,5点餐助手端,6自助点餐机")
     * @Apidoc\Param("brand", type="string", default="product", desc="品牌：0最新的一条, 1TTPOS， 2jbc")
     * @Apidoc\Returned("brand", type="string", desc="品牌")
     * @Apidoc\Returned("version_name", type="string", desc="版本名称")
     * @Apidoc\Returned("version_number", type="string", desc="版本号")
     * @Apidoc\Returned("apk_version_code", type="string", desc="apk包的版本code")
     * @Apidoc\Returned("forced_update", type="string", desc="是否强制更新 0否 1是")
     * @Apidoc\Returned("download_url", type="string", desc="下载地址")
     * @Apidoc\Returned("update_log", type="string", desc="更新日志")
     * @Apidoc\Returned("apk_data", type="array", desc="apk包数据", children={
     *    @Apidoc\Property("name",type="string", desc=""),
     *    @Apidoc\Property("versionCode", type="string", desc=""),
     *    @Apidoc\Property("versionName", type="string", desc=""),
     *    @Apidoc\Property("platformBuildVersionName", type="string", desc=""),
     *    @Apidoc\Property("compileSdkVersion", type="string", desc=""),
     *    @Apidoc\Property("compileSdkVersionCodename", type="string", desc=""),
     * })
     */
    public function getNewVersion()
    {
        $param = $this->postData();
        //
        $type = $param['type'] ?? 1;
        $brand = $param['brand'] ?? 0;
        //
        $credentialsPath = root_path('runtime/storage') . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        if (!file_exists($credentialsPath)) {
            $credentialsPath = '/var/certificate/' . env('GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME');
        }
        $bucket = env('GOOGLE_APPLICATION_BUCKET_NAME');
        $env = env('GOOGLE_APPLICATION_BUCKET_ENV');
        if (!file_exists($credentialsPath) || !$bucket || !$env) {
            return $this->renderSuccess('');
        }
        //
        $result = ClientVersionModel::field('md5_hash, brand, version_number, version_name, apk_version_code, apk_data, forced_update, update_log, original_name')
            ->field("package_url as download_url")
            ->orderRaw('CAST(SUBSTRING_INDEX(version_name, ".", 1) AS UNSIGNED) DESC, 
                      CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(version_name, ".", 2), ".", -1) AS UNSIGNED) DESC,
                      CAST(SUBSTRING_INDEX(version_name, ".", -1) AS UNSIGNED) DESC')
            ->where('type', $type)
            ->where('is_publish', 1)
            ->when($brand, function ($q) use ($brand) {
                $q->where('brand', $brand);
            })
            ->find();
        //
        if ($result) {
            // 获取bucket下面的文件列表
            $storage = new StorageClient(['keyFilePath' => $credentialsPath]);
            $objects = $storage->bucket($bucket)->object($result->original_name);
            $info = $objects->info();
            //
            if ($result->md5_hash != $info['md5Hash']) {
                $result->md5_hash = $info['md5Hash'];
                $result->create_time = strtotime($info['timeCreated']);
                $result->update_time = strtotime($info['updated']);
                $result->save();
            }
            //
            $result = $result?->toArray() ?: [];
            $result['update_log'] =  $result['update_log_text'];
        }
        //
        return $this->renderSuccess('', $result);
    }
}
