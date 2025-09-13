<?php

namespace app\common\model\shop;

use app\common\model\app\App;
use app\common\model\BaseModel;

/**
 * 商家用户权限模型
 */
class Access extends BaseModel
{
    protected $name = 'access';
    protected $pk = 'id';

    /**
     * 追加属性
     */
    protected $append = ['access_id', 'parent_id'];

    // 路由筛选名称
    const SHOP_ROUTE_NAME = '管理后台';
    const CASHIER_ROUTE_NAME = '收银机';
    const ASSISTANT_ROUTE_NAME = '点餐助手';

    /**
     * 兼容字段
     */
    public function getAccessIdAttr()
    {
        return $this->uuid ?: 0;
    }

    /**
     * 兼容字段
     */
    public function getParentIdAttr()
    {
        return $this->parent_uuid ?: 0;
    }

    /*
     * 获取所有权限
     */
    protected static function getAll($isShow = 1, $user_type = 0, $supplier = '')
    {
        $model = (new static)::withoutGlobalScope();
        if ($user_type == 1) {
            if ($supplier && $supplier['category_set'] == 10) {
                $model = $model->where('path', 'not in', ['/product/takeaway/category/index', '/product/store/category/index']);
            }
            $model = $model->where('is_supplier', '=', 1);
        }
        $enableErp = App::detail(self::$app_id)->isEnableErp();
        if ($enableErp) {
            $model = $model->where('path', 'not in', [
                '/purchase/order/index', 
                '/purchase/supplier/index',
                '/inventory/wastage/index',
                '/inventory/statement/index',
            ]);
        }
        if ($isShow != -1) {
            $data = $model->where('is_show', '=', $isShow)
                ->order(['sort' => 'asc', 'create_time' => 'asc'])
                ->select();
        } else {
            $data = $model
                ->order(['sort' => 'asc', 'create_time' => 'asc'])
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
            return static::where('uuid', '=', $where)->find();
        }
    }

    /**
     * 获取权限url集
     */
    public static function getAccessUrls($accessIds, $user_type, $supplier)
    {
        $urls = [];
        foreach (static::getAll(1, $user_type, $supplier) as $item) {
            in_array($item['uuid'], $accessIds) && $urls[] = $item['path'];
        }
        return $urls;
    }

    /**
     * 获取后台路径权限url集
     */
    public static function getApiAccessUrls($accessIds, $user_type, $supplier)
    {
        $urls = [];
        foreach (static::getAll(1, $user_type, $supplier) as $item) {
            in_array($item['uuid'], $accessIds) && $urls[] = $item['api_path'];
        }
        return $urls;
    }

    /**
     * 获取权限url集
     */
    public static function getAccessList($accessIds, $user_type, $supplier)
    {
        $model = (new static)::withoutGlobalScope();
        if ($user_type == 1) {
            if ($supplier['category_set'] == 10) {
                $model = $model->where('path', 'not in', ['/product/takeaway/category/index', '/product/store/category/index']);
            }
            $model = $model->where('is_supplier', '=', 1);
        }
        return $model->where('uuid', 'in', $accessIds)
            ->order(['sort' => 'asc', 'create_time' => 'asc'])
            ->select()?->toArray() ?: [];
    }

    /**
     * 获取收银机权限url集
     */
    public static function getCashierAccessList($accessIds, $user_type, $supplier)
    {
        $model = (new static)::withoutGlobalScope();
        return $model->where('uuid', 'in', $accessIds)
            ->order(['sort' => 'asc', 'create_time' => 'asc'])
            ->select()?->toArray() ?: [];
    }


    /**
     * 通过插件分类id查询
     */
    public static function getListByPlusCategoryId($category_id)
    {
        $model = new static();
        return $model::withoutGlobalScope()->where('plus_category_id', '=', $category_id)
            ->where('is_show', '=', 1)
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
     * 根据名称获取菜单的子菜单（v1.0.9）
     * @param array $menus
     * @param string $name
     * @return array
     */
    public function getRouteMenu2(array $menus, string $name): array
    {
        $adminMenu = array_filter($menus, function ($val) use ($name) {
            return $val['name'] == $name;
        });
        if (!empty($adminMenu)) {
            return reset($adminMenu)['children'];
        }
        return [];
    }

    /**
     * 获取权限
     * @return array
     */
    public function getPermission($name, $user, $supplier)
    {
        if ($user['is_super'] == 1) {
            $menus = self::getList($user['user_type'], $supplier);
        } else {
            $menus = self::getCashierListByUser($user['uuid'], $user['user_type'], $supplier);
            foreach ($menus as $key => $val) {
                if (!isset($val['children'][0]['path'])) {
                    continue;
                }
                if ($val['redirect_name'] != $val['children'][0]['path']) {
                    $menus[$key]['redirect_name'] = $menus[$key]['children'][0]['path'];
                }
            }
        }
        return $this->getRouteMenu2($menus, $name);
    }

    /**
     * 获取后台用户权限列表
     */
    public function getListByUser($shop_user_id, $user_type, $supplier)
    {
        // 获取当前用户的角色集
        $roleIds = UserRole::getRoleIds($shop_user_id);
        // 根据已分配的权限
        $accessIds = RoleAccess::getAccessIds($roleIds);
        // 获取当前角色所有权限链接
        $menus_list = self::getAccessList($accessIds, $user_type, $supplier);
        $menus_list = $this->recursiveMenuArray($menus_list, 0);
        $menus_list = array_values($this->foo($menus_list));
        return $menus_list;
    }

    /**
     * 获取收银机权限列表
     */
    public function getCashierListByUser($shop_user_id, $user_type, $supplier)
    {
        // 获取当前用户的角色集
        $roleIds = UserRole::getRoleIds($shop_user_id);
        // 根据已分配的权限
        $accessIds = RoleAccess::getAccessIds($roleIds);
        // 获取当前角色所有权限链接
        $menus_list = self::getCashierAccessList($accessIds, $user_type, $supplier);
        $menus_list = $this->recursiveMenuArray($menus_list, 0);
        $menus_list = array_values($this->foo($menus_list));
        return $menus_list;
    }

    /**
     * 循环获取分类
     */
    private function formatTreeData($all, $parent_id = 0)
    {
        $tree = array();
        foreach ($all as $k => $v) {
            if ($v['parent_uuid'] == $parent_id) {
                //父亲找到儿子
                $v['children'] = $this->formatTreeData($all, $v['uuid']);
                $tree[] = $v;
            }
        }
        return $tree;
    }

    /**
     * 获取权限列表
     */
    public function getList($user_type = 0, $supplier = "")
    {
        $all = static::getAll(1, $user_type, $supplier);
        $res = $this->recursiveMenuArray($all, 0);
        return array_values($this->foo($res));
    }

    /**
     * 移除自定义权限
     *
     * @param [type] $array
     * @return array
     */
    public function removeCustomAccess(&$array)
    {
        $licenses = request()->licenses;
        if (isset($licenses['sale']) && $licenses['sale'] == 1) {
            return $array;
        }

        // 暂时去掉外卖管理
        $array = array_filter($array, function ($value) {
            return (!isset($value['uuid']) || ($value['uuid'] != 1711006072 && $value['uuid'] != 1711009130))
                && ($value['uuid'] != 1626688443);
        });

        $array = array_values($array);

        return $array;
    }

    /**
     * 递归获取获取分类
     */
    private function recursiveMenuArray($data, $pid)
    {
        $re_data = [];
        $licenses = request()->licenses;
        foreach ($data as $key => $value) {
            // 暂时去掉外卖管理
            if ($value['uuid'] == 1626688443) {
                continue;
            }
            // 暂时去掉收银交班权限
            if ($value['uuid'] == 1704881155) {
                continue;
            }
            // 授权无进销存权限
            if (isset($licenses['sale']) && $licenses['sale'] == 0) {
                if ($value['uuid'] == 1711006072 || $value['uuid'] == 1711009130) {
                    continue;
                }
            }
            // 授权无会员权限
            if (isset($licenses['is_open_member']) && $licenses['is_open_member'] == 0) {
                if ($value['uuid'] == 1636183779 || $value['uuid'] == 1704881218) {
                    continue;
                }
            }
            // 授权无平板点餐权限
            if (isset($licenses['is_open_tablet']) && $licenses['is_open_tablet'] == 0) {
                if ($value['uuid'] == 87) {
                    continue;
                }
            }
            // 授权无H5点餐权限
            if (isset($licenses['is_open_h5']) && $licenses['is_open_h5'] == 0) {
                if ($value['uuid'] == 1724220505) {
                    continue;
                }
            }
            // 授权无点餐助手权限
            if (isset($licenses['is_open_assistant']) && $licenses['is_open_assistant'] == 0) {
                if ($value['uuid'] == 1720753338) {
                    continue;
                }
            }
            // 授权无后厨权限
            if (isset($licenses['is_open_kitchen_kds']) && $licenses['is_open_kitchen_kds'] == 0) {
                if ($value['uuid'] == 88) {
                    continue;
                }
            }
            // 授权无自助餐权限
            if (isset($licenses['is_open_buffet']) && $licenses['is_open_buffet'] == 0) {
                if ($value['uuid'] == 1708671616) {
                    continue;
                }
            }
            // 授权无扫码点餐接单权限
            if (isset($licenses['is_open_h5_order']) && $licenses['is_open_h5_order'] == 0) {
                if ($value['uuid'] == 1724320522) {
                    continue;
                }
            }
            // 授权无优惠券权限
            if (($licenses['is_open_coupon'] ?? 0) == 0) {
                if ($value['uuid'] == 1731155609 || $value['uuid'] == 1731155610) {
                    continue;
                }
            }
            // 授权无营销活动权限
            if (($licenses['is_open_marketing'] ?? 0) == 0) {
                if ($value['uuid'] == 1731155710 || $value['uuid'] == 1731155711) {
                    continue;
                }
            }
            // 授权无会员卡权限
            if (($licenses['is_open_coupon'] ?? 0) == 0 && ($licenses['is_open_marketing'] ?? 0) == 0) {
                if ($value['uuid'] == 1731155608) {
                    continue;
                }
            }
            // 未开启外送，看不到外送订单
            if (($licenses['is_open_delivery'] ?? 0) == 0) {
                if ($value['uuid'] == 1741590679 || $value['uuid'] == 1750930155 || $value['uuid'] == 1752716650) {
                    continue;
                }
            }
            //
            if ($value['parent_uuid'] == $pid) {
                $re_data[$value['uuid']] = $value;
                $re_data[$value['uuid']]['children'] = $this->recursiveMenuArray($data, $value['uuid']);
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
}
