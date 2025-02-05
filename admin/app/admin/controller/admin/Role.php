<?php

namespace app\admin\controller\admin;

use hg\apidoc\annotation as Apidoc;
use app\admin\controller\Controller;
use app\admin\validate\AdminRoleValidate;
use app\admin\model\admin\Role as RoleModel;
use app\common\model\admin\Access as AccessModel;

/**
 * 角色管理
 * @Apidoc\Group("user")
 * @Apidoc\Sort(2)
 */
class Role extends Controller
{
    /**
     * @Apidoc\Title("列表")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/api/admin/admin.role/index")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\admin\model\admin\Role\getTreeData", desc="角色列表")
     */
    public function index()
    {
        $model = new RoleModel();
        $list = $model->getTreeData();
        return $this->renderSuccess('', compact('list'));
    }

    /**
     * @Apidoc\Title("新增")
     * @Apidoc\Desc("get请求是获取新增所有的信息，post请求是提交添加")
     * @Apidoc\Method("GET,POST")
     * @Apidoc\Url("/api/admin/admin.role/add")
     * @Apidoc\Param("role_name", type="string", require=true, desc="角色名称")
     * @Apidoc\Param("access_id", type="array", require=true, desc="权限id数组")
     */
    public function add(AdminRoleValidate $validate)
    {
        if ($this->request->isGet()) {
            $menu = (new AccessModel())->getList();
            return $this->renderSuccess('', compact('menu'));
        }
        $data = $validate->goCheck('add');
        $model = new RoleModel();
        if ($model->add($data)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("修改")
     * @Apidoc\Desc("get请求是获取新增所有的信息，post请求是提交添加")
     * @Apidoc\Method("GET,POST")
     * @Apidoc\Url ("/api/admin/admin.role/edit")
     * @Apidoc\Param("id", type="int", require=true, desc="角色id")
     * @Apidoc\Param("role_name", type="string", require=true, desc="角色名称")
     * @Apidoc\Param("access_id", type="array", require=true, desc="权限id数组")
     * @Apidoc\Returned()
     */
    public function edit(AdminRoleValidate $validate)
    {
        if ($this->request->isGet()) {
            $data = $validate->goCheck('id');
            $menu = (new AccessModel())->getList();
            $model = RoleModel::detail($data['id']);
            if (!$model) {
                return $this->renderError('角色不存在');
            }
            $select_menu = array_column($model->toArray()['access'], 'access_id');
            $roleList = $model->getTreeData();
            return $this->renderSuccess('', compact('model', 'roleList', 'menu', 'select_menu'));
        }
        //
        $data = $validate->goCheck('edit');
        $model = RoleModel::detail($data['id']);
        if ($model->edit($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/api/admin/admin.role/delete")
     * @Apidoc\Param("id", type="int", require=true, desc="角色id")
     * @Apidoc\Returned()
     */
    public function delete(AdminRoleValidate $validate)
    {
        $data = $validate->goCheck('id');
        $model = new RoleModel();
        if ($model->del($data['id'])) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }
}
