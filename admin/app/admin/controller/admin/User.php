<?php

namespace app\admin\controller\admin;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\validate\AdminUserValidate;
use app\admin\model\admin\User as AdminUserModel;

/**
 * 管理员列表
 * @Apidoc\Group("user")
 * @Apidoc\Sort(1)
 */
class User extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/admin.user/index")
     * @Apidoc\Param("keyword", type="string", require=true, default="", desc="用户名,姓名, 用户ID")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\admin\model\admin\User\getList", desc="管理员列表")
     */
    public function index()
    {
        $model = new AdminUserModel();
        $list = $model->getList($this->postData());
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.user/add")
     * @Apidoc\Param("user_name", type="string", require=true, default="", desc="邮箱")
     * @Apidoc\Param("phone", type="string", require=true, default="", desc="手机号（v1.0.7）")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="登录密码")
     * @Apidoc\Param("confirm_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Param("real_name", type="string", require=true, default="", desc="姓名")
     * @Apidoc\Param("role_id", type="array", require=true, desc="角色ids")
     */
    public function add(AdminUserValidate $validate)
    {
        $param = $validate->goCheck('add');
        $model = new AdminUserModel;
        if ($model->add($param)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.user/edit")
     * @Apidoc\Param("admin_user_id", type="int", require=true, desc="管理员id")
     * @Apidoc\Param("user_name", type="string", require=true, default="", desc="邮箱")
     * @Apidoc\Param("phone", type="string", require=true, default="", desc="手机号（v1.0.7）")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="登录密码")
     * @Apidoc\Param("confirm_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Param("real_name", type="string", require=true, default="", desc="姓名")
     * @Apidoc\Param("role_id", type="array", require=true, desc="角色ids")
     */
    public function edit(AdminUserValidate $validate)
    {
        $param = $validate->goCheck('edit');
        $model = new AdminUserModel;
        if ($model->edit($param)) {
            return $this->renderSuccess('编辑成功');
        }
        return $this->renderError($model->getError() ?: '编辑失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.user/delete")
     * @Apidoc\Param("admin_user_id", type="int", require=true, desc="管理员id")
     */
    public function delete(AdminUserValidate $validate)
    {
        $param = $validate->goCheck('id');
        if ($param['admin_user_id'] == $this->admin['user']['admin_user_id']) {
            return $this->renderError('不能删除当前登录账号');
        }
        $model = new AdminUserModel;
        if ($model->del($param['admin_user_id'])) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("启用禁用状态")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.user/updateStatus")
     * @Apidoc\Param("admin_user_id", type="int", require=true, default="", desc="管理员id")
     */
    public function updateStatus(AdminUserValidate $validate)
    {
        $param = $validate->goCheck('id');
        $model = AdminUserModel::detail($param['admin_user_id']);
        if (!$model->updateStatus()) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }
}
