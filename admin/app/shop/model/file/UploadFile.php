<?php

namespace app\shop\model\file;

use app\common\model\file\UploadFile as UploadFileModel;

/**
 * 图片模型
 */
class UploadFile extends UploadFileModel
{

    /**
     * 软删除
     */
    public function softDelete($fileIds)
    {
        return $this->where('uuid', 'in', $fileIds)->delete();
    }

    /**
     * 批量移动文件分组
     */
    public function moveGroup($group_id, $fileIds)
    {
        return $this->where('uuid', 'in', $fileIds)->update(['group_uuid' => $group_id]);
    }
}
