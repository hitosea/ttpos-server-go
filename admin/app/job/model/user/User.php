<?php

namespace app\job\model\user;

use app\common\model\user\User as UserModel;
use app\common\enum\user\grade\ChangeTypeEnum;
use app\job\model\user\GradeLog as GradeLogModel;

/**
 * 用户模型
 */
class User extends UserModel
{
    /**
     * 批量设置会员等级
     */
    public function upgradeGrade($user, $upgradeGrade)
    {
        // 更新会员等级的数据
        $this->where('uuid', '=', $user['user_id'])
            ->update([
                'grade_id' => $upgradeGrade['grade_id']
            ]);
        (new GradeLogModel)->save([
            'old_grade_id' => $user['grade_id'],
            'new_grade_id' => $upgradeGrade['grade_id'],
            'change_type' => ChangeTypeEnum::AUTO_UPGRADE,
            'member_uuid' => $user['user_id'],
        ]);
        return true;
    }

    /**
     * 获取生日会员
     */
    public function getBirthList()
    {
        $birthSql = "UNIX_TIMESTAMP(concat(YEAR(NOW()),FROM_UNIXTIME(birthday,'-%m-%d')))-UNIX_TIMESTAMP(FROM_UNIXTIME(unix_timestamp(now()),'%Y-%m-%d'))=86400";
        $sendSql = "YEAR(FROM_UNIXTIME(send_time,'%Y-%m-%d'))<>YEAR(now())";
        $list = UserModel::where('delete_time', '=', 0)->where($birthSql)->where($sendSql)->limit(10)->select();
        return $list;
    }
}
