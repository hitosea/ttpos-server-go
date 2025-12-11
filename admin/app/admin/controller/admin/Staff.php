<?php

namespace app\admin\controller\admin;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\validate\StaffValidate;
use app\admin\model\admin\Staff as StaffModel;

/**
 * 统一账号管理
 * @Apidoc\Group("staff")
 * @Apidoc\Sort(1)
 */
class Staff extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/api/admin/admin.staff/index")
     * @Apidoc\Param("keyword", type="string", require=false, default="", desc="搜索关键词（邮箱、手机号、员工ID）")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\admin\model\admin\Staff\getList", desc="统一账号列表")
     */
    public function index()
    {
        $model = new StaffModel();
        $list = $model->getList($this->postData());
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.staff/add")
     * @Apidoc\Param("email", type="string", require=true, default="", desc="邮箱（全平台唯一）")
     * @Apidoc\Param("phone", type="string", require=false, default="", desc="手机号（全平台唯一，允许空字符串）")
     * @Apidoc\Param("real_name", type="string", require=false, default="", desc="姓名")
     * @Apidoc\Param("password", type="string", require=true, default="", desc="登录密码")
     * @Apidoc\Param("confirm_password", type="string", require=true, default="", desc="确认密码")
     * @Apidoc\Param("company_uuid", type="int", require=true, desc="关联的门店UUID")
     * @Apidoc\Param("role_uuids", type="array", require=false, desc="在该门店的角色UUID列表")
     */
    public function add(StaffValidate $validate)
    {
        $param = $validate->goCheck('add');
        $model = new StaffModel();
        if ($model->add($param)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.staff/edit")
     * @Apidoc\Param("uuid", type="int", require=true, desc="员工UUID")
     * @Apidoc\Param("email", type="string", require=false, default="", desc="邮箱（全平台唯一）")
     * @Apidoc\Param("phone", type="string", require=false, default="", desc="手机号（全平台唯一，允许空字符串）")
     * @Apidoc\Param("real_name", type="string", require=false, default="", desc="姓名")
     * @Apidoc\Param("password", type="string", require=false, default="", desc="登录密码（可选）")
     * @Apidoc\Param("confirm_password", type="string", require=false, default="", desc="确认密码（可选）")
     * @Apidoc\Param("company_list", type="array", require=false, desc="关联的门店列表（可选，用于编辑时更新门店关联）")
     */
    public function edit(StaffValidate $validate)
    {
        $param = $validate->goCheck('edit');
        $model = new StaffModel();
        if ($model->edit($param)) {
            return $this->renderSuccess('编辑成功');
        }
        return $this->renderError($model->getError() ?: '编辑失败');
    }

    /**
     * @Apidoc\Title("启用禁用状态")
     * @Apidoc\Method("post")
     * @Apidoc\Url("/api/admin/admin.staff/updateStatus")
     * @Apidoc\Param("uuid", type="int", require=true, default="", desc="员工UUID")
     */
    public function updateStatus(StaffValidate $validate)
    {
        $param = $validate->goCheck('uuid');
        $model = StaffModel::detail($param['uuid']);
        if (!$model->updateStatus()) {
            return $this->renderError('操作失败');
        }
        return $this->renderSuccess('操作成功');
    }
}
