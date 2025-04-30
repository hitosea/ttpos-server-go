<?php

namespace app\common\model_old\admin;

use app\common\model_old\BaseModel;


/**
 * Class Access
 *  商家用户权限模型
 */
class Access extends BaseModel
{

    protected $name = 'admin_access';
    protected $pk = 'id';

    // 路由筛选名称
    const SHOP_ROUTE_NAME = '管理后台';

    /*
     * 获取所有权限
     */
    protected static function getAll($isShow = 1)
    {
        $model = (new static)::withoutGlobalScope();
        if ($isShow != -1) {
            $data = $model->where('is_show', '=', $isShow)
                ->order(['sort' => 'asc', 'create_time' => 'asc'])
                ->select();
        } else {
            $data = $model->order(['sort' => 'asc', 'create_time' => 'asc'])
                ->select();
        }

        return $data ? $data->toArray() : [];
    }


    /**
     * 权限信息
     */
    public static function detail($where)
    {
        if (is_array($where)) {
            return static::where($where)->find();
        } else {
            return static::where('id', '=', $where)->find();
        }
    }

    /**
     * 获取权限url集
     */
    public static function getAccessUrls($accessIds, $user_type)
    {
        $urls = [];
        foreach (static::getAll(1, $user_type) as $item) {
            in_array($item['id'], $accessIds) && $urls[] = $item['path'];
        }
        return $urls;
    }

    /**
     * 获取后台路径权限url集
     */
    public static function getApiAccessUrls($accessIds)
    {
        $urls = [];
        $aIds = [];
        $all = static::getAll(1);
        foreach ($accessIds as $accessId) {
            $aIds = array_merge($aIds, self::findParents($all, $accessId));
        }
        foreach (static::getAll(1) as $item) {
            in_array($item['id'], $aIds) && $urls[] = $item['api_path'];
        }
        return $urls;
    }

    /**
     * 获取权限url集
     */
    public static function getAccessList($accessIds, $user_type)
    {
        $model = (new static)::withoutGlobalScope();
        if ($user_type == 1) {
            $model = $model->where('is_supplier', '=', 1);
        }
        return $model->where('id', 'in', $accessIds)
            ->order(['sort' => 'asc', 'create_time' => 'asc'])
            ->select();
    }

    /**
     * 根据名称获取菜单的子菜单
     * @param array $menus
     * @param string $name
     * @return array
     */
    public function getRouteMenu(array $menus, string $name): array
    {
        foreach ($menus as $menu) {
            if (!isset($menu['children'])) {
                continue;
            }
            $adminMenu = array_filter($menu['children'], function ($val) use ($name) {
                return $val['name'] == $name;
            });
            if (!empty($adminMenu)) {
                return reset($adminMenu)['children'];
            }
        }
        return [];
    }

    /**
     * 获取权限
     * @return array
     */
    public function getPermission($name, $user)
    {
        if ($user['is_super'] == 1) {
            $menus = self::getList($user['user_type']);
        } else {
            $menus = self::getCashierListByUser($user['shop_user_id'], $user['user_type']);
            foreach ($menus as $key => $val) {
                if (!isset($val['children'][0]['path'])) {
                    continue;
                }
                if ($val['redirect_name'] != $val['children'][0]['path']) {
                    $menus[$key]['redirect_name'] = $menus[$key]['children'][0]['path'];
                }
            }
        }
        return $this->getRouteMenu($menus, $name);
    }

    /**
     * 获取所有父级
     */
    public static function findParents($array, $childId, $parents = array())
    {
        foreach ($array as $item) {
            if ($item['id'] == $childId) {
                $parents[] = $item['id'];
                if ($item['parent_id'] != 0) {
                    $parents = self::findParents($array, $item['parent_id'], $parents);
                }
            }
        }
        return $parents;
    }

    /**
     * 获取后台用户权限列表
     */
    public function getListByUser($shop_user_id, $user_type)
    {
        // 获取当前用户的角色集
        $roleIds = UserRole::getRoleIds($shop_user_id);
        // 根据已分配的权限
        $accessIds = RoleAccess::getAccessIds($roleIds);
        $aIds = [];
        $all = static::getAll(1);
        foreach ($accessIds as $accessId) {
            $aIds = array_merge($aIds, self::findParents($all, $accessId));
        }
        // 获取当前角色所有权限链接
        $menus_list = self::getAccessList($aIds, $user_type);
        // 格式化
        return $this->formatTreeData($menus_list, 101);
    }

    /**
     * 获取收银机权限列表
     */
    public function getCashierListByUser($shop_user_id, $user_type)
    {
        // 获取当前用户的角色集
        $roleIds = UserRole::getRoleIds($shop_user_id);
        // 根据已分配的权限
        $accessIds = RoleAccess::getAccessIds($roleIds);
        // 获取当前角色所有权限链接
        $menus_list = self::getCashierAccessList($accessIds, $user_type);
        // 格式化
        return $this->formatTreeData($menus_list, 0);
    }

    /**
     * 循环获取分类
     */
    private function formatTreeData($all, $parent_id = 0)
    {
        $tree = array();
        foreach ($all as $k => $v) {
            if ($v['parent_id'] == $parent_id) {
                //父亲找到儿子
                $v['children'] = $this->formatTreeData($all, $v['id']);
                $tree[] = $v;
            }
        }
        return $tree;
    }

    /**
     * 获取权限列表
     */
    public function getList()
    {
        $all = static::getAll(1);
        $res = $this->recursiveMenuArray($all, 0);
        return array_values($this->foo($res));
    }

    /**
     * 递归获取获取分类
     */
    private function recursiveMenuArray($data, $pid)
    {
        $re_data = [];
        foreach ($data as $key => $value) {
            if ($value['parent_id'] == $pid) {
                $re_data[$value['id']] = $value;
                $re_data[$value['id']]['children'] = $this->recursiveMenuArray($data, $value['id']);
            } else {
                continue;
            }
        }
        return $re_data;
    }

    /**
     * 格式化递归数组下标
     */
    private function foo(&$ar)
    {
        if (!is_array($ar)) return;
        foreach ($ar as $k => &$v) {
            if (is_array($v)) $this->foo($v);
            if ($k == 'children') $v = array_values($v);
        }
        return $ar;
    }

    /**
     * 获取商家后台路由
     */
    public function formatShopMenu($menus)
    {
        return $this->getRouteMenu($menus, self::SHOP_ROUTE_NAME);
    }
}
