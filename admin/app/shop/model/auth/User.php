<?php

namespace app\shop\model\auth;

use think\facade\Env;
use help\ValidateHelp;
use think\facade\Cache;
use app\common\model\shop\User as UserModel;


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
            ->field('shop_user_id, user_name, real_name, is_super, user_type, is_status, shop_supplier_id, app_id, create_time')
            ->where('is_delete', '=', 0)
            ->order(['create_time' => 'desc'])
            ->paginate($limit);
    }

    /**
     * 基础列表
     */
    public function getBaseList($limit = 20)
    {
        return $this->field('shop_user_id, user_name, real_name, shop_supplier_id, app_id')
            ->where('is_delete', '=', 0)
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
        $appId = self::$app_id;
        // 邮箱
        $email = $data['user_name'] ?? '';
        $phone = $data['phone'] ?? '';
        $emailText = ValidateHelp::validateEmail($email);
        if ($emailText !== true) {
            $this->error = $emailText;
            return false;
        }
        $num = self::setAppId(0)->withoutGlobalScope()->where('user_name', $email)->count();
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
        $phoneNum = self::setAppId(0)->withoutGlobalScope()->where('phone', $phone)->count();
        if ($phoneNum > 0) {
            $this->error = "手机号已存在";
            return false;
        }
        // 角色是否存在
        $role = (new Role([], $appId))->where('role_id', 'in', $data['role_id'])->count();
        if ($role != count($data['role_id'])) {
            $this->error = '角色参数错误';
            return false;
        }
        // 密码校验
        if (!ValidateHelp::validateAlphaPassword($data['password'])) {
            $this->error = '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种';
            return false;
        }

        $this->startTrans();
        try {
            //
            $arr = [
                'phone' => trim($data['phone']),
                'user_name' => trim($data['user_name']),
                'password' => salt_hash($data['password']),
                'real_name' => trim($data['real_name']),
                'role_id' => $data['role_id'],
                'shop_supplier_id' => $user['shop_supplier_id'],
                'user_type' => $user['user_type'],
                'app_id' => $appId
            ];
            // 添加到平台主表
            $res = self::setAppId(0)->create($arr);
            $arr['shop_user_id'] = $res['shop_user_id'];
            self::setAppId($appId)->create($arr);
            //
            $add_arr = [];
            $model = new UserRole();
            foreach ($data['role_id'] as $val) {
                $add_arr[] = [
                    'shop_user_id' => $res['shop_user_id'],
                    'role_id' => $val,
                    'app_id' => self::$app_id,
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
        $appId = self::$app_id;
        // 邮箱
        $email = $data['user_name'] ?? '';
        $phone = $data['phone'] ?? '';
        $emailText = ValidateHelp::validateEmail($email);
        if ($emailText !== true) {
            $this->error = $emailText;
            return false;
        }
        //
        $num = self::setAppId(0)->withoutGlobalScope()->where('user_name', $email)->where('shop_user_id', '<>', $data['shop_user_id'])->count();
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
        $phoneNum = self::setAppId(0)->withoutGlobalScope()->where('phone', $phone)->where('shop_user_id', '<>', $data['shop_user_id'])->count();
        if ($phoneNum > 0) {
            $this->error = "手机号已存在";
            return false;
        }
        // 角色是否存在
        $role = (new Role([], $appId))->setAppId($appId)->where('role_id', 'in', $data['role_id'])->count();
        if ($role != count($data['role_id'])) {
            $this->error = '角色参数错误';
            return false;
        }
        // 密码校验
        if ($data['password'] && !ValidateHelp::validateAlphaPassword($data['password'])) {
            $this->error = '不能包括空格，长度为8-16个字符必须包含字母、数字、符号中至少2种';
            return false;
        }

        $this->startTrans();
        try {
            $arr = [
                'phone' => $data['phone'],
                'user_name' => $data['user_name'],
                'password' => salt_hash($data['password']),
                'real_name' => $data['real_name'],
            ];
            if (empty($data['password'])) {
                unset($arr['password']);
            }

            $where['shop_user_id'] = $data['shop_user_id'];
            self::setAppId(0)->update($arr, $where);
            self::setAppId($appId)->update($arr, $where);

            $model = new UserRole();
            UserRole::destroy($where);
            $add_arr = [];
            foreach ($data['role_id'] as $val) {
                $add_arr[] = [
                    'shop_user_id' => $data['shop_user_id'],
                    'role_id' => $val,
                    'app_id' => self::$app_id
                ];
            }
            $model->saveAll($add_arr);
            // 事务提交
            $this->commit();
            // 删除收银机缓存
            Cache::tag('cashier')->clear();
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

        $userToDelete = self::find($shop_user_id);
        if (!$userToDelete) {
            $this->error = '用户不存在';
            return false;
        }
        if ($userToDelete['is_super'] == 1) {
            $this->error = '超级管理员不能被删除';
            return false;
        }
        $userToDelete->is_delete = 1;
        $userToDelete->save();
        //
        (new self([], 0))->where('shop_user_id', $shop_user_id)->find()?->save(['is_delete' => 1]);
        return UserRole::destroy(['shop_user_id' => $shop_user_id]);
    }

    /**
     * 更改状态
     */
    public function setStatus($status)
    {
        if ($this->checkCashierOnline($this['shop_user_id'])) {
            $this->error = '当前人员未交班，请先交班';
            return false;
        }
        if ($this['is_super'] == 1) {
            $this->error = '超级管理员不能被禁用';
            return false;
        }
        // 删除收银机缓存
        Cache::tag('cashier')->clear();
        //
        (new self([], 0))->where('shop_user_id', $this['shop_user_id'])->find()?->save(['is_status' => $status]);
        return $this->save([
            'is_status' => $status
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
            'shop_user_id' => $shop_user_id,
            'cashier_online' => 1,
        ];
        return self::where($where)->count() > 0;
    }
}
