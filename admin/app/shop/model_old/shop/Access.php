<?php

namespace app\shop\model_old\shop;

use app\common\model_old\shop\Access as AccessModel;

/**
 * Class Access
 *  商家用户权限模型
 */
class Access extends AccessModel
{
    /**
     * 获取商家后台路由
     */
    public function formatShopMenu($menus)
    {
        return $this->getRouteMenu2($menus, AccessModel::SHOP_ROUTE_NAME);
    }
}
