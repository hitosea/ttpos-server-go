<?php

namespace app\admin\controller;

use hg\apidoc\annotation as Apidoc;
use app\admin\model\Access as AccessModel;
use app\common\model\admin\User as UserModel;

/**
 * 权限相关
 * @Apidoc\Group("base")
 * @Apidoc\Sort(2)
 */
class Access extends Controller
{
    /**
     * @Apidoc\Title("权限列表")
     * @Apidoc\Method("get")
     * @Apidoc\Url("/api/admin/access/index")
     * NotParse
     */
    public function index()
    {
        $model = new AccessModel;
        $list = $model->getAdminList();
        return $this->renderSuccess('', $list);
    }

    /**
     * @Apidoc\Title("获取菜单权限")
     * @Apidoc\Method("GET")
     * @Apidoc\Url("/api/admin/access/getRoleList")
     */
    public function getRoleList()
    {
        $user = UserModel::find($this->admin['user']['admin_user_id']);
        //
        if ($user['is_super'] == 1) {
            $menus = (new AccessModel())->getList();
        } else {
            $model = new AccessModel();
            $menus = $model->getListByUser($user->admin_user_id, $user['is_super']);
        }
        //
        if (($menus[0]['id'] ?? 0) == 101) {
            $menus = $menus[0]['children'] ?? [];
        }
        return $this->renderSuccess('', compact('menus'));
    }

    /**
     * 添加权限
     */
    public function add()
    {
        $model = new AccessModel;
        $data = $this->postData();

        if ($model->add($data)) {
            return $this->renderSuccess('添加成功', compact('model'));
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * 更新权限
     */
    public function edit()
    {
        $data = $this->postData();
        // 权限详情
        $model = AccessModel::detail($data['access_id']);
        // 更新记录
        if ($model->edit($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * 删除权限
     */
    public function delete($access_id)
    {
        $model = new  AccessModel();
        $num = $model->getChildCount(['parent_id' => $access_id]);
        if ($num > 0) {
            return $this->renderError('当前菜单下存在子权限，请先删除');
        }
        if ($model->remove($access_id)) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * 权限状态
     */
    public function status($access_id, $status)
    {
        $model = AccessModel::detail($access_id);
        if ($model->status($status)) {
            return $this->renderSuccess('修改成功');
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }

    /**
     * 门店菜单状态
     */
    public function supplier($access_id, $status)
    {
        $model = AccessModel::detail($access_id);
        if ($model->supplier($status)) {
            return $this->renderSuccess('修改成功');
        }
        return $this->renderError($model->getError() ?: '修改失败');
    }
}
