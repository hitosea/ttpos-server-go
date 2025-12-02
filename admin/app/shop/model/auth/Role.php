<?php

namespace app\shop\model\auth;

use think\facade\Cache;
use app\common\service\websocket\Websocket;
use app\common\model\shop\Role as RoleModel;
use app\shop\model\auth\UserRole as UserRoleModel;

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
        $data = $this->field('*, name as role_name')->order(['create_time' => 'desc'])->where('delete_time', 0)->select();
        return $data ? $data->toArray() : [];
    }

    /**
     * 获取权限列表
     */
    private function formatTreeData(&$all, $parent_id = 0, $deep = 1)
    {
        static $tempTreeArr = [];
        foreach ($all as $key => $val) {
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
            $role = self::where('name', $data['name'])->find();
            if ($role) {
                $role->save([
                    'name' => $data['name'],
                    'sort' => max($data['sort'] ?? 1, 1),
                ]);
                // 先删后增
                $roleAccessModel->where(['uuid' => $role['uuid']])->delete();
            } else {
                $role = self::create([
                    'uuid' => createUuid(),
                    'name' => $data['name'],
                    'sort' => max($data['sort'] ?? 1, 1),
                ]);
            }

            $roleAccessData = array_map(function ($accessId) use ($role, $data) {
                return [
                    'role_uuid' => $role['uuid'],
                    'access_uuid' => $accessId,
                ];
            }, $data['access_uuid']);

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
                'uuid' => createUuid(),
                'name' => $data['role_name'],
                'sort' => max($data['sort'] ?? 1, 1),
            ]);
            $model = new RoleAccess();
            $roleAccess = [];
            foreach ($data['access_id'] as $val) {
                $roleAccess[] = [
                    'uuid' => createUuid(),
                    'role_uuid' => $res['uuid'],
                    'access_uuid' => $val,
                ];
            }
            $model->saveAll($roleAccess);
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
        $role_id = $data['role_id'] ?? 0;
        $this->startTrans();
        try {
            $this->save([
                'name' => $data['role_name'],
                'sort' => $data['sort'],
            ]);
            if (!isset($data['access_id'])) {
                $this->commit();
                return true;
            }

            $access_list = [];
            $access_model = new RoleAccess();
            $access_model->destroy(function ($query) use ($role_id) {
                $query->where('role_uuid', $role_id);
            });

            foreach ($data['access_id'] as $val) {
                $access_list[] = [
                    'uuid' => createUuid(),
                    'role_uuid' => $role_id,
                    'access_uuid' => $val,
                ];
            }

            $access_model->saveAll($access_list);
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
                0,
                ['update_time' => time()]
            );
            //
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }


    public function del($role_id)
    {
        $role = self::where('uuid', $role_id)->find();
        if (!$role) {
            $this->error = '角色不存在';
            return false;
        }
        
        //如果角色下有用户，则不能删除
        if (UserRoleModel::getUserRoleCount($role['uuid']) > 0) {
            $this->error = '当前角色下存在用户，不允许删除';
            return false;
        }
        return $role->transaction(function () use ($role) {
            if (!RoleAccess::destroy(function ($query) use ($role) {
                $query->where('role_uuid', $role['uuid']);
            })) {
                return false;
            }
            return $role->delete();
        });
    }
}
