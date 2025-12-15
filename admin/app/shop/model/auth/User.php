<?php

namespace app\shop\model\auth;

use think\facade\Env;
use help\ValidateHelp;
use think\facade\Cache;
use app\admin\model\CompanyStaff;
use app\common\service\websocket\Websocket;
use app\common\model\shop\User as UserModel;
use app\admin\model\admin\Staff as SaasStaffModel;


/**
 * 角色模型
 */
class User extends UserModel
{

    /**
     * 获取列表
     * @param mixed $limit
     * @return \think\Paginator
     */
    public function getList($limit = 20)
    {
        return $this->with(['userRole.role', 'supplier'])
            ->field('uuid, uuid as shop_user_id, username as user_name, real_name, is_super, user_type, is_disable as is_status, create_time')
            ->order(['create_time' => 'desc'])
            ->paginate($limit);
    }

    /**
     * 基础列表
     */
    public function getBaseList($limit = 20)
    {
        return $this->field('uuid, uuid as shop_user_id, username as user_name, real_name')
            ->order(['create_time' => 'desc'])
            ->paginate($limit);
    }

    /**
     * 获取所有上级id集
     */
    public function getTopRoleIds($role_id, &$all = null)
    {
        static $ids = [];
        is_null($all) && $all = $this->getAll();
        foreach ($all as $item) {
            if ($item['role_id'] == $role_id && $item['parent_id'] > 0) {
                $ids[] = $item['parent_id'];
                $this->getTopRoleIds($item['parent_id'], $all);
            }
        }
        return $ids;
    }

    /**
     * 获取所有角色
     */
    private function getAll()
    {
        $data = $this->order(['create_time' => 'asc'])->select();
        return $data ? $data->toArray() : [];
    }

    public function add($data, $user)
    {
        $appId = request()->appId;
        // 邮箱
        $email = $data['user_name'] ?? '';
        $phone = $data['phone'] ?? '';
        $emailText = ValidateHelp::validateEmail($email);
        if ($emailText !== true) {
            $this->error = $emailText;
            return false;
        }
        $num = (new CompanyStaff([], 0))->setAppId(0)->withoutGlobalScope()->where('username', $email)->count();
        if ($num > 0) {
            $this->error = "邮箱已存在";
            return false;
        }
        // 手机号
        $phoneText = ValidateHelp::validatePhone($phone);
        if ($phoneText !== true) {
            $this->error = $phoneText;
            return false;
        }
        $phoneNum = (new CompanyStaff([], 0))->setAppId(0)->withoutGlobalScope()->where('phone', $phone)->count();
        if ($phoneNum > 0) {
            $this->error = "手机号已存在";
            return false;
        }
        // 角色是否存在
        $role = (new Role([], $appId))->where('uuid', 'in', $data['role_id'])->count();
        if ($role != count($data['role_id'])) {
            $this->error = '角色参数错误';
            return false;
        }
        // 密码校验
        if (!ValidateHelp::validateAlphaPassword($data['password'])) {
            $this->error = '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种';
            return false;
        }
        // 权限密码校验（必填）
        if (empty($data['permission_password'])) {
            $this->error = '权限密码不能为空';
            return false;
        }
        if (!ValidateHelp::validatePermissionPassword($data['permission_password'])) {
            $this->error = '密码必须为 4 - 8 位数字';
            return false;
        }

        $this->startTrans();
        try {
            //
            $arr = [
                'uuid' => createUuid(),
                'phone' => trim($data['phone']),
                'username' => trim($data['user_name']),
                'password' => salt_hash($data['password']),
                'permission_password' => salt_hash($data['permission_password']),
                'real_name' => trim($data['real_name']),
                'user_type' => $user['user_type'],
                'company_uuid' => $appId
            ];
            // 平台saas.ttpos_staff表
            (new SaasStaffModel([], 0))->setAppId(0)->create([
                "uuid" => $arr['uuid'],
                "email" => $arr['username'],
                "phone" => $arr['phone'],
                "real_name" => $arr['real_name'],
                "password" => $arr['password'],
                "last_company_uuid" => $appId,
            ]);
            // 添加到平台主表
            $res = (new CompanyStaff([], 0))->setAppId(0)->create($arr);
            self::setAppId($appId)->create($arr);
            //
            $add_arr = [];
            $model = new UserRole();
            foreach ($data['role_id'] as $val) {
                $add_arr[] = [
                    'uuid' => createUuid(),
                    'staff_uuid' => $res['uuid'],
                    'role_uuid' => $val,
                ];
            }
            $model->saveAll($add_arr);
            // 事务提交
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    public function getUserName($where, $shop_user_id = 0)
    {
        if ($shop_user_id > 0) {
            return $this->withoutGlobalScope()->where($where)->where('shop_user_id', '<>', $shop_user_id)->count();
        }
        return $this->withoutGlobalScope()->where($where)->count();
    }

    public function getSupplierUserName($where, $shop_user_id = 0)
    {
        if ($shop_user_id > 0) {
            return $this->withoutGlobalScope()->where($where)->where('shop_user_id', '<>', $shop_user_id)->count();
        }
        return $this->withoutGlobalScope()->where($where)->count();
    }

    public function edit($data)
    {
        $appId = request()->appId;
        // 邮箱
        $email = $data['user_name'] ?? '';
        $phone = $data['phone'] ?? '';
        $emailText = ValidateHelp::validateEmail($email);
        if ($emailText !== true) {
            $this->error = $emailText;
            return false;
        }
        //
        $num = (new CompanyStaff([], 0))->setAppId(0)->withoutGlobalScope()->where('username', $email)->where('uuid', '<>', $data['shop_user_id'])->count();
        if ($num > 0) {
            $this->error = "邮箱已存在";
            return false;
        }
        // 手机号
        $phoneText = ValidateHelp::validatePhone($phone);
        if ($phoneText !== true) {
            $this->error = $phoneText;
            return false;
        }
        $phoneNum = (new CompanyStaff([], 0))->setAppId(0)->withoutGlobalScope()->where('phone', $phone)->where('uuid', '<>', $data['shop_user_id'])->count();
        if ($phoneNum > 0) {
            $this->error = "手机号已存在";
            return false;
        }
        // 角色是否存在
        $role = (new Role([], $appId))->setAppId($appId)->where('uuid', 'in', $data['role_id'])->count();
        if ($role != count($data['role_id'])) {
            $this->error = '角色参数错误';
            return false;
        }
        // 密码校验
        if ($data['password'] && !ValidateHelp::validateAlphaPassword($data['password'])) {
            $this->error = '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种';
            return false;
        }
        // 权限密码校验（非必填，但如果传了则必须符合格式）
        if (!empty($data['permission_password']) && !ValidateHelp::validatePermissionPassword($data['permission_password'])) {
            $this->error = '密码必须为 4 - 8 位数字';
            return false;
        }

        $this->startTrans();
        try {
            $arr = [
                'phone' => $data['phone'],
                'username' => $data['user_name'],
                'password' => salt_hash($data['password']),
                'real_name' => $data['real_name'],
            ];
            if (empty($data['password'])) {
                unset($arr['password']);
            } else {
                $arr['password_change_time'] = time();
            }
            // 权限密码处理：如果传了且不为空，则更新；否则不更新（保持原值）
            if (!empty($data['permission_password'])) {
                $arr['permission_password'] = salt_hash($data['permission_password']);
            }

            $where['uuid'] = $data['shop_user_id'];
            $saasStaffUpdate = $arr;
            $saasStaffUpdate['email'] = $data['user_name'];
            (new SaasStaffModel([], 0))->setAppId(0)->update($saasStaffUpdate, $where);
            (new CompanyStaff([], 0))->setAppId(0)->update($arr, $where);
            self::setAppId($appId)->update($arr, $where);

            $model = new UserRole();
            UserRole::destroy(['staff_uuid' => $data['shop_user_id']]);
            $add_arr = [];
            foreach ($data['role_id'] as $val) {
                $add_arr[] = [
                    'uuid' => createUuid(),
                    'staff_uuid' => $data['shop_user_id'],
                    'role_uuid' => $val,
                ];
            }
            $model->saveAll($add_arr);
            // 事务提交
            $this->commit();
            // 删除收银机缓存
            Cache::tag('cashier')->clear();
            // 推送配置更新
            Websocket::pushClient(
                request()->appId,
                Websocket::SOURCE_All,
                Websocket::SOURCE_All,
                Websocket::UPDATE_PERMISSION,
                $data['shop_user_id'],
                [
                    'staff_uuid' => $data['shop_user_id'],
                    'update_time' => time(),
                ]
            );
            //
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    public function getChild($where)
    {
        return $this->where($where)->count();
    }

    /**
     * 删除
     *
     */
    public function del($shop_user_id, $user)
    {
        if ($this->checkCashierOnline($shop_user_id)) {
            $this->error = '当前人员未交班，请先交班';
            return false;
        }

        if ($user['shop_user_id'] == $shop_user_id) {
            $this->error = '不能删除当前登录账号';
            return false;
        }

        $userToDelete = self::where('uuid', $shop_user_id)->find();
        if (!$userToDelete) {
            $this->error = '用户不存在';
            return false;
        }
        if ($userToDelete['is_super'] == 1) {
            $this->error = '超级管理员不能被删除';
            return false;
        }
        $userToDelete->delete();
        UserRole::destroy(['staff_uuid' => $shop_user_id]);
        //
        return (new CompanyStaff([], 0))->setAppId(0)->where('uuid', $shop_user_id)->find()?->delete();
    }

    /**
     * 更改状态
     */
    public function setStatus($status)
    {
        if ($this->checkCashierOnline($this['uuid'])) {
            $this->error = '当前人员未交班，请先交班';
            return false;
        }
        if ($this['is_super'] == 1) {
            $this->error = '超级管理员不能被禁用';
            return false;
        }
        // 删除收银机缓存
        Cache::tag('cashier')->clear();
        return $this->save([
            'is_disable' => $status
        ]);
    }

    /**
     * 检测用户是否在收银坐班
     *
     * @param int $shop_user_id
     * @return bool
     */
    public function checkCashierOnline($shop_user_id)
    {
        $where = [
            'uuid' => $shop_user_id,
            'cashier_online' => 1,
        ];
        return self::where($where)->count() > 0;
    }
}
