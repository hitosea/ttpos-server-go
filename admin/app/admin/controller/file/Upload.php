<?php

namespace app\admin\controller\file;

use think\facade\Validate;
use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\model\file\UploadFile;
use app\common\model\settings\Setting as SettingModel;
use app\common\library\storage\Driver as StorageDriver;

/**
 * 文件库管理
 * @Apidoc\Group("base")
 * @Apidoc\Sort(2)
 */
class Upload extends Controller
{
    /**
     * @Apidoc\Title("图片上传接口")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/file.Upload/image")
     * @Apidoc\Param("iFile", type="file", default="", desc="文件")
     * @Apidoc\Returned("id", type="string", default="1", desc="文件的id")
     * @Apidoc\Returned("file_path", type="string", default="http://127.0.0.1/uploads/20240327/8018f81f7285a7403937635c338e747c.png", desc="使用这个字段去预览保存")
     */
    public function image($group_id = -1)
    {
        // 图片信息
        $fileInfo = request()->file('iFile');
        $validate = Validate::rule([
            'image' => 'file|fileExt:jpg,png,jpeg',
        ])->message([
            'image.file' => '请上传文件',
            'image.fileMime' => '文件类型必须为图片格式',
        ]);
        if (!$validate->check(['image' => $fileInfo])) {
            $errors = $validate->getError();
            return $this->renderError($errors);
        }
        // 实例化存储驱动
        $config = SettingModel::getItem('storage');
        $storageDriver = new StorageDriver($config);
        // 设置上传文件的信息
        $storageDriver->setUploadFile('iFile');
        // 上传图片
        $saveName = $storageDriver->upload();
        if ($saveName == '') {
            return json(['code' => 0, 'msg' => '图片上传失败' . $storageDriver->getError()]);
        }
        $saveName = str_replace('\\', '/', $saveName);
        // 图片上传路径
        $fileName = $storageDriver->getFileName();
        // 访问参数
        $urlParam = $storageDriver->getUrlParam();
        // 添加文件库记录
        $uploadFile = $this->addUploadFile($group_id, $fileName, $fileInfo, request()->param('file_type') ?? 'image', $saveName, $urlParam);
        // 图片上传成功
        return json(['code' => 1, 'msg' => '图片上传成功', 'data' => $uploadFile]);
    }

    /**
     * 添加文件库上传记录
     */
    private function addUploadFile($group_id, $fileName, $fileInfo, $fileType, $savename, $urlParam='')
    {
        // 存储引擎
        $config = SettingModel::getItem('storage');
        $storage = $config['default'];
        // 存储域名
        $fileUrl = isset($config['engine'][$storage]['domain'])
            ? $config['engine'][$storage]['domain'] : '';
        // 添加文件库记录
        $model = new UploadFile;
        $model->save([
            'group_id' => $group_id > 0 ? (int)$group_id : 0,
            'storage' => $storage,
            'file_url' => $fileUrl,
            'file_name' => $fileName,
            'save_name' => $savename,
            'file_size' => $fileInfo->getSize(),
            'file_type' => $fileType,
            'extension' => $fileInfo->getOriginalExtension(),
            'real_name' => $fileInfo->getOriginalName(),
            'url_param' => $urlParam,
        ]);
        return $model;
    }

    /**
     * 批量移动文件分组
     */
    public function moveFiles($group_id, $fileIds)
    {
        $model = new UploadFile;
        if ($model->moveGroup($group_id, $fileIds) !== false) {
            return $this->renderSuccess('移动成功');
        }
        return $this->renderError($model->getError() ?: '移动失败');
    }
}
