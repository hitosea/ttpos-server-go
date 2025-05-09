<?php

namespace app\shop\controller\auth;

use app\shop\model\auth\Role;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\model\auth\User as UserModel;
use app\shop\model\auth\User as AuthUserModel;
use app\shop\model\shop\Access as AccessModel;
use app\common\model\settings\Setting as SettingModel;
use app\common\model\app\App as AppModel;

/**
 * 用户管理
 * @Apidoc\Group("shop_user")
 * @Apidoc\Sort(8)
 */
class User extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/auth.user/index")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\auth\User\getList", desc="管理员列表")
     * @Apidoc\Returned("roleList", type="array", ref="app\shop\model\auth\Role\getTreeData", desc="角色列表")
     */
    public function index()
    {
        $model = new UserModel();
        $list = $model->getList($this->postData());
        $roleModel = new Role();
        // 角色列表
        $roleList = $roleModel->getTreeData();
        return $this->renderSuccess('', compact('list', 'roleList'));
    }

    /**
     * 新增信息
     * @return \think\response\Json
     */
    public function addInfo()
    {
        $model = new Role();
        // 角色列表
        $roleList = $model->getTreeData();
        return $this->renderSuccess('', compact('roleList'));
    }

    /**
     * @Apidoc\Title("新增")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/auth.user/add")
     * @Apidoc\Param("phone", type="string", require=true, default="", desc="手机号（v1.0.6）")
     * @Apidoc\Param("user_name", type="string", require=true, default="001", desc="用户名")
     * @Apidoc\Param("role_id", type="array", require=true, desc="角色ids")
     * @Apidoc\Param("password", type="string", require=true, default="123456", desc="密码")
     * @Apidoc\Param("confirm_password", type="string", require=true, default="123456", desc="确认密码")
     * @Apidoc\Param("real_name", type="string", require=true, desc="真实姓名")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $data = $this->postData();
        $model = new UserModel();
        $password = $data['password'] ?? '';
        if (empty($data['real_name'])) {
            return $this->renderError('请输入姓名');
        }
        if (mb_strlen($data['real_name']) > 50) {
            return $this->renderError('姓名长度不能超过50个字符');
        }
        if (!isset($data['role_id']) || empty($data['role_id'])) {
            return $this->renderError('请选择所属角色');
        }
        if ($data['confirm_password'] != $password) {
            return $this->renderError('确认密码和登录密码不一致');
        }
        $model = new UserModel();
        if ($model->add($data, $this->store['user'])) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * 修改信息
     * @param $shop_user_id
     * @return \think\response\Json
     */
    public function editInfo($shop_user_id)
    {
        $info = UserModel::detail(['uuid' => $shop_user_id], ['UserRole']);

        $role_arr = array_column($info->toArray()['UserRole'], 'role_uuid');

        $model = new Role();
        // 角色列表
        $roleList = $model->getTreeData();
        return $this->renderSuccess('', compact('info', 'roleList', 'role_arr'));
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/auth.user/edit")
     * @Apidoc\Param("shop_user_id", type="int", require=true, desc="管理员id")
     * @Apidoc\Param("phone", type="string", require=true, default="", desc="手机号（v1.0.6）")
     * @Apidoc\Param("user_name", type="string", require=true, default="001", desc="用户名")
     * @Apidoc\Param("role_id", type="array", require=true, desc="角色ids")
     * @Apidoc\Param("password", type="string", require=true, default="123456", desc="密码")
     * @Apidoc\Param("confirm_password", type="string", require=true, default="123456", desc="确认密码")
     * @Apidoc\Param("real_name", type="string", require=true, desc="真实姓名")
     * @Apidoc\Returned()
     */
    public function edit($shop_user_id)
    {
        $data = $this->postData();
        if ($this->request->isGet()) {
            return $this->editInfo($shop_user_id);
        }
        $password = $data['password'] ?? '';
        $model = new UserModel();
        if (empty($data['real_name'])) {
            return $this->renderError('请输入姓名');
        }
        if (mb_strlen($data['real_name']) > 50) {
            return $this->renderError('姓名长度不能超过50个字符');
        }
        if (!isset($data['role_id'])) {
            return $this->renderError('请选择所属角色');
        }
        if (isset($password) && !empty($password)) {
            if (!isset($data['confirm_password'])) {
                return $this->renderError('请输入确认密码');
            } else {
                if ($data['confirm_password'] != $password) {
                    return $this->renderError('确认密码和登录密码不一致');
                }
            }
        }
        if (empty($password)) {
            if (isset($data['confirm_password']) && !empty($data['confirm_password'])) {
                return $this->renderError('请输入登录密码');
            }
        }
        // 更新记录
        $data['shop_user_id'] = $shop_user_id;
        if ($model->edit($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/auth.user/delete")
     * @Apidoc\Param("shop_user_id", type="int", require=true, desc="管理员id")
     * @Apidoc\Returned()
     */
    public function delete($shop_user_id)
    {
        $model = new UserModel();
        if ($model->del($shop_user_id, $this->store['user'])) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * 获取角色菜单信息
     */
    public function getRoleList()
    {
        $user = $this->store['user'];
        $supplier = $this->store['supplier'];
        $menus = [];
        $user_info = AuthUserModel::where('uuid', $user['uuid'])->find();
        if (empty($user_info)) {
            return $this->renderError('用户不存在');
        }

        if ($user_info['is_super'] == 1) {
            /** @var AccessModel $model  */
            $model = new AccessModel();
            $menus = $model->getList($user['user_type'], $supplier);
        } else {
            /** @var AccessModel $model  */
            $model = new AccessModel();
            $menus = $model->getListByUser($user['uuid'], $user['user_type'], $supplier);
            foreach ($menus as $key => $val) {
                if (!isset($val['children'][0]['path'])) {
                    continue;
                }
                if ($val['redirect_name'] != $val['children'][0]['path']) {
                    $menus[$key]['redirect_name'] = $menus[$key]['children'][0]['path'];
                }
            }
        }
        // 格式化菜单
        $menus = $model->formatShopMenu($menus);
        // 删除权限
        $menus = $model->removeCustomAccess($menus);
        // 订单管理
        $app = AppModel::where('uuid', request()->appId)->find();
        if ($app->old_company_id) {
            foreach ($menus as $key => $val) {
                if ($val['uuid'] == 1626506017) {
                    $menu = $menus[$key]['children'][0] ?? [];
                    if ($menu) {
                        $menu['name'] = "历史用餐订单";
                        $menu['path'] = "/store/history_order/index";
                        $menu['api_path'] = "/store/history_order/index";
                        $menu['id'] = 1734320550;
                        $menu['access_id'] = 1734320550;
                        $menu['uuid'] = 1734320550;
                        $menu['children'][0]['path'] = "/store/history_order/detail";
                        $menu['children'][0]['api_path'] = "/store/history_order/detail";
                        // 
                        $menuOne = $menus[$key]['children'][1] ?? [];
                        $menus[$key]['children'][1] = $menu;
                        if ($menuOne) {
                            $menus[$key]['children'][2] = $menuOne;
                        }
                    }
                }
            }
        }
        // 
        return $this->renderSuccess('', compact('menus'));
    }

    /**
     * 获取用户信息
     */
    public function getUserInfo()
    {
        $store = session('jjjshop_store');
        $user = [];
        if (!empty($store)) {
            $user = $store['user'];
        }
        // 商城名称
        $shop_name = SettingModel::getItem('store')['name'];
        //当前系统版本
        $version = get_version();
        //门店名称
        $supplier_name = $store['supplier']['name'];
        return $this->renderSuccess('', compact('supplier_name', 'user', 'shop_name', 'version'));
    }

    /**
     * @Apidoc\Title("权限状态")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/auth.user/setStatus")
     * @Apidoc\Param("shop_user_id", type="int", require=true, desc="管理员id")
     * @Apidoc\Param("status", type="int", require=true, desc="状态")
     * @Apidoc\Returned()
     */
    public function setStatus($shop_user_id, $status)
    {
        if (in_array($status, [0, 1]) === false) {
            return $this->renderError('状态参数错误');
        }
        /** @var UserModel $model */
        $model = UserModel::detail($shop_user_id);
        if (!$model) {
            return $this->renderError('用户不存在');
        }
        if ($model->setStatus($status)) {
            return $this->renderSuccess('修改成功');
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }
}
