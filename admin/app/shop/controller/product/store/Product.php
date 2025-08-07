<?php

namespace app\shop\controller\product\store;

use help\ArrayHelp;
use app\shop\service\CheckService;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\shop\service\ProductService;
use app\common\model\tax\TaxCategory;
use app\shop\validate\ProductImportsValidate;
use app\common\model\product\Spec as SpecModel;
use app\common\model\product\Unit as UnitModel;
use app\shop\model\product\Product as ProductModel;
use app\common\model\product\Material as MaterialModel;
use app\common\model\product\ProductPackageRecommend;
use app\shop\model\product\Category as CategoryModel;
use app\shop\validate\ProductPackageRecommendValidate;

/**
 * 店内商品
 * @Apidoc\Group("product")
 * @Apidoc\Sort(4)
 */
class Product extends Controller
{
    /**
     * @Apidoc\Title("商品列表(全部)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/index")
     * @Apidoc\Param("product_name", type="string", require=false, desc="商品名称")
     * @Apidoc\Param("category_id", type="int", default=0, require=false, desc="分类id")
     * @Apidoc\Param("type", type="string", require=false, desc="是否上架 sell-上架 lower-下架")
     * @Apidoc\Param("stock", type="int", default=0, require=false, desc="库存 0-全部 10-低于10 20-低于20 ....")
     * @Apidoc\Param("product_ids", type="string", require=false, desc="商品ids，逗号分隔")
     * @Apidoc\Param("material_type", type="int", default=10, require=false, desc="v1.0.2 类型 10-成品 20-材料 30-套餐")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\Product\getList")
     */
    public function index()
    {
        // 获取全部商品列表
        $model = new ProductModel;
        $list = $model->getList(array_merge([
            'status' => -1,
            'product_type' => 1,
        ], $this->postData()));
        // 商品分类
        $category = CategoryModel::getCacheTree(1, 0, $this->store);
        // 打印标签
        $label = [];
        // 数量
        $product_count = [
            'lower' => $model->getCount('lower', $this->store['user']['shop_supplier_id'], 1),
        ];
        return $this->renderSuccess('', compact('list', 'category', 'label', 'product_count'));
    }

    /**
     * @Apidoc\Title("商品列表(在售)")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/lists")
     * @Apidoc\Param("product_name", type="string", require=false, desc="商品名称")
     * @Apidoc\Param("category_id", type="int", default=0, require=false, desc="分类id")
     * @Apidoc\Param("type", type="string", require=false, desc="是否上架 sell-上架 lower-下架")
     * @Apidoc\Param("stock", type="int", default=0, require=false, desc="库存 0-全部 10-低于10 20-低于20 ....")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", ref="app\shop\model\product\Product\getLists")
     */
    public function lists()
    {
        // 获取全部商品列表
        $model = new ProductModel;
        $list = $model->getLists($this->postData());
        // 商品分类
        $catgory = CategoryModel::getCacheTree(1, 0, $this->store);
        return $this->renderSuccess('', compact('list', 'catgory'));
    }

    /**
     * @Apidoc\Title("添加商品")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/add")
     * @Apidoc\Param("params", type="object", require=false, desc="商品数据", children={
     *   @Apidoc\Param("product_name", type="string", require=false, desc="商品名称"),
     *   @Apidoc\Param("category_id", type="int", default=0, require=false, desc="分类id"),
     *   @Apidoc\Param("image", type="array", require=true, desc="商品图片", children={
     *      @Apidoc\Param("file_id", type="int", require=true, desc="图片id"),
     *      @Apidoc\Param("file_path", type="string", require=true, desc="图片路径"),
     *   }),
     *   @Apidoc\Param("selling_point", type="string", require=true, desc="商品卖点"),
     *   @Apidoc\Param("spec_type", type="int", require=true, desc="产品规格(10单规格 20多规格)"),
     *   @Apidoc\Param("deduct_stock_type", type="int", require=true, desc="库存计算方式(10下单减库存 20付款减库存)"),
     *   @Apidoc\Param("num_type", type="int", require=true, desc="数量计算方法, 0-整数 1-小数"),
     *   @Apidoc\Param("is_alone_grade", type="int", require=true, desc="会员折扣设置(0默认等级折扣 1单独设置折扣)"),
     *   @Apidoc\Param("open_overall_discount", type="int", require=true, desc="是否开启整单折扣(0否 1是)"),
     *   @Apidoc\Param("sku", type="array", require=true, desc="商品sku", children={
     *      @Apidoc\Param("spec_sku_id", type="string", require=true, desc="规格id"),
     *      @Apidoc\Param("spec_name", type="string", require=true, desc="规格名称"),
     *      @Apidoc\Param("product_price", type="decimal", require=true, desc="价格"),
     *      @Apidoc\Param("stock_num", type="int", require=true, desc="库存"),
     *      @Apidoc\Param("product_weight", type="decimal", require=true, desc="产品重量(Kg)"),
     *      @Apidoc\Param("cost_price", type="decimal", require=true, desc="成本价"),
     *   }),
     *   @Apidoc\Param("product_attr", type="array", require=true, desc="商品属性", children={
     *      @Apidoc\Param("attribute_name", type="string", require=true, desc="属性名称"),
     *      @Apidoc\Param("attribute_value", type="array", require=true, desc="属性值"),
     *      @Apidoc\Param("much", type="int", require=true, desc="属性可选值数量"),
     *   }),
     *   @Apidoc\Param("product_feed", type="array", require=true, desc="商品加料", children={
     *      @Apidoc\Param("feed_name", type="string", require=true, desc="加料名称"),
     *      @Apidoc\Param("price", type="decimal", require=true, desc="加料价格"),
     *   }),
     *   @Apidoc\Param("min_buy", type="int", require=false, desc="最小购买量"),
     *   @Apidoc\Param("product_unit", type="string", default=0, require=false, desc="商品单位"),
     *   @Apidoc\Param("content", type="string", require=true, desc="产品详情"),
     *   @Apidoc\Param("product_status", type="int", require=true, desc="产品状态(10上架 20下架 )"),
     *   @Apidoc\Param("is_show_cashier", type="int", require=true, desc="是否显示在收银台(1显示 2不显示)"),
     *   @Apidoc\Param("is_show_tablet", type="int", require=true, desc="是否显示在平板电脑(1显示 2不显示)"),
     *   @Apidoc\Param("is_show_kitchen", type="int", require=true, desc="是否显示在厨显(1显示 2不显示)"),
     *   @Apidoc\Param("is_show_assistant", type="int", require=true, desc="是否显示在点餐助手(1显示 2不显示)（v1.0.5）"),
     *   @Apidoc\Param("is_show_h5", type="int", require=true, desc="是否显示在h5(1显示 2不显示)（v1.0.7）"),
     *   @Apidoc\Param("sales_initial", type="int", require=true, desc="初始销量"),
     *   @Apidoc\Param("product_sort", type="int", require=true, desc="产品排序(数字越小越靠前)"),
     *   @Apidoc\Param("limit_num", type="int", require=true, desc="限购数量0为不限"),
     *   @Apidoc\Param("special_id", type="int", require=true, desc="特殊分类id"),
     *   @Apidoc\Param("is_points_gift", type="int", require=true, desc="是否开启积分赠送(1开启 0关闭)"),
     *   @Apidoc\Param("is_enable_grade", type="int", require=true, desc="是否开启会员折扣(1开启 0关闭)"),
     *   @Apidoc\Param("alone_grade_type", type="int", require=true, desc="折扣金额类型(10百分比 20固定金额)"),
     *   @Apidoc\Param("label_id", type="int", require=true, desc="打印标签id"),
     *   @Apidoc\Param("alone_grade_equity", type="string", require=true, desc="单独设置折扣的配置"),
     *   @Apidoc\Param("stock", type="int", default=0, require=false, desc="库存 0-全部 10-低于10 20-低于20 ...."),
     *   @Apidoc\Param("type", type="int", default=10, require=false, desc="v1.0.2 类型 10-成品 20-材料 30-套餐"),
     *   @Apidoc\Param("erp_supplier_id", type="int", default=0, require=false, desc="v1.0.2 erp供应商id"),
     *   @Apidoc\Param("productTaxes", type="array", require=true, desc="产品关联税类", children={
     *      @Apidoc\Param("product_tax_type", type="int", require=true, desc="产品关联税类类型，1-堂食税类，2-外带税类"),
     *      @Apidoc\Param("tax_category_id", type="int", require=true, desc="税类id"),
     *   }),
     *   @Apidoc\Param("package_price", type="decimal", require=false, desc="套餐价格"),
     *   @Apidoc\Param("is_open_stock", type="int", require=false, desc="是否开启库存"),
     *   @Apidoc\Param("package_stock", type="decimal", require=false, desc="套餐可售卖库存"),
     *   @Apidoc\Param("package_group", type="array", require=false, desc="套餐分组", children={
     *      @Apidoc\Param("group_name", type="string", require=true, desc="套餐分组名称"),
     *      @Apidoc\Param("product_list", type="array", require=true, desc="套餐分组商品", children => {
     *          @Apidoc\Param("product_id", type="int", require=true, desc="商品id: product_bom_uuid"),
     *          @Apidoc\Param("sort", type="int", require=true, desc="商品排序"),
     *          @Apidoc\Param("num", type="int", require=true, desc="商品数量"),
     *      }),
     *   }),
     * })
     */
    public function add($scene = 'add')
    {
        if ($this->request->isGet()) {
            return $this->getBaseData();
        }
        $data = json_decode($this->postData()['params'], true);
        $model = in_array(($data['type'] ?? 10), [10, 30]) ? new ProductModel : new MaterialModel;
        $data['shop_user_id'] = $this->store['user']['shop_user_id'];
        if ($model->add($data)) {
            return $this->renderSuccess('添加成功');
        }
        return $this->renderError($model->getError() ?: '添加失败', $model->getErrorData() ?: []);
    }

    /**
     * 获取基础数据
     */
    public function getBaseData()
    {
        return $this->renderSuccess('', array_merge(ProductService::getEditData(1, $this->store), []));
    }

    /**
     * @Apidoc\Title("编辑商品")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/edit")
     * @Apidoc\Param("type ", type="int", default=10, require=false, desc="v1.0.2 类型 10-成品 20-材料 30-套餐"),
     * @Apidoc\Param("product_id ", type="int", default="", require=false, desc="商品id"),
     * @Apidoc\Param("params", type="object", require=false, desc="商品数据", children={
     *   @Apidoc\Param("erp_supplier_id", type="int", default=0, require=false, desc="v1.0.2 erp供应商id"),
     *   @Apidoc\Param("product_name", type="string", require=false, desc="商品名称"),
     *   @Apidoc\Param("product_id", type="int", require=true, desc="商品id"),
     *   @Apidoc\Param("category_id", type="int", default=0, require=false, desc="分类id"),
     *   @Apidoc\Param("image", type="array", require=true, desc="商品图片", children={
     *      @Apidoc\Param("file_id", type="int", require=true, desc="图片id"),
     *      @Apidoc\Param("file_path", type="string", require=true, desc="图片路径"),
     *   }),
     *   @Apidoc\Param("selling_point", type="string", require=true, desc="商品卖点"),
     *   @Apidoc\Param("spec_type", type="int", require=true, desc="产品规格(10单规格 20多规格)"),
     *   @Apidoc\Param("deduct_stock_type", type="int", require=true, desc="库存计算方式(10下单减库存 20付款减库存)"),
     *   @Apidoc\Param("is_alone_grade", type="int", require=true, desc="会员折扣设置(0默认等级折扣 1单独设置折扣)"),
     *   @Apidoc\Param("open_overall_discount", type="int", require=true, desc="是否开启整单折扣(0否 1是)"),
     *   @Apidoc\Param("sku", type="array", require=true, desc="商品sku", children={
     *      @Apidoc\Param("spec_sku_id", type="string", require=true, desc="规格id"),
     *      @Apidoc\Param("spec_name", type="string", require=true, desc="规格名称"),
     *      @Apidoc\Param("product_price", type="decimal", require=true, desc="价格"),
     *      @Apidoc\Param("stock_num", type="int", require=true, desc="库存"),
     *      @Apidoc\Param("product_weight", type="decimal", require=true, desc="产品重量(Kg)"),
     *      @Apidoc\Param("cost_price", type="decimal", require=true, desc="成本价"),
     *   }),
     *   @Apidoc\Param("product_attr", type="array", require=true, desc="商品属性", children={
     *      @Apidoc\Param("attribute_name", type="string", require=true, desc="属性名称"),
     *      @Apidoc\Param("attribute_value", type="array", require=true, desc="属性值"),
     *      @Apidoc\Param("much", type="int", require=true, desc="属性可选值数量"),
     *   }),
     *   @Apidoc\Param("product_feed", type="array", require=true, desc="商品加料", children={
     *      @Apidoc\Param("feed_name", type="string", require=true, desc="加料名称"),
     *      @Apidoc\Param("price", type="decimal", require=true, desc="加料价格"),
     *   }),
     *   @Apidoc\Param("min_buy", type="int", require=false, desc="最小购买量"),
     *   @Apidoc\Param("product_unit", type="string", default=0, require=false, desc="商品单位"),
     *   @Apidoc\Param("content", type="string", require=true, desc="产品详情"),
     *   @Apidoc\Param("product_status", type="int", require=true, desc="产品状态(10上架 20下架 )"),
     *   @Apidoc\Param("is_show_cashier", type="int", require=true, desc="是否显示在收银台(1显示 2不显示)"),
     *   @Apidoc\Param("is_show_tablet", type="int", require=true, desc="是否显示在平板电脑(1显示 2不显示)"),
     *   @Apidoc\Param("is_show_kitchen", type="int", require=true, desc="是否显示在厨显(1显示 2不显示)"),
     *   @Apidoc\Param("is_show_assistant", type="int", require=true, desc="是否显示在点餐助手(1显示 2不显示)（v1.0.5）"),
     *   @Apidoc\Param("is_show_h5", type="int", require=true, desc="是否显示在h5(1显示 2不显示)（v1.0.7）"),
     *   @Apidoc\Param("is_show_delivery", type="int", require=true, desc="是否显示在外送(1显示 2不显示)（v2.3.0）"),
     *   @Apidoc\Param("sales_initial", type="int", require=true, desc="初始销量"),
     *   @Apidoc\Param("product_sort", type="int", require=true, desc="产品排序(数字越小越靠前)"),
     *   @Apidoc\Param("limit_num", type="int", require=true, desc="限购数量0为不限"),
     *   @Apidoc\Param("special_id", type="int", require=true, desc="特殊分类id"),
     *   @Apidoc\Param("is_points_gift", type="int", require=true, desc="是否开启积分赠送(1开启 0关闭)"),
     *   @Apidoc\Param("is_ind_agent", type="int", require=true, desc="是否开启单独分销(0关闭 1开启)"),
     *   @Apidoc\Param("is_enable_grade", type="int", require=true, desc="是否开启会员折扣(1开启 0关闭)"),
     *   @Apidoc\Param("alone_grade_type", type="int", require=true, desc="折扣金额类型(10百分比 20固定金额)"),
     *   @Apidoc\Param("label_id", type="int", require=true, desc="打印标签id"),
     *   @Apidoc\Param("alone_grade_equity", type="string", require=true, desc="单独设置折扣的配置"),
     *   @Apidoc\Param("productTaxes", type="array", require=true, desc="产品关联税类", children={
     *      @Apidoc\Param("product_tax_type", type="int", require=true, desc="产品关联税类类型，1-堂食税类，2-外带税类"),
     *      @Apidoc\Param("tax_category_id", type="int", require=true, desc="税类id"),
     *  }),
     *  @Apidoc\Param("package_price", type="decimal", require=false, desc="套餐价格"),
     *  @Apidoc\Param("is_open_stock", type="int", require=false, desc="是否开启库存"),
     *  @Apidoc\Param("package_stock", type="decimal", require=false, desc="套餐可售卖库存"),
     *  @Apidoc\Param("package_group", type="array", require=false, desc="套餐分组", children={
     *     @Apidoc\Param("group_id", type="int", require=true, desc="套餐分组id"),
     *     @Apidoc\Param("group_name", type="string", require=true, desc="套餐分组名称"),
     *     @Apidoc\Param("product_list", type="array", require=true, desc="套餐分组商品", children => {
     *         @Apidoc\Param("item_id", type="int", require=true, desc="套餐分组商品id"),
     *         @Apidoc\Param("product_id", type="int", require=true, desc="商品id: product_bom_uuid"),
     *         @Apidoc\Param("sort", type="int", require=true, desc="商品排序"),
     *         @Apidoc\Param("num", type="int", require=true, desc="商品数量"),
     *     }),
     *   }),
     * })
     */
    public function edit($product_id, $scene = 'edit')
    {
        if ($this->request->isGet()) {
            $model = ProductModel::detail($product_id);
            if (!$model) {
                $model = MaterialModel::detail($product_id);
            }
            return $this->renderSuccess('', array_merge(ProductService::getEditData(1, $this->store), compact('model')));
        }
        $data = array_merge(json_decode($this->postData()['params'], true), ['shop_user_id' => $this->store['user']['shop_user_id']]);
        /** @var ProductModel $model */
        $model = in_array(($data['type'] ?? 10), [10, 30]) ? ProductModel::detail($product_id) : MaterialModel::detail($product_id);
        if ($model->edit($data)) {
            return $this->renderSuccess('更新成功');
        }
        return $this->renderError($model->getError() ?: '更新失败', $model->getErrorData() ?: []);
    }

    /**
     * @Apidoc\Title("修改商品状态")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/state")
     * @Apidoc\Param("product_id", type="int", require=true, desc="商品id")
     * @Apidoc\Param("state", type="int", require=true, desc="状态 0-下架 1-上架")
     * @Apidoc\Returned()
     */
    public function state($product_id, $state)
    {
        /** @var ProductModel $model */
        $model = ProductModel::detail($product_id);
        if ($model) {
            if (!$model->setStatus($state)) {
                return $this->renderError($model->getError() ?: '操作失败');
            }
            return $this->renderSuccess('操作成功');
        }
        /** @var MaterialModel $model */
        $material = MaterialModel::detail($product_id);
        if ($material) {
            if ($state != 10 && $material->relatedMaterial()->count() > 0) {
                return $this->renderError('该材料已被使用，无法操作');
            }
            $res = $material->save(['status' => $state == 10 ? 1 : 0]);
            if ($res === false) {
                return $this->renderError('操作失败');
            }
            return $this->renderSuccess('操作成功');
        }
        // 
        return $this->renderError('操作失败');
    }

    /**
     * @Apidoc\Title("删除商品")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/delete")
     * @Apidoc\Param("product_id", type="int", require=true, desc="商品id，多个逗号隔开")
     * @Apidoc\Returned()
     */
    public function delete($product_id)
    {
        $model = new ProductModel;
        if (!$model->setDelete($product_id, $this->store['user']['shop_user_id'])) {
            return $this->renderError($model->getError() ?: '删除失败');
        }
        return $this->renderSuccess('删除成功');
    }

    /**
     * @Apidoc\Title("同步商品到门店")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/transmit")
     * @Apidoc\Param("product_id", type="int", require=true, desc="商品id")
     * @Apidoc\Param("sku", type="array", require=true, desc="商品sku")
     * @Apidoc\Returned()
     */
    public function transmit()
    {
        $model = new ProductModel;
        if (!$model->transmit($this->postData())) {
            return $this->renderError($model->getError() ?: '操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("商品导入")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/imports")
     * @Apidoc\Param("mode", type="int", default="get", require=true, desc="请求模式：get=获取基础列表，save=保存")
     * @Apidoc\Param("list", type="array", require=true, desc="excel数据数组列表 - get请求时使用", children={
     *      @Apidoc\Param("product_name", type="string",  default="get", default="简体中文:饮料", require=true, desc="商品名称"),
     *      @Apidoc\Param("category_name", type="string", require=true, default="分类/主分类", desc="所属分类"),
     *      @Apidoc\Param("deduct_stock_type", type="string", require=true, default="1", desc="库存计算方式"),
     *      @Apidoc\Param("num_type", type="string", require=true, default="1", desc="数量计算方法, 1-整数 2-小数"),
     *      @Apidoc\Param("product_unit", type="string", require=true, default="个", desc="商品单位"),
     *      @Apidoc\Param("spec_name", type="string", require=true, default="规格", desc="规格名称"),
     *      @Apidoc\Param("img_name", type="string", require=true, default="图片名称", desc="图片名称"),
     *      @Apidoc\Param("product_stock", type="string", require=true, default="9999", desc="库存数量"),
     *      @Apidoc\Param("barcode", type="string", require=true, default="7618263962", desc="商品条码"),
     *      @Apidoc\Param("product_price", type="string", require=true, default="88.6", desc="商品价格"),
     *      @Apidoc\Param("product_status", type="string", require=true, default="1",  desc="商品状态"),
     *      @Apidoc\Param("product_ratin_tax_type", type="string", default="税类1", require=true, desc="堂食税类"),
     *      @Apidoc\Param("product_takeout_tax_type", type="string", default="税类2", require=true, desc="外带税类"),
     *      @Apidoc\Param("shows", type="int", require=true, default="12345", desc="显示：123456"),
     *      @Apidoc\Param("product_sort", type="string", require=true, desc="商品排序"),
     *      @Apidoc\Param("limit_num", type="int", require=true, default="0", desc="限购数量"),
     *      @Apidoc\Param("is_enable_grade", type="int", require=true, default="1", desc="是否开启会员折扣(1开启 0关闭)"),
     *      @Apidoc\Param("open_overall_discount", type="int", require=true, default="1", desc="整单折扣(1开启 0关闭)"),
     *      @Apidoc\Param("row", type="int", require=true, default="1", desc="excel表的行编号"),
     * }),
     * @Apidoc\Returned("category_list", type="array", ref="app\common\model\product\Category", desc="下拉分类列表")
     * @Apidoc\Returned("unit_list", type="array", ref="app\common\model\product\Unit", desc="下拉单位列表")
     * @Apidoc\Returned("spec_list", type="array", ref="app\common\model\product\Spec", desc="下拉规格列表")
     * @Apidoc\Returned("tax_list", type="array", ref="app\common\model\tax\TaxCategory", desc="下拉税类列表")
     * @Apidoc\Returned("list", type="array", require=true, desc="数据数组列表 - save请求时使用", children={
     *      @Apidoc\Param("product_name", type="string",  default="get", default="{'zh':'饮料'}", require=true, desc="商品名称"),
     *      @Apidoc\Param("category_name", type="string", require=true, default="分类/主分类", desc="所属分类"),
     *      @Apidoc\Param("deduct_stock_type", type="string", require=true, default="10", desc="库存计算方式(10下单减库存 20付款减库存)"),
     *      @Apidoc\Param("num_type", type="string", require=true, default="1", desc="数量计算方法, 1-整数 2-小数"),
     *      @Apidoc\Param("product_unit", type="string", require=true, default="个", desc="商品单位"),
     *      @Apidoc\Param("spec_name", type="string", require=true, default="规格", desc="规格名称"),
     *      @Apidoc\Param("img_name", type="string", require=true, default="图片名称", desc="图片名称"),
     *      @Apidoc\Param("product_stock", type="string", require=true, default="9999", desc="库存数量"),
     *      @Apidoc\Param("barcode", type="string", require=true, default="7618263962", desc="商品条码"),
     *      @Apidoc\Param("product_price", type="string", require=true, default="88.6", desc="商品价格"),
     *      @Apidoc\Param("product_status", type="string", require=true, default="1",  desc="商品状态 状态(10上架 20下架)"),
     *      @Apidoc\Param("product_ratin_tax_type", type="string", require=true, desc="堂食税类"),
     *      @Apidoc\Param("product_takeout_tax_type", type="string", require=true, desc="外带税类"),
     *      @Apidoc\Param("shows", type="int", require=true, default="12345", desc="显示：123456"),
     *      @Apidoc\Param("product_sort", type="string", require=true, desc="商品排序"),
     *      @Apidoc\Param("limit_num", type="int", require=true, default="0", desc="限购数量"),
     *      @Apidoc\Param("is_enable_grade", type="int", require=true, default="1", desc="是否开启会员折扣(1开启 0关闭)"),
     *      @Apidoc\Param("open_overall_discount", type="int", require=true, default="1", desc="整单折扣(1开启 0关闭)"),
     *      @Apidoc\Param("row", type="int", require=true, default="1", desc="excel表的行编号"),
     *      @Apidoc\Param("is_show_cashier", type="int", desc="是否显示在收银端 1-显示 2-不显示'"),
     *      @Apidoc\Param("is_show_tablet", type="int", desc="是否显示在平板端 1-显示 2-不显示"),
     *      @Apidoc\Param("is_show_kitchen", type="int", desc="是否显示在送厨端 1-显示 2-不显示"),
     *      @Apidoc\Param("is_show_assistant", type="int", desc="是否显示在点餐助手 1-显示 2-不显示"),
     *      @Apidoc\Param("is_show_h5", type="int", desc="是否显示在h5 1-显示 2-不显示"),
     *      @Apidoc\Param("is_show_delivery", type="int", desc="是否显示在外送 1-显示 2-不显示"),
     *      @Apidoc\Param("category_id", type="int", desc="所选的分类Id"),
     *      @Apidoc\Param("unit_id", type="int", desc="所选的单位Id"),
     *      @Apidoc\Param("spec_id", type="int", desc="所选的规格Id"),
     *      @Apidoc\Param("ratin_tax_id", type="int", desc="所选的堂食税类Id"),
     *      @Apidoc\Param("takeout_tax_id", type="int", desc="所选的外带税类Id"),
     *      @Apidoc\Param("product_name_is_exist", type="array", desc="商品名称是否存在"),
     *      @Apidoc\Param("img_name_is_exist", type="bool", desc="图片名称是否存在"),
     *      @Apidoc\Param("barcode_is_exist", type="bool", desc="条形码是否存在"),
     * }),
     */
    public function imports(ProductImportsValidate $validate)
    {
        $mode = $this->request->param('mode', 'get');
        $list = $this->request->param('list');
        //
        if (($mode != 'get' && $mode != 'save') || !$list) {
            return $this->renderError('参数错误');
        }
        //
        $shop_supplier_id = $this->store['user']['shop_supplier_id'];

        // 语言key
        $langKeys = [
            'English' => 'en',
            'ภาษาไทย' => 'th',
            '简体中文' => 'zh',
            '繁體中文' => 'zhtw',
            'Türkçe' => 'tr',
            'မြန်မာဘာသာ' => 'my',
            '日本語' => 'ja',
            '한국어' => 'ko',
            'Svenska' => 'sv',
        ];

        // 验证 格式化参数
        foreach ($list as $key => &$val) {
            try {
                // 验证参数
                $validate->goCheck($mode, $val);
                // 处理数值转换
                $val['deduct_stock_type'] = strstr($val['deduct_stock_type'], '1') ? 10 : 20;
                $val['product_status'] = strstr($val['product_status'], '1') ? 10 : 20;
                $val['product_sort'] = $val['product_sort'] ?: 0;
                // 处理商品名称
                if ($mode == 'get') {
                    // 验证
                    $val['category_id'] = $validate->checkCategoryNameExist($val['category_name'], true);
                    $val['unit_id'] = $validate->checkProductUnitExist($val['product_unit'], true);
                    $val['spec_id'] = $validate->checkSpecNameExist($val['spec_name'], true);
                    $val['ratin_tax_id'] = $validate->checkTaxExist($val['product_ratin_tax_type'], true);
                    $val['takeout_tax_id'] = $validate->checkTaxExist($val['product_takeout_tax_type'], true);
                    // 处理显示
                    $val['is_show_cashier'] = strstr($val['shows'], '1') ? 1 : 2;
                    $val['is_show_tablet'] = strstr($val['shows'], '2') ? 1 : 2;
                    $val['is_show_kitchen'] = strstr($val['shows'], '3') ? 1 : 2;
                    $val['is_show_assistant'] = strstr($val['shows'], '4') ? 1 : 2;
                    $val['is_show_h5'] = strstr($val['shows'], '5') ? 1 : 2;
                    $val['is_show_delivery'] = strstr($val['shows'], '6') ? 1 : 2;
                    // 按小数计价，不在助手、平板、扫码端显示
                    if ($val['num_type'] == 2 && ($val['is_show_tablet'] != 2 ||  $val['is_show_assistant'] != 2 ||  $val['is_show_h5'] != 2 ||  $val['is_show_delivery'] != 2)) {
                        return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('按小数计价只能显示到收银机和厨显'), $val);
                    }
                    // 
                    if ($this->store['supplier']['delivery_status'] != 1 && $val['is_show_delivery'] != 2) {
                        return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('未配置外送渠道，无法选择在外送显示'), $val);
                    }
                    //
                    $productName = [];
                    foreach (explode("\n",  $val['product_name']) as $name) {
                        $name = explode(":", $name);
                        if (count($name) < 2) {
                            return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('商品名称格式错误'), $val);
                        }
                        if (!isset($langKeys[$name[0]])) {
                            return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('商品名称对应语言不存在') . "[{$name[0]}]", $val);
                        }
                        $productName[$langKeys[$name[0]]] = $name[1];
                    }
                    $val['product_name'] = json_encode($productName, JSON_UNESCAPED_UNICODE);
                } else {
                    $val['num_type'] = strstr($val['num_type'], '1') ? 0 : 1; // excel内容中1表示整数计价，2表示小数计价；对应数据库值0表示整数，1表示小数
                    $productName = json_decode($val['product_name'], true);
                    if (!$productName) {
                        return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('商品名称格式错误'), $val);
                    }
                }
                // 验证是否已经存在
                $val['product_name_is_exist'] = CheckService::checkNameExist('product', $productName, $shop_supplier_id);
                $val['img_name_is_exist'] = $val['img_name'] ? CheckService::checkNameExist('product_img', ['exist' => $val['img_name']], $shop_supplier_id)['exist'] : false;
                $val['barcode_is_exist'] = CheckService::checkNameExist('product_bom_barcode', ['exist' => $val['barcode']], $shop_supplier_id)['exist'];
            } catch (\Throwable $th) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __($th->getMessage()));
            }
        }

        // 返回基础列表
        if ($mode == 'get') {
            $category_list = CategoryModel::getCacheTree(1, 0, $this->store);
            $unit_list = (new UnitModel)->getLists($shop_supplier_id);
            $spec_list = (new SpecModel)->getLists($shop_supplier_id);
            $tax_list = (new TaxCategory)->getList();
            foreach ($tax_list as $key => $tax_item) {
                $tax_list[$key]['id'] = $tax_item['tax_category_id'];
            }
            return $this->renderSuccess('', compact('list', 'category_list', 'unit_list', 'spec_list', 'tax_list'));
        }

        // 商品条码不能重复
        if ($duplicateRows = ArrayHelp::findDuplicateIds($list, 'barcode', 'row', true)) {
            return $this->renderError(__('行') . '[' . (implode(',', $duplicateRows)) . ']: ' . __('商品条码不能重复'), compact('duplicateRows'));
        };
        // die;
        foreach ($list as $key => &$val) {
            if (in_array(true, $val['product_name_is_exist'])) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('商品名称已存在'), $val);
            }
            if ($val['img_name_is_exist'] == true) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('图片名称已存在'), $val);
            }
            if ($val['barcode_is_exist'] == true) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('商品条码已存在'), $val);
            }
        }

        // 多规格合并
        $lists = [];
        $productNames = [];
        foreach ($list as $key => &$val) {
            $md5Key = md5($val['product_name']);
            //
            $spec = SpecModel::where('uuid',  $val['spec_id'])->find();
            if (!$spec) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('规格不存在'));
            }
            //
            $sku = [];
            if (isset($lists[$md5Key])) {
                $sku = $lists[$md5Key]['sku'];
            }
            $sku[$val['spec_id']] = [
                "spec_id" => $val['spec_id'],
                "spec_sku_id" => $val['spec_id'],
                "product_price" => $val['product_price'],
                "spec_name" => $spec->spec_name,
                "stock_num" => $val['product_stock'],
                "product_weight" => 0,
                "cost_price" => $val['product_price'],
                "barcode" => $val['barcode'] ?? '',
                "row" => $val['row'],
            ];
            //
            if (!isset($lists[$md5Key])) {
                $lists[$md5Key] = $val;
                $productNames[$val['row']] = json_decode($val['product_name'], true);
            }
            //
            $lists[$md5Key]['sku'] = $sku;
        }

        // 验证值是否存在重覆
        foreach ($langKeys as $langKey) {
            if ($duplicateRows = ArrayHelp::findDuplicateIds($productNames, $langKey)) {
                return $this->renderError(__('行') . '[' . (implode(',', $duplicateRows)) . ']: ' . __('商品名称不能重复'), compact('duplicateRows'));
            };
        }
        if ($duplicateRows = ArrayHelp::findDuplicateIds($lists, 'img_name')) {
            return $this->renderError(__('行') . '[' . (implode(',', $duplicateRows)) . ']: ' . __('图片名称不能重复'), compact('duplicateRows'));
        };

        // 验证数据id 是否存在
        foreach ($lists as $key => &$val) {
            $val['shop_user_id'] = $this->store['user']['shop_user_id'];
            $val['shop_supplier_id'] = $shop_supplier_id;
            $val['type'] = 10;
            $val['spec_type'] = 20;
            $val['product_type'] = 0; 
            $val['image'] = [];
            // 分类
            $category = CategoryModel::where('uuid',  $val['category_id'])->find();
            if (!$category) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('分类不存在'));
            }
            // 单位
            $unit = UnitModel::where('uuid',  $val['unit_id'])->find();
            if (!$unit) {
                return $this->renderError(__('行') . '[' . ($val['row'] ?? 1) . ']: ' . __('单位不存在'));
            }
            $val['product_unit'] = $unit->unit_name;
            // 税类
            $val['productTaxes'] = [];
            if ($val['ratin_tax_id']) {
                $val['productTaxes'][] = [
                    'product_tax_type' => 1,
                    'tax_category_id' => $val['ratin_tax_id'],
                ];
            }
            if ($val['takeout_tax_id']) {
                $val['productTaxes'][] = [
                    'product_tax_type' => 2,
                    'tax_category_id' => $val['takeout_tax_id'],
                ];
            }
            // 排序规格
            $val['sku'] = array_values($val['sku']);
        }

        // 保存数据
        foreach ($lists as $key => &$val) {
            $model = new ProductModel;
            if (!$model->add($val)) {
                return $this->renderError($model->getError() ?: '添加失败');
            }
        }
        //
        return $this->renderSuccess('');
    }

    /**
     * @Apidoc\Title("批量修改分类（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/batchUpdateCategory")
     * @Apidoc\Param("product_ids", type="string", require=true, desc="商品id集，逗号分隔")
     * @Apidoc\Param("category_id", type="int", require=true, desc="分类id")
     * @Apidoc\Returned()
     */
    public function batchUpdateCategory()
    {
        $model = new ProductModel;
        if (!$model->batchUpdateCategory($this->postData())) {
            return $this->renderError($model->getError() ?: '操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("批量修改税类（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/batchUpdateTax")
     * @Apidoc\Param("product_ids", type="string", require=true, desc="商品id集，逗号分隔")
     * @Apidoc\Param("productTaxes", type="array", require=true, desc="产品关联税类", children={
     *      @Apidoc\Param("product_tax_type", type="int", require=true, desc="产品关联税类类型，1-堂食税类，2-外带税类"),
     *      @Apidoc\Param("tax_category_id", type="int", require=true, desc="税类id"),
     *  }),
     * @Apidoc\Returned()
     */
    public function batchUpdateTax()
    {
        $model = new ProductModel;
        if (!$model->batchUpdateTax($this->postData())) {
            return $this->renderError($model->getError() ?: '操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("批量修改整单折扣")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/batchUpdateOpenOverallDiscount")
     * @Apidoc\Param("product_ids", type="array", require=true, desc="商品uuid集")
     * @Apidoc\Returned()
     */
    public function batchUpdateOpenOverallDiscount()
    {
        $model = new ProductModel;
        if (!$model->batchUpdateOpenOverallDiscount($this->postData())) {
            return $this->renderError($model->getError() ?: '操作失败');
        }
        return $this->renderSuccess('操作成功');
    }

    /**
     * @Apidoc\Title("商品推荐")
     * @Apidoc\Desc("get请求是获取；post请求是提交修改")
     * @Apidoc\Method ("GET,POST")
     * @Apidoc\Url ("/index.php/shop/product.store.product/recommend")
     * @Apidoc\Returned("status", type="int", desc="状态")
     * @Apidoc\Returned("title", type="string", desc="推荐标题")
     * @Apidoc\Returned("product_packages", type="array", desc="商品列表", children={
     *      @Apidoc\Param("uuid", type="string", desc="商品UUID"),
     *      @Apidoc\Param("sort", type="int", desc="排序"),
     *      @Apidoc\Param("name", type="string", desc="商品名称"),
     * }),
     * @Apidoc\Returned()
     */
    public function recommend(ProductPackageRecommendValidate $validate)
    {
        if (request()->licenses['is_open_delivery'] == 0) {
            return $this->renderError('当前没有权限使用此功能');
        }
        $model = new ProductPackageRecommend();
        $data = $model->where('delete_time', 0)->hidden(['id', 'uuid', 'create_time', 'update_time', 'delete_time'])->find();
        if ($this->request->isGet()) {
            if (!is_null($data)) {
                $json = json_decode($data['product_packages'], true);
                $productPackageUuids = array_column($json, 'uuid');
                // 获取商品和语言
                $productPackages = ProductModel::with('MultiLanguageName')->where('uuid', 'in', $productPackageUuids)->select();
                $map = [];
                foreach ($productPackages as $productPackage) {
                    $map[$productPackage['uuid']] = $productPackage;
                }
                foreach ($json as $k => $item) {
                    $defaultNames = ["en_name" => "", "zh_name" => "", "zh_tw_name" => "", "th_name" => "", "my_name" => "", "ja_name" => "", "ko_name" => "", "tr_name" => ""];
                    if (!isset($map[$item['uuid']])) {
                        unset($json[$k]);
                        continue;
                    }
                    $productPackage = $map[$item['uuid']];
                    if (!is_null($productPackage) && !is_null($productPackage->MultiLanguageName)) {
                        $defaultNames = $productPackage->MultiLanguageName->hidden(['id', 'uuid', 'create_time', 'update_time', 'delete_time'])->toArray();
                    }
                    $lang = checkDetect();
                    $json[$k]['name'] = $defaultNames[($lang == 'zhtw' ? 'zh_tw' : $lang) . '_name'];
                }
                $data['product_packages'] = $json;
            } else {
                $data['status'] = 0;
                $data['title'] = '';
                $data['product_packages'] = [];
            }
            return $this->renderSuccess('操作成功', $data);
        }

        $param = $validate->goCheck('save');

        if (is_null($data)) {
            $data = $model;
        }

        $deliveryProductPackages = ProductModel::where('uuid', 'in', array_column($param['product_packages'], 'uuid'))->where('is_show_delivery', 1)->where('delete_time', 0)->select()->toArray();
        // 判断$param['product_packages'] 是否在$count中
        foreach ($param['product_packages'] as $productPackage) {
            if (!in_array($productPackage['uuid'], array_column($deliveryProductPackages, 'uuid'))) {
                return $this->renderError('【' . $productPackage['name'] . '】' . __('为非外送显示的商品，请移除'));
            }
        }

        $data->status = intval($param['status']);
        $data->title = $param['title'];
        $data->product_packages = json_encode($param['product_packages']);

        if (!$data->save()) {
            return $this->renderError($model->getError() ?: '操作失败');
        }
        return $this->renderSuccess('操作成功');
    }
}
