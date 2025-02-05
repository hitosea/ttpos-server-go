<?php

namespace app\admin\model\admin;

use think\facade\Cache;
use app\common\model\admin\Role as RoleModel;
use app\admin\model\admin\UserRole as UserRoleModel;

/**
 * 角色模型
 */
class Role extends RoleModel
{
    /**
     * 获取所有角色列表
     */
    public function getTreeData()
    {
        $all = $this->getAll();
        return $this->formatTreeData($all);
    }

    /**
     * 获取所有角色
     */
    private function getAll()
    {
        $data = $this->order(['create_time' => 'desc'])->select();
        return $data ? $data->toArray() : [];
    }

    /**
     * 获取权限列表
     */
    private function formatTreeData(&$all, $parent_id = 0, $deep = 1)
    {
        static $tempTreeArr = [];
        foreach ($all as $key => $val) {
            // 判断是否被使用
            $val['is_used'] = UserRoleModel::where(['role_id' => $val['id']])->count() > 0 ? 1 : 0;
            // 根据角色深度处理名称前缀
            $val['role_name_h1'] = $this->htmlPrefix($deep) . $val['role_name'];
            $tempTreeArr[] = $val;
        }
        return $tempTreeArr;
    }

    /**
     * 角色名称 html格式前缀
     */
    private function htmlPrefix($deep)
    {
        // 根据角色深度处理名称前缀
        $prefix = '';
        if ($deep > 1) {
            for ($i = 1; $i <= $deep - 1; $i++) {
                $prefix .= '   ├ ';
            }
            $prefix .= ' ';
        }
        return $prefix;
    }

    /**
     * 添加/编辑 - 自动识别添加或编辑
     *
     * @param array $data
     * @return bool
     */
    public function saveFromMigrate(array $data)
    {
        $this->startTrans();
        try {
            $roleAccessModel = new RoleAccess();
            // 检查是否存在
            $role = self::where('role_name', $data['role_name'])->where('app_id', $data['app_id'])->find();
            if ($role) {
                $role->save([
                    'role_name' => $data['role_name'],
                    'sort' => max($data['sort'] ?? 1, 1),
                ]);
                // 先删后增
                $roleAccessModel->where(['role_id' => $role['id']])->delete();
            } else {
                $role = self::create([
                    'role_name' => $data['role_name'],
                    'sort' => max($data['sort'] ?? 1, 1),
                    'app_id' => $data['app_id']
                ]);
            }

            $roleAccessData = array_map(function ($accessId) use ($role, $data) {
                return [
                    'role_id' => $role['id'],
                    'access_id' => $accessId,
                    'app_id' => $data['app_id']
                ];
            }, $data['access_id']);

            $roleAccessModel->saveAll($roleAccessData);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 添加
     *
     * @param [type] $data
     * @return bool
     */
    public function add($data)
    {
        $this->startTrans();
        try {
            $res = self::create([
                'role_name' => $data['role_name'],
                'sort' => max($data['sort'] ?? 1, 1),
            ]);
            $model = new RoleAccess();
            $roleAccess = [];
            foreach ($data['access_id'] as $val) {
                $roleAccess[] = [
                    'role_id' => $res['id'],
                    'access_id' => $val,
                ];
            }
            $res = $model->saveAll($roleAccess);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 编辑
     * @param $data
     * @return bool
     */
    public function edit($data)
    {
        $this->startTrans();
        try {
            $this->save([
                'role_name' => $data['role_name'],
                'sort' => $data['sort'] ?? 0,
            ]);
            if (!isset($data['access_id'])) {
                $this->commit();
                return true;
            }
            $access_list = [];
            $access_model = new RoleAccess();
            $access_model->where(['role_id' => $data['id']])->delete();
            foreach ($data['access_id'] as $val) {
                $access_list[] = [
                    'role_id' => $data['id'],
                    'access_id' => $val,
                    'app_id' => self::$app_id
                ];
            }
            $access_model->saveAll($access_list);
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
     * 删除
     * @param $role_id
     * @return bool
     */
    public function del($role_id)
    {
        //如果角色下有用户，则不能删除
        if (UserRoleModel::getUserRoleCount($role_id) > 0) {
            $this->error = '当前角色下存在用户，不允许删除';
            return false;
        }
        $res = RoleAccess::where('role_id', $role_id)->delete();
        return $res && self::destroy(['id', '=', $role_id]);
    }
}
