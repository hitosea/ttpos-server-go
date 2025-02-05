<?php

namespace app\shop\controller\store\table;

use help\ArrayHelp;
use app\shop\service\CheckService;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\service\qrcode\TableService;
use app\shop\validate\TableImportsValidate;
use app\shop\model\store\Table as TableModel;
use app\shop\model\store\TableArea as TableAreaModel;
use app\shop\model\store\TableType as TableTypeModel;

/**
 * 桌码管理
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(6)
 */
class Table extends Controller
{
    /**
     * @Apidoc\Title("桌位列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/index")
     * @Apidoc\Param("table_no", type="string", require=true, default="", desc="桌位名称")
     * @Apidoc\Param("area_id", type="int", require=true, default=0, desc="区域id")
     * @Apidoc\Param("type_id", type="int", require=true, default=0, desc="类型id")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\store\Table\getList", desc="列表")
     * @Apidoc\Returned("area_list", type="array", ref="app\shop\model\store\TableArea\getAllList", desc="区域列表")
     * @Apidoc\Returned("type_list", type="array", ref="app\shop\model\store\TableType\getAllList", desc="类型列表")
     */
    public function index()
    {
        // 桌位列表
        $model = new TableModel;
        $list = $model->getList($this->postData(), $this->store['user']['shop_supplier_id']);
        // 区域列表
        $area_list = TableAreaModel::getAllList($this->store['user']['shop_supplier_id']);
        // 类型列表
        $type_list = TableTypeModel::getAllList($this->store['user']['shop_supplier_id']);
        return $this->renderSuccess('', compact('list', 'area_list', 'type_list'));
    }

    /**
     * @Apidoc\Title("添加")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/add")
     * @Apidoc\Param("table_no", type="string", require=true, default="", desc="桌位名称")
     * @Apidoc\Param("area_id", type="int", require=true, default=0, desc="区域id")
     * @Apidoc\Param("type_id", type="int", require=true, default=0, desc="类型id")
     * @Apidoc\Param("sort", type="int", require=true, default=0, desc="排序")
     * @Apidoc\Returned()
     */
    public function add()
    {
        $model = new TableModel;
        // 传过来的信息
        $data = $this->postData();
        $data['shop_supplier_id'] = $this->store['user']['shop_supplier_id'];
        // 新增记录
        if ($model->add($data)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败');
    }

    /**
     * @Apidoc\Title("编辑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/edit")
     * @Apidoc\Param("table_id", type="int", require=true, default=0, desc="桌位id")
     * @Apidoc\Param("table_no", type="string", require=true, default="", desc="桌位名称")
     * @Apidoc\Param("area_id", type="int", require=true, default=0, desc="区域id")
     * @Apidoc\Param("type_id", type="int", require=true, default=0, desc="类型id")
     * @Apidoc\Param("sort", type="int", require=true, default=0, desc="排序")
     * @Apidoc\Returned()
     */
    public function edit($table_id)
    {
        $model = TableModel::detail($table_id);
        //编辑店员的数据
        if ($model->edit($this->postData())) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败');
    }

    /**
     * @Apidoc\Title("删除")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/delete")
     * @Apidoc\Param("table_id", type="int", require=true, default=0, desc="桌位id")
     * @Apidoc\Returned()
     */
    public function delete($table_id)
    {
        /** @var TableModel $model */
        $model = TableModel::detail($table_id);
        if ($model->setDelete()) {
            return $this->renderSuccess('', '删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("批量删除（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/batchDelete")
     * @Apidoc\Param("ids", type="int", require=true, default=0, desc="桌位ids，多个英文逗号隔开")
     * @Apidoc\Returned()
     */
    public function batchDelete()
    {
        /** @var TableModel $model */
        $model = new TableModel;
        $data = $this->postData();
        $ids = $data['ids'] ?? '';
        if (!is_string($ids) || empty($ids)) {
            return $this->renderError('参数错误');
        }
        $ids = explode(',', $ids);
        if ($model->batchDelete($ids)) {
            return $this->renderSuccess('删除成功');
        }
        return $this->renderError($model->getError() ?: '删除失败');
    }

    /**
     * @Apidoc\Title("二维码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/qrcode")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="桌位id")
     * @Apidoc\Param("source", type="string", require=true, default="", desc="来源")
     * @Apidoc\Param("action", type="string", require=true, default="", desc="动作：get-获取/update-更新")
     * @Apidoc\Returned()
     */
    public function qrcode()
    {
        $data = $this->postData();
        $id = $data['id'] ?? 0;
        $source = $data['source'] ?? '';
        $action = $data['action'] ?? '';
        if ($id == 0 || empty($source) || empty($action)) {
            return $this->renderError('参数错误');
        }
        $qrcode = new TableService($id, $source, $action);
        return $this->renderSuccess('', $qrcode->getImage());
    }

    /**
     * @Apidoc\Title("批量二维码（v1.0.7）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/batchQrcode")
     * @Apidoc\Param("ids", type="string", require=true, default=0, desc="桌位ids，多个英文逗号隔开")
     * @Apidoc\Param("source", type="string", require=true, default="", desc="来源")
     * @Apidoc\Param("action", type="string", require=true, default="", desc="动作：get-获取/update-更新")
     * @Apidoc\Returned()
     */
    public function batchQrcode()
    {
        $data = $this->postData();
        $ids = explode(',', $data['ids']) ?? 0;
        $source = $data['source'] ?? '';
        $action = $data['action'] ?? '';
        if ($ids == 0 || !is_array($ids) || empty($source) || empty($action)) {
            return $this->renderError('参数错误');
        }
        $qrcode = new TableService(0, $source, $action);
        return $this->renderSuccess('', $qrcode->generateBatchQrCodes($ids));
    }

    /**
     * @Apidoc\Title("下载二维码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/download")
     * @Apidoc\Param("id", type="int", require=true, default=0, desc="桌位id")
     * @Apidoc\Param("source", type="string", require=true, default="", desc="来源")
     * @Apidoc\Returned()
     */
    public function download()
    {
        $data = $this->postData();
        $id = $data['id'] ?? 0;
        $source = $data['source'] ?? '';
        if ($id == 0 || empty($source)) {
            return $this->renderError('参数错误');
        }
        $qrcode = new TableService($id, $source);
        $qrcode->getDownload();
    }

    /**
     * @Apidoc\Title("解绑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/unbind")
     * @Apidoc\Param("table_id", type="int", require=true, default=0, desc="桌位id")
     * @Apidoc\Returned()
     */
    public function unbind($table_id)
    {
        /** @var TableModel $model */
        $model = TableModel::detail($table_id);
        if ($model->setUnbind()) {
            return $this->renderSuccess('', '解绑成功');
        }
        return $this->renderError($model->getError() ?: '解绑失败');
    }

    /**
     * @Apidoc\Title("桌台开关状态（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/switchStatus")
     * @Apidoc\Param("table_id", type="int", require=true, default=0, desc="桌位id")
     * @Apidoc\Param("switch_status", type="int", require=true, default=0, desc="桌台开关状态 0-关 1-开")
     * @Apidoc\Returned()
     */
    public function switchStatus()
    {
        $data = $this->postData();
        $table_id = $data['table_id'] ?? 0;
        $switch_status = $data['switch_status'] ?? 0;
        /** @var TableModel $model */
        $model = TableModel::detail($table_id);
        if ($model->setSwitchStatus($switch_status)) {
            return $this->renderSuccess('', '操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("桌位导入")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/store.table.table/imports")
     * @Apidoc\Param("mode", type="int", default="get", require=true, desc="请求模式：get=获取基础列表，save=保存")
     * @Apidoc\Param("list", type="array", require=true, desc="excel数据数组列表 - get请求时使用", children={
     *      @Apidoc\Param("table_no", type="string", require=true, default="001", desc="桌位名称"),
     *      @Apidoc\Param("type_name", type="string", require=true, default="Small table", desc="所属类型"),
     *      @Apidoc\Param("area_name", type="string", require=true, default="所有", desc="所属区域"),
     *      @Apidoc\Param("sort", type="string", require=true, default=0, desc="所属区域"),
     *      @Apidoc\Param("row", type="int", require=true, default="1", desc="excel表的行编号")
     * })
     * @Apidoc\Returned("list", type="array", require=true, desc="数据数组列表 - save请求时使用", children={
     *      @Apidoc\Param("table_no", type="string", require=true, default="", desc="桌位名称"),
     *      @Apidoc\Param("type_name", type="string", require=true, default=0, desc="所属类型"),
     *      @Apidoc\Param("type_id", type="int", require=true, default=0, desc="所属类型ID"),
     *      @Apidoc\Param("area_name", type="string", require=true, default=0, desc="所属区域"),
     *      @Apidoc\Param("area_id", type="int", require=true, default=0, desc="所属区域ID"),
     *      @Apidoc\Param("sort", type="int", require=true, default=0, desc="所属区域"),
     *      @Apidoc\Param("row", type="int", require=true, default="1", desc="excel表的行编号"),
     *      @Apidoc\Param("table_no_is_exist", type="bool", require=true, default="1", desc="桌号是否已经存在"),
     * })
     * @Apidoc\Returned("area_list", type="array", ref="app\shop\model\store\TableArea\getAllList", desc="下拉选择的区域列表")
     * @Apidoc\Returned("type_list", type="array", ref="app\shop\model\store\TableType\getAllList", desc="下拉选择的类型列表")
     */
    public function imports(TableImportsValidate $validate)
    {
        $mode = $this->request->param('mode', 'get');
        $list = $this->request->param('list');
        //
        if (($mode != 'get' && $mode != 'save') || !$list) {
            return $this->renderError('参数错误');
        }
        //
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];
        // 
        foreach ($list as $key => &$val) {
            try {
                // 验证参数
                $validate->goCheck($mode, $val);
                // 
                if ($mode == 'get') {
                    $val['type_id'] = TableTypeModel::where('type_name', $val['type_name'])->value('type_id') ?: 0;
                    $val['area_id'] = TableAreaModel::where('area_name', $val['area_name'])->value('area_id') ?: 0;
                }
                // 
                $val['table_no_is_exist'] = CheckService::checkNameExist('table', ['exist'=>$val['table_no']], $shop_supplier_id)['exist'];
            } catch (\Throwable $th) {
                return $this->renderError(__('行').'[' . ($val['row'] ?? 1) .']: ' . __($th->getMessage()));
            }
        }
        // 返回基础列表
        if ($mode == 'get') {
            $area_list = TableAreaModel::getAllList($shop_supplier_id);
            $type_list = TableTypeModel::getAllList($shop_supplier_id);
            return $this->renderSuccess('', compact('list', 'area_list', 'type_list'));
        }
        // 验证值是否存在重覆
        if ($duplicateRows = ArrayHelp::findDuplicateIds($list, 'table_no')) {
            return $this->renderError(__('行').'[' . (implode(',', $duplicateRows)) .']: ' . __('桌位名称不能重复'), compact('duplicateRows'));
        };
        // 验证数据id 是否存在
        foreach ($list as $key => &$val) {
            // 传过来的信息
            $val['shop_supplier_id'] = $shop_supplier_id;
            // 
            $tableType = TableTypeModel::where('type_id',  $val['type_id'])->find();
            if (!$tableType) {
                return $this->renderError(__('行').'[' . ($val['row'] ?? 1) .']: ' . __('所属类型不存在'));
            }
            $tableArea = TableAreaModel::where('area_id',  $val['area_id'])->find();
            if (!$tableArea) {
                return $this->renderError(__('行').'[' . ($val['row'] ?? 1) .']: ' . __('所属区域不存在'));
            }
        }
        // 
        $model = new TableModel;
        $licenses = request()->licenses;
        $tl = ($licenses['z_l'] ?? 0);
        $thatnum = $model->where('shop_supplier_id', '=',$shop_supplier_id)->count();
        if ($tl != -1 && ($thatnum + count($list)) > $tl) {
            return $this->renderError(__('桌台数量已达上限，最大可导入数量：') . $tl - $thatnum);
        }
        foreach ($list as $key => &$val) {
            if (!$model->add($val)) {
                return $this->renderError($model->getError() ?: '添加失败');
            }
        }
        // 
        return $this->renderSuccess('');
    }
}
