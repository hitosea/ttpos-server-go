<?php

namespace app\shop\controller;

use help\DiskHelp;
use think\facade\Db;
use help\LicenseHelp;
use app\shop\service\ShopService;
use app\shop\service\CheckService;
use hg\apidoc\annotation as Apidoc;
use app\common\model\product\Product;
use app\common\model\shop\BindRecord;
use app\common\model\product\Material;
use app\common\service\sync\SyncService;
use app\common\enum\settings\SettingEnum;
use app\common\library\language\engine\OpenAi;
use app\shop\model\product\Label as LabelModel;
use app\shop\model\product\Category as CategoryModel;
use app\common\model\settings\Setting as SettingModel;

/**
 * 后台首页
 * @Apidoc\Group("home")
 * @Apidoc\Sort(1)
 */
class Index extends Controller
{
    /**
     * @Apidoc\Title("后台首页")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/index")
     * @Apidoc\Param("username", type="string", require=true, default="", desc="用户名")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("top_data", type="array", desc="数据总览", children={
     *    @Apidoc\Returned("total_money", type="float", desc="营业总额"),
     *    @Apidoc\Returned("total_discount_money", type="float", desc="折扣总额"),
     *    @Apidoc\Returned("discount_money", type="float", desc="优惠折扣（v1.0.7）"),
     *    @Apidoc\Returned("user_discount_money", type="float", desc="会员折扣（v1.0.7）"),
     *    @Apidoc\Returned("user_total", type="int", desc="会员数"),
     *    @Apidoc\Returned("order_total", type="int", desc="订单数"),
     *    @Apidoc\Returned("refund_money", type="float", desc="退款金额"),
     *    @Apidoc\Returned("income_money", type="float", desc="预计收入"),
     *    @Apidoc\Returned("supplier_total", type="float", desc="店铺总数"),
     *    @Apidoc\Returned("product_total", type="float", desc="商品总数"),
     * })
     * @Apidoc\Returned("today_data", type="array", desc="今日概况", children={
     *    @Apidoc\Returned("order_total_price", type="array", desc="营业总额", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("order_discount_money", type="array", desc="折扣总额", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("discount_money", type="array", desc="优惠折扣（v1.0.7）", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("user_discount_money", type="array", desc="会员折扣（v1.0.7）", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("order_refund_money", type="array", desc="退款金额", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("income_money", type="array", desc="预计收入", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("new_supplier_total", type="array", desc="新供应商数", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("order_user_total", type="array", desc="下单用户数", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("order_total", type="array", desc="订单数", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     *    @Apidoc\Returned("new_user_total", type="array", desc="会员数", children={
     *        @Apidoc\Returned("tday", type="string", desc="今日"),
     *        @Apidoc\Returned("ytd", type="string", desc="昨日"),
     *    }),
     * })
     * @Apidoc\Returned("wait_data", type="array", desc="待办事项", children={
     *    @Apidoc\Returned("order", type="array", desc="订单", children={
     *        @Apidoc\Returned("disposal", type="string", desc="待处理订单数量"),
     *    }),
     *    @Apidoc\Returned("stock", type="array", desc="库存", children={
     *        @Apidoc\Returned("product", type="string", desc="库存预警数量"),
     *    }),
     *    @Apidoc\Returned("purchase", type="array", desc="采购单", children={
     *        @Apidoc\Returned("apply", type="string", desc="待审核数量"),
     *    }),
     *    @Apidoc\Returned("supplier", type="array", desc="商家", children={
     *        @Apidoc\Returned("cash_apply", type="int", desc="待审核"),
     *        @Apidoc\Returned("cash_money", type="int", desc="审核通过）"),
     *    }),
     * })
     * @Apidoc\Returned("product_data", type="array", desc="列表数据", children={
     *    @Apidoc\Returned("salesNumRank", type="array", desc="销量排行", children={
     *        @Apidoc\Returned("product_name", type="string", desc="商品名称（多语言）"),
     *        @Apidoc\Returned("product_name_text", type="string", desc="商品名称"),
     *        @Apidoc\Returned("total_num", type="int", desc="销量"),
     *        @Apidoc\Returned("total_price", type="float", desc="销售额"),
     *    }),
     *    @Apidoc\Returned("salesMoneyRank", type="array", desc="销售额排行", children={
     *        @Apidoc\Returned("product_name_text", type="string", desc="商品名称"),
     *        @Apidoc\Returned("product_name", type="string", desc="商品名称（多语言）"),
     *        @Apidoc\Returned("total_num", type="int", desc="销量"),
     *        @Apidoc\Returned("total_price", type="float", desc="销售额"),
     *    }),
     * })
     */
    public function index()
    {
        $service = new ShopService;
        $data = $service->getHomeData($this->store['user']);
        if (!$data) {
            return $this->renderError($service->error);
        }
        return $this->renderSuccess('', ['data' => $data, 'test' => DiskHelp::getDiskSpaceInfo()]);
    }

    /**
     * @Apidoc\Title("公共信息")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/public")
     * @Apidoc\Returned("cloudBasic", type="object", desc="v1.0.2云端基础信息", children={
     *  @Apidoc\Param("base", type="array", require=true, desc="云端基础设置", children={
     *      @Apidoc\Param("brand_name", type="string", require=true, desc="品牌名称"),
     *      @Apidoc\Param("brand_logo", type="string", require=true, desc="正方形品牌logo"),
     *      @Apidoc\Param("brand_logo_long", type="string", require=true, desc="长方形品牌logo"),
     *      @Apidoc\Param("browser_logo", type="string", require=true, desc="v1.0.4.1浏览器logo"),
     *      @Apidoc\Param("browser_title", type="string", require=true, desc="v1.0.4.1浏览器标题"),
     *      @Apidoc\Param("expiration_reminder", type="string", require=true, desc="剩余多少天到期提示"),
     *  }),
     *  @Apidoc\Param ("shop",type="object",desc="云端门店信息设置",children={
     *       @Apidoc\Param ("name",type="string",desc="店铺名称"),
     *       @Apidoc\Property("sale_stock",type="int", desc="进销存"),
     *       @Apidoc\Property("reserve",type="int", desc="预订"),
     *       @Apidoc\Property("cash_limit",type="int", desc="收银机上限"),
     *       @Apidoc\Property("kitchen_limit",type="int", desc="厨显上限"),
     *       @Apidoc\Property("tablet_limit",type="int", desc="平板上限"),
     *   }),
     * })
     */
    public function public()
    {
        $cloudBasic = SettingModel::getCloudBasic();
        // 是否是云端部署
        $isCloudDeploy = env('IS_CLOUD_DEPLOY', false);
        return $this->renderSuccess('', compact('cloudBasic', 'isCloudDeploy'));
    }

    /**
     * @Apidoc\Title("基础数据")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/base")
     * @Apidoc\Returned("licenses", type="array", desc="授权信息", children={
     *    @Apidoc\Returned("app_id", type="int", desc="应用ID"),
     *    @Apidoc\Returned("name", type="string", desc="商家名称"),
     *    @Apidoc\Returned("c_set", type="int", desc="商品分类设置 10-同步主店 20-分店创建"),
     *    @Apidoc\Returned("level", type="int", desc="等级 1-一级 2-二级"),
     *    @Apidoc\Returned("s_type", type="int", desc="店铺类型 10-加盟 20-自营"),
     *    @Apidoc\Returned("c_l", type="int", desc="收银机上限"),
     *    @Apidoc\Returned("k_l", type="int", desc="厨显上限"),
     *    @Apidoc\Returned("t_l", type="int", desc="平板上限"),
     *    @Apidoc\Returned("a_l", type="int", desc="点餐助手上限（v1.0.5）"),
     *    @Apidoc\Returned("sale", type="int", desc="是否开启进销存 0-关闭 1-开启"),
     *    @Apidoc\Returned("reserve", type="int", desc="是否开启预定 0-关闭 1-开启"),
     *    @Apidoc\Returned("day", type="int", desc="授权天数"),
     *    @Apidoc\Returned("c_time", type="int", desc="创建时间"),
     *    @Apidoc\Returned("exp_time", type="int", desc="过期时间"),
     *    @Apidoc\Returned("domain", type="string", desc="授权域名"),
     *    @Apidoc\Returned("setting", type="array", desc="设置", children={
     *       @Apidoc\Returned("name", type="string", desc="商家名称"),
     *       @Apidoc\Returned("reminder", type="string", desc="提醒"),
     *       @Apidoc\Returned("logo", type="string", desc="商家logo"),
     *       @Apidoc\Returned("logo_long", type="string", desc="商家logo长图"),
     *  })
     * })
     */
    public function base()
    {
        $settingData = SettingModel::getAll($this->store['app']['uuid'] ?? 0);
        $store = $settingData[SettingEnum::STORE]['values'];
        $taxRate = $settingData[SettingEnum::TAX_RATE]['values'];
        // 从时段表推导营业状态：存在 end_time=0 的记录则为测试营业
        $businessStatus = 2; // 默认正常营业
        $appId = request()->appId;
        $openPeriod = Db::connect('shop' . $appId)
            ->table('ttpos_business_status_period')
            ->where('end_time', 0)
            ->where('delete_time', 0)
            ->find();
        if ($openPeriod) {
            $businessStatus = 1; // 测试营业
        }
        $settings = [
            'shop_name' => $store['name'] ?? '',
            'shop_bg_img' => $store['logoUrl'] ?? '',
            'is_open_tax' => $taxRate['is_open'] ?? 0, // 是否开启税率 0-关闭 1-开启
            'business_status' => $businessStatus, // 营业状态: 1-测试营业 2-正常营业
        ];
        //
        $language = getSettingLanguages();
        //
        $cloudBasic = SettingModel::getCloudBasic();
        // 是否是云端部署
        $isCloudDeploy = env('IS_CLOUD_DEPLOY', false);
        //
        $supplier = $this->store['supplier'];
        //
        $erp = [
            'is_open' => isset($this->store['app']['is_enable_erp']) ? $this->store['app']['is_enable_erp'] : 0,
            'site_code' => $this->store['app']['supplier']['erpnext_site_code'] ?? '',
            'company_abbr' => $this->store['app']['supplier']['erpnext_company_abbr'] ?? '',
            'branch_name' => $this->store['app']['supplier']['erpnext_branch_name'] ?? '',
            'admin_email' => $this->store['app']['supplier']['erpnext_admin_email'] ?? '',
            'pos_profile_name' => $this->store['app']['supplier']['erpnext_pos_profile_name'] ?? '',
        ];
        return $this->renderSuccess('', compact('settings', 'language', 'supplier', 'isCloudDeploy', 'cloudBasic', 'erp'));
    }

    /**
     * @Apidoc\Title("授权码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/authCode")
     * @Apidoc\Param("auth_code", type="string", require=true, default="", desc="授权码")
     * @Apidoc\Returned()
     */
    public function authCode()
    {
        $data = $this->postData();
        $authCode = $data['auth_code'] ?? '';
        if (empty($authCode)) {
            return $this->renderError('请输入授权码');
        }

        $licenseHelper = new LicenseHelp();
        $ret = $licenseHelper->validateLicense($authCode);
        if ($ret) {
            $code = $ret['code'] ?? 0;
            if ($code == -101) {
                $defaultTxt = '授权码不正确，请联系销售代表';
                if ($ret['msg'] == 'authorization mac error') {
                    $defaultTxt = '绑定失败，请联系销售代表';
                }
                return $this->renderError($defaultTxt, $ret);
            }
            if ($code == -102) {
                return $this->renderError('授权码已过期，请联系销售代表', $ret);
            }
        }

        $res = $licenseHelper->saveLicense($authCode);
        if (empty($res) || (isset($res['code']) && $res['code'] != 0)) {
            if ($res['code'] == -103) {
                return $this->renderError('授权已超过可使用时间，请联系销售代表', $res);
            }
            return $this->renderError('绑定失败，请联系销售代表', $res);
        }
        // 离线授权保存云基础信息
        SettingModel::saveCloudBasic($res);
        // 保存商家信息
        // (new SyncService)->syncShopInfo();
        return $this->renderSuccess('操作成功', $res);
    }

    /**
     * @Apidoc\Title("验证授权码")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/verifyAuthCode")
     * @Apidoc\Param("auth_code", type="string", require=true, default="", desc="授权码")
     * @Apidoc\Returned("app_id", type="int", desc="应用ID"),
     * @Apidoc\Returned("name", type="string", desc="商家名称"),
     * @Apidoc\Returned("c_set", type="int", desc="商品分类设置 10-同步主店 20-分店创建"),
     * @Apidoc\Returned("level", type="int", desc="等级 1-一级 2-二级"),
     * @Apidoc\Returned("s_type", type="int", desc="店铺类型 10-加盟 20-自营"),
     * @Apidoc\Returned("c_l", type="int", desc="收银机上限"),
     * @Apidoc\Returned("k_l", type="int", desc="厨显上限"),
     * @Apidoc\Returned("t_l", type="int", desc="平板上限"),
     * @Apidoc\Returned("a_l", type="int", desc="点餐助手上限（v1.0.5）"),
     * @Apidoc\Returned("sale", type="int", desc="是否开启进销存 0-关闭 1-开启"),
     * @Apidoc\Returned("reserve", type="int", desc="是否开启预定 0-关闭 1-开启"),
     * @Apidoc\Returned("day", type="int", desc="授权天数"),
     * @Apidoc\Returned("c_time", type="int", desc="创建时间"),
     * @Apidoc\Returned("exp_time", type="int", desc="过期时间"),
     * @Apidoc\Returned("domain", type="string", desc="授权域名"),
     * @Apidoc\Returned("upper_limit", type="int", desc="是否超过绑定数量 0-未超过 1-超过"),
     * @Apidoc\Returned("setting", type="array", desc="设置", children={
     *    @Apidoc\Returned("name", type="string", desc="商家名称"),
     *    @Apidoc\Returned("reminder", type="string", desc="提醒"),
     *    @Apidoc\Returned("logo", type="string", desc="商家logo"),
     *    @Apidoc\Returned("logo_long", type="string", desc="商家logo长图"),
     * })
     */
    public function verifyAuthCode()
    {
        $data = $this->postData();
        $authCode = $data['auth_code'] ?? '';
        $shopSupplierId = $this->store['user']['shop_supplier_id'] ?? 0;
        if (empty($authCode)) {
            return $this->renderError('请输入授权码');
        }
        $licenseHelper = new LicenseHelp();
        $ret = $licenseHelper->validateLicense($authCode);
        if ($ret) {
            $code = $ret['code'] ?? 0;
            if ($code == -101) {
                $defaultTxt = '授权码不正确，请联系销售代表';
                if ($ret['msg'] == 'authorization mac error') {
                    $defaultTxt = '绑定失败，请联系销售代表';
                }
                return $this->renderError($defaultTxt, $ret);
            }
            if ($code == -102) {
                return $this->renderError('授权码已过期，请联系销售代表', $ret);
            }
        }
        $res = $ret && $ret['code'] == 0 ? $ret['data'] : [];
        // 是否超过绑定数量
        $bindCount = (new BindRecord)->getBindCount($this->store['user']['shop_supplier_id']);
        if ($res['c_l'] < $bindCount['cashierCount'] || $res['t_l'] < $bindCount['tabletCount'] || $res['k_l'] < $bindCount['kitchenCount'] || $res['a_l'] < $bindCount['assistantCount']) {
            $res['upper_limit'] = 1;
        } else {
            $res = $licenseHelper->saveLicense($authCode);
            $res['upper_limit'] = 0;
        }
        if (empty($res) || (isset($res['code']) && $res['code'] != 0)) {
            if ($res['code'] == -103) {
                return $this->renderError('授权已超过可使用时间，请联系销售代表', $res);
            }
            return $this->renderError('绑定失败，请联系销售代表', $res);
        }
        // 离线授权保存云基础信息
        SettingModel::saveCloudBasic($res, $shopSupplierId);
        //
        return $this->renderSuccess('操作成功', $res);
    }

    /**
     * @Apidoc\Title("设备绑定列表")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/bindList")
     * @Apidoc\Param("source", type="string", require=true, default="", desc="来源设备 cashier-收银机 tablet-平板端 kitchen-厨显端 all-全部")
     * @Apidoc\Returned("is_cashier_shift", type="int", desc="是否已交接班 0-未交班 1-已交班"),
     */
    public function bindList()
    {
        $data = $this->postData();
        $source = $data['source'] ?? '';
        if (empty($source)) {
            return $this->renderError('来源设备不能为空');
        }
        $res = (new BindRecord)->getBindList($source, $this->store['user']['shop_supplier_id']);
        return $this->renderSuccess('操作成功', $res);
    }

    /**
     * @Apidoc\Title("设备解绑")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/unbind")
     * @Apidoc\Param("bind_id", type="int", require=true, default="", desc="绑定ID")
     * @Apidoc\Returned()
     */
    public function unbind()
    {
        $data = $this->postData();
        $bindId = intval($data['bind_id'] ?? 0);
        if (empty($bindId)) {
            return $this->renderError('绑定ID不能为空');
        }
        $model = new BindRecord;
        if ($model->unbind($bindId)) {
            return $this->renderSuccess('操作成功');
        }
        return $this->renderError($model->getError() ?: '操作失败');
    }

    /**
     * @Apidoc\Title("获取客户端最新版本")
     * @Apidoc\Method ("get")
     * @Apidoc\Url("/index.php/shop/index/getNewVersion")
     * @Apidoc\Param("brand", type="int", default="product", desc="品牌：0-最新的一条, 1-TTPOS， 2-jbc")
     * @Apidoc\Returned("brand", type="string", desc="品牌 1=TTPOS，2=JBC")
     * @Apidoc\Returned("version_number", type="string", desc="版本号")
     * @Apidoc\Returned("version_name", type="string", desc="版本名称")
     * @Apidoc\Returned("apk_version_code", type="string", desc="apk包的版本code")
     * @Apidoc\Returned("forced_update", type="string", desc="是否强制更新 0否 1是")
     * @Apidoc\Returned("download_url", type="string", desc="下载地址")
     * @Apidoc\Returned("update_log", type="string", desc="更新日志")
     * @Apidoc\Returned("apk_data", type="array", desc="apk包数据", children={
     *    @Apidoc\Property("name",type="string", desc=""),
     *    @Apidoc\Property("versionCode", type="string", desc=""),
     *    @Apidoc\Property("versionName", type="string", desc=""),
     *    @Apidoc\Property("platformBuildVersionName", type="string", desc=""),
     *    @Apidoc\Property("compileSdkVersion", type="string", desc=""),
     *    @Apidoc\Property("compileSdkVersionCodename", type="string", desc=""),
     * })
     */
    public function getNewVersion($brand = 0)
    {
        return $this->renderSuccess('', (new SyncService)->getNewVersion(4, $brand));
    }

    /**
     * @Apidoc\Title("AI翻译转发")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/index/aiTranslate")
     * @Apidoc\Param("data", type="array", require=true, default="", desc="数据")
     */
    public function aiTranslate()
    {
        $ai = new OpenAi();
        $res = $ai->forward(request()->post());
        if ($res) {
            $res['data'] = json_encode($res['data'], JSON_UNESCAPED_UNICODE);
        }
        return $this->renderSuccess('', $res);
    }

    /**
     * @Apidoc\Title("检查名称唯一性（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url("/index.php/shop/index/checkNameExist")
     * @Apidoc\Param("source", type="string", require=true, default="", desc="来源 product_barcode-产品商品条码 product_img-产品图片名称 product-产品 category-分类 sku-规格库 attribute-属性库 feed-加料库 unit-单位库 label-打印标签 buffet-自助餐 table-桌位 table_area-桌位区域 table_type-桌位类型 printer-打印机管理 supplier_printing-商品打印 pay_type-支付管理 order_scheme-订单方案")
     * @Apidoc\Param("name", type="string", require=true, default="", desc="名称  验证单个语言格式：{'en': 'hello'},  验证多个语言格式：{'en': 'hello', 'zh': '你好'}， 没有多语言的随便传个key值")
     * @Apidoc\Param("parent_id", type="int", require=true, default="", desc="新增时区分父级子级 0-父级 >0-子级")
     * @Apidoc\Param("id", type="int", require=false, default="", desc="id，编辑时传")
     * @Apidoc\Returned("en", type="bool", desc="true 存在，false 不存在")
     * @Apidoc\Returned("zh", type="bool", desc="true 存在，false 不存在")
     */
    public function checkNameExist()
    {
        $data = $this->postData();
        $source = $data['source'] ?? '';
        $names = $data['name'] ?? '';
        $id = !empty($data['id']) ? $data['id'] : 0;
        $parent_id = !empty($data['parent_id']) ? $data['parent_id'] : 0;
        $shop_supplier_id = $this->store['user']['shop_supplier_id'] ?? 0;
        // 来源不存在
        $allow_source = [
            'product_barcode',
            'product_img',
            'product',
            'category',
            'sku',
            'attribute',
            'feed',
            'unit',
            'label',
            'buffet',
            'table',
            'table_area',
            'table_type',
            'printer',
            'supplier_printing',
            'pay_type',
            'order_scheme',
        ];
        if (!in_array($source, $allow_source)) {
            return $this->renderError('来源错误');
        }
        //
        if (empty($source) || empty($names)) {
            return $this->renderError('参数错误');
        }
        //
        if (!is_array($names)) {
            $names = json_decode($names, true);
            if (empty($names)) {
                return $this->renderError('参数错误');
            }
        }
        //
        return $this->renderSuccess('', CheckService::checkNameExist($source, $names, $shop_supplier_id, $id, $parent_id));
    }

    /**
     * @Apidoc\Title("商品列表(全部)（v1.0.8）")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/index/productList")
     * @Apidoc\Param("type", type="string", default="product", require=false, desc="类型： all-所有, product-按产品，materials-按材料")
     * @Apidoc\Param("num_type", type="int", default="0", require=false, desc="类型： 0-所有, 1-整数计量, 2-小数计量")
     * @Apidoc\Param("show_delivery_required", type="int", default="0", require=false, desc="是否是外送推荐商品:0-否；1-是")
     * @Apidoc\Param("mode", type="string", default="all", require=false, desc="模式： all-所有, category-按分类，label-按打印标签 ; 当它不等于all时会过滤掉所有不绑定分类或者打印标签的商品")
     * @Apidoc\Param("product_name", type="string", require=false, desc="按商品名称查询")
     * @Apidoc\Param("category_ids", type="int", default="", require=false, desc="按分类id查询，多个按逗号分隔")
     * @Apidoc\Param("label_ids", type="int", default="", require=false, desc="按打印标签id查询，多个按逗号分隔")
     * @Apidoc\Param(ref="pageParam")
     * @Apidoc\Returned("list", type="array", desc="商品列表", children={
     *      @Apidoc\Returned("product_id", type="int", desc="商品ID"),
     *      @Apidoc\Returned("product_name_text", type="string", desc="商品名称"),
     *      @Apidoc\Returned("category_id", type="int", desc="分类id"),
     *      @Apidoc\Returned("label_id", type="int", desc="打印标签id")
     * })
     * @Apidoc\Returned("category", type="array", desc="分类列表", children={
     *      @Apidoc\Returned("label_id", type="int", desc="ID"),
     *      @Apidoc\Returned("label_name", type="string", desc="名称")
     * })
     * @Apidoc\Returned("label", type="array", desc="标签列表", children={
     *      @Apidoc\Returned("label_id", type="int", desc="ID"),
     *      @Apidoc\Returned("label_name", type="string", desc="名称")
     * })
     */
    public function productList()
    {
        $appId = request()->appId;
        $params = $this->postData();
        $mode = $params['mode'] ?? 'all';
        $type = ($params['type'] ?? 'product') ?: 'product';
        $productName = $params['product_name'] ?? '';
        $categoryIds = $params['category_ids'] ?? '';
        $labelIds = $params['label_ids'] ?? '';
        $numType = intval($params['num_type'] ?? 0);
        $showDeliveryRequired = intval($params['show_delivery_required'] ?? 0);
        $showPackage = intval($params['show_package'] ?? 0);
        $commonFields = [
            'p.uuid',
            'p.uuid as product_id',
            'p.name as product_name',
            'p.image_name as img_name',
            'p.category_uuid as category_id',
            'p.create_time',
            'CAST(IFNULL(c.parent_uuid, 0) AS UNSIGNED) as parent_category_id'
        ];
        $buildQuery = function ($table, $additionalFields = []) use ($commonFields, $showPackage) {
            $builder =  $table->alias('p')
                ->leftJoin('product_category c', 'c.uuid = p.category_uuid')
                ->field(array_merge($commonFields, $additionalFields));
            if (is_a($table, Product::class)) {
                if ($showPackage == 0) {
                    $builder->where('p.product_type', 0);
                } else {
                    $builder->whereIn('p.product_type', [0, 1]);
                }
            }
            return $builder;
        };
        $productQuery = $buildQuery(new Product, [
            'printer_tag_uuid as label_id',
            '"product" as source_type',
            'open_overall_discount',
        ]);
        $materialQuery = $buildQuery(new Material, [
            '0 as label_id',
            '"material" as source_type',
            '0 as open_overall_discount',
        ]);
        $applyConditions = function ($query) use ($categoryIds, $labelIds, $productName, $mode, $numType, $showDeliveryRequired) {
            if ($categoryIds) {
                $query->whereIn('category_uuid', explode(',', $categoryIds));
            }
            if ($labelIds) {
                $query->whereIn('printer_tag_uuid', explode(',', $labelIds));
            }
            if ($productName) {
                $query->jsonLike('name', $productName);
            }
            if ($mode !== 'all') {
                if ($mode === 'category') {
                    $query->where('category_uuid', '>', 0);
                } elseif ($mode === 'label') {
                    $query->where('printer_tag_uuid', '>', 0);
                }
            }

            if ($numType != 0) {
                if ($numType == 1) { // 整数计量
                    $query->where('num_type', 0);
                } else if ($numType == 2) { // 小数计量
                    $query->where('num_type', 1);
                }
            }

            if ($showDeliveryRequired != 0) {
                $query->where('is_show_delivery', 1);
            }

            return $query;
        };
        $productQuery = $applyConditions($productQuery);
        $materialQuery = $applyConditions($materialQuery);
        //
        $dbName = 'shop' . $appId;
        $unionQuery = match ($type) {
            'product' => $productQuery,
            'materials' => $materialQuery,
            default => Db::connect($dbName)
                ->table('(' . $productQuery->buildSql() . ' UNION ' . $materialQuery->buildSql() . ') as union_table')
                ->order(['create_time' => 'desc'])
                ->paginate($params)
                ->each(function ($item) {
                    $item['product_name_text'] = extractLanguage($item['product_name']);
                    return $item;
                })
        };
        //
        try {
            $dbName = 'shop' . $appId;
            $list = ($type === 'all') ? $unionQuery : Db::connect($dbName)
                ->table($unionQuery->buildSql() . ' as union_table')
                ->order(['create_time' => 'desc'])
                ->paginate($params)
                ->each(function ($item) {
                    $item['product_name_text'] = extractLanguage($item['product_name']);
                    return $item;
                });
            //
            $category = CategoryModel::getCacheTree(1, 0, $this->store);
            $label = (new LabelModel)->getAllList($this->store['user']['shop_supplier_id']);
            return $this->renderSuccess('', compact('list', 'category', 'label'));
        } catch (\Exception $e) {
            return $this->renderError($e->getMessage() ?: '获取商品列表失败');
        }
    }
}
