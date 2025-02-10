<?php

namespace app\common\model\file;

use app\common\model\BaseModel;

/**
 * 文件库模型
 */
class UploadFile extends BaseModel
{
    protected $name = 'file';
    protected $pk = 'id';
     /**
     * 追加属性
     */
    protected $append = ['file_path', 'file_id', 'group_id'];

    /**
     * 兼容ID字段
     */
    public function getFileIdAttr()
    {
        return $this->uuid ?? 0;
    }

    /**
     * 兼容ID字段
     */
    public function getGroupIdAttr()
    {
        return $this->group_uuid ?? 0;
    }

    /**
     * 关联文件库分组表
     */
    public function uploadGroup()
    {
        return $this->belongsTo('UploadGroup', 'group_uuid', 'uuid');
    }

    /**
     * 获取图片完整路径
     * @param $value
     * @param $data
     * @return string
     */
    public function getFilePathAttr($value, $data)
    {
        $videoExtensions = ['.mp4', '.avi', '.mov', '.wmv', '.flv', '.mkv', '.webm', '.m4v', '.mpeg', '.3gp'];
        # 存储驱动程序上传图片等使用: google,local
        if ($data['storage'] === 'local' && (!($data['url_param'] ?? '') || env('STORAGE_DRIVER', 'local') == 'local')) {
            if (($data['file_type'] ?? '') === '.video' || in_array('.' . pathinfo($data['save_name'], PATHINFO_EXTENSION), $videoExtensions)) {
                return  base_url() . 'uploads/' . $data['save_name'];
            }
            return  base_url() . 'api/product/thumbs/uploads/' . $data['save_name'];
        }
        if ($data['storage'] === 'google' || strstr($data['url_param'] ?? '', 'GoogleAccessId')) {
            return str_replace('https://storage.googleapis.com/', base_url() . 'storage_googleapis/', ($data['file_url'] ?? '') . '/' . ($data['save_name'] ?? '') . '?' . ($data['url_param'] ?? '')) . '&';
        }
        return ($data['file_url'] ?? '') . '/' . ($data['file_name'] ?? '');
    }

    /**
     * 文件详情
     */
    public static function detail($file_id)
    {
        return self::where('uuid', $file_id)->find();
    }

    /**
     * 根据文件名查询文件id
     */
    public static function getFildIdByName($fileName)
    {
        return (new static)->where(['file_name' => $fileName])->value('uuid');
    }

    /**
     * 查询文件id
     */
    public static function getFileName($fileId)
    {
        return (new static)->where(['uuid' => $fileId])->value('file_name');
    }

    /**
     * 添加新记录
     */
    public function add($data)
    {
        return $this->save($data);
    }

    /**
     * 获取列表记录
     */
    public function getList($groupId, $fileType, $pageSize, $isRecycle, $shop_supplier_id = 0)
    {
        $model = $this;
        // 文件分组
        if ($groupId != 0) {
            $model = $model->where('group_uuid', '=', (int)$groupId);
        }
        // 文件类型
        !empty($fileType) && $model = $model->where('file_type', '=', trim($fileType));
        // 是否在回收站
        $isRecycle > -1 && $model = $model->where('is_recycle', '=', (int)$isRecycle);
        // 查询列表数据
        return $model->with(['upload_group'])
            ->where(['is_user' => 0])
            ->order(['id' => 'desc'])
            ->paginate($pageSize);
    }
}
