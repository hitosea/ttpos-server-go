<?php

namespace app\admin\model\admin;

use hg\apidoc\annotation as Apidoc;
use app\common\model\admin\LoginLog;
use app\common\model\admin\User as UserModel;

/**
 * 超管后台用户模型
 */
class User extends UserModel
{
    /**
     * 超管后台用户登录
     */
    public function login($data)
    {
        // 验证用户名密码是否正确
        $user = self::withTrashed()->whereRaw('BINARY username = :username', ['username' => $data['username'] ?? ''])->order('admin_user_id', 'desc')->order('delete_time')->find();
        if (!$user) {
            LoginLog::add($data['username'] ?? '', \request()->ip(), '登录失败', 0);
            $this->error = '账号或密码错误';
            return false;
        }
        
        // 验证密码（支持 MD5 和 bcrypt）
        list($isValid, $needUpgrade) = verify_password($data['password'] ?? '', $user->password);

        if (!$isValid) {
            LoginLog::add($data['username'] ?? '', \request()->ip(), '登录失败', 0);
            $this->error = '账号或密码错误';
            return false;
        }
        
        // 如果需要升级，异步升级为 bcrypt
        // 注意：超管表在 saas 库中
        if ($needUpgrade) {
            upgrade_password_async($user['admin_user_id'], 'ttpos_admin_user', 'admin_user_id', 'password', $data['password'] ?? '', 0);
        }
        
        if ($user->delete_time) {
            LoginLog::add($data['username'] ?? '', \request()->ip(), '登录失败', 0);
            $this->error = '账号已删除, 请联系管理员';
            return false;
        }
        if (!$user->status) {
            LoginLog::add($data['username'] ?? '', \request()->ip(), '登录失败', 0);
            $this->error = '账号已禁用, 请联系管理员';
            return false;
        }
        // 保存登录状态
        $user['token'] = signToken($user['admin_user_id'], 'admin', '', md5($user->password));
        // 记录到日志
        LoginLog::add($user->username, \request()->ip(), '登录成功', $user->admin_user_id);
        //
        return $user;
    }

    /**
     * 超管用户信息
     */
    public static function detail($admin_user_id)
    {
        $model = new static();
        return $model::find($admin_user_id);
    }

    /**
     * 更新当前管理员信息
     */
    public function renew($data)
    {
        // 验证旧密码（支持 MD5 和 bcrypt）
        list($isValid, $needUpgrade) = verify_password($data['oldPass'] ?? '', $this->password);
        if (!$isValid) {
            $this->error = '原密码不正确';
            return false;
        }
        
        // 检查新密码是否与旧密码相同
        list($isSame, $_) = verify_password($data['pass'], $this->password);
        if ($isSame) {
            $this->error = '新密码不能与旧密码相同';
            return false;
        }
        
        // 使用 bcrypt 加密新密码
        return $this->save([
            'password' => hash_password_bcrypt($data['pass'])
        ]);
    }

    /**
     * 获取用户信息
     */
    public static function getUser($data)
    {
        return (new static())->where(['admin_user_id' => $data['uid']])->find();
    }

    /**
     * 获取用户列表
     * @Apidoc\withoutField("password")
     * @Apidoc\AddField("userRole",type="array", default="[]", desc="角色列表")
     */
    public function getList($param)
    {
        $keyword = $param['keyword'] ?? '';
        //
        return self::with(['userRole.role'])
            ->field('admin_user_id, username as user_name, phone, real_name, status, create_time')
            ->where('is_super', 0)
            ->when($keyword, function ($q) use ($keyword) {
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('username', $keyword);
                    $qq->orLike('real_name', $keyword);
                    $qq->orLike('admin_user_id', $keyword);
                });
            })
            ->order('create_time', 'desc')
            ->paginate($param);
    }

    /**
     * 添加用户
     */
    public function add($data)
    {
        $this->startTrans();
        try {
            // 使用 bcrypt 加密密码
            $res = self::create([
                'username' => trim($data['user_name']),
                'phone' => trim($data['phone']),
                'password' => hash_password_bcrypt($data['password']),
                'real_name' => trim($data['real_name']),
                'is_super' => 0,
            ]);
            $add_arr = [];
            $model = new UserRole();
            foreach ($data['role_id'] as $val) {
                $add_arr[] = [
                    'admin_user_id' => $res['admin_user_id'],
                    'role_id' => $val
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

    /**
     * 编辑用户
     */
    public function edit($data)
    {
        $this->startTrans();
        try {
            $where = ['admin_user_id' => $data['admin_user_id']];
            $arr = [
                'username' => $data['user_name'],
                'phone' => trim($data['phone']),
                'real_name' => $data['real_name'],
            ];
            
            // 如果提供了新密码，使用 bcrypt 加密
            if (!empty($data['password'])) {
                $arr['password'] = hash_password_bcrypt($data['password']);
            }
            self::update($arr, $where);
            //
            $model = new UserRole();
            UserRole::destroy($where);
            $add_arr = [];
            foreach ($data['role_id'] as $val) {
                $add_arr[] = [
                    'admin_user_id' => $data['admin_user_id'],
                    'role_id' => $val,
                ];
            }
            $model->saveAll($add_arr);
            // 事务提交
            $this->commit();
            //
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 删除用户
     */
    public function del($admin_user_id)
    {
        $userToDelete = self::find($admin_user_id);
        if (!$userToDelete) {
            $this->error = '用户不存在';
            return false;
        }
        $userToDelete->delete();
        //
        return UserRole::destroy(['admin_user_id' => $admin_user_id]);
    }
}
