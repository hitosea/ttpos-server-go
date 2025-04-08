<?php

namespace app\shop\model\user;

use app\common\model\user\Card;
use app\common\model\user\MemberCard;
use app\common\model\user\CardRecord as CardRecordModel;

/**
 * 会员卡记录模型
 */
class CardRecord extends CardRecordModel
{
    /**
     * 获取列表记录
     */
    public function getList($data)
    {
        $model = $this->withTrashed()->alias('r')
                ->field('r.*')
                ->with(['card', 'user'])
                ->join('member u', 'u.uuid=r.member_uuid')
                ->join('member_card_type c', 'c.uuid=r.member_card_type_uuid')
                ->where('r.delete_time', '=', 0)
                ->where('u.delete_time', '=', 0)
                ->order(['r.create_time' => 'desc']);

        if (isset($data['card_name']) && $data['card_name'] != '') {
            $model = $model->like('c.name', $data['card_name']);
        }

        if (isset($data['status']) && $data['status'] >= 0) {
            $model = $model->where(function ($query) use ($data) {
                $query = $query->where('r.delete_time', '=', 0);
                if ($data['status'] == 0) {
                    $query->where('r.expire', '<', time())->where('r.expire', '>', 0);
                } else {
                    $query->where('r.expire', '>=', time())
                        ->whereOr('r.expire', 0);
                }
            });
        }

        $list = $model->paginate($data) ?: [];
        //
        foreach ($list as &$item) {
            $item['is_used'] = (new Card)->checkUserConsumeRecord($item['user_id'], $item['card_id']) ? 1 : 0;
            if ($item['delete_time'] == 0) {
                $memberCard = (new MemberCard)->where('member_uuid', $item['member_uuid'])->find();
                $item['expire_time_text'] = date('Y-m-d', $memberCard['expire'] ?: 0);
            } else {
                $item['expire_time_text'] = date('Y-m-d', $item['expire']);
            }
        }
        return $list;
    }

    /**
     * 延期
     */
    public function delay($data)
    {
        $isExist = MemberCard::checkExistByUserId($this['member_uuid'], $this['member_card_type_uuid']);
        if ($isExist?->isEmpty()) {
            $this->error = "会员卡不存在";
            return false;
        }
        $update['expire_time'] = strtotime($data['expire_time']);
        return $isExist->save($update);
    }
}
