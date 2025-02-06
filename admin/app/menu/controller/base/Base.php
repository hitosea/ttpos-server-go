<?php

namespace app\menu\controller\base;

use app\common\model\order\Order;
use hg\apidoc\annotation as Apidoc;
use app\menu\controller\Controller;
use app\common\model\store\Table;
use app\common\model\supplier\Supplier as SupplierModel;


/**
 * 基础
 */
class Base extends Controller
{
    /**
     * @Apidoc\Title("基础信息")
     * @Apidoc\Desc("基础信息")
     * @Apidoc\Method("POST")
     * @Apidoc\Url("/index.php/menu/base.base/getInfo")
     * @Apidoc\Returned("currency", type="object", desc="货币", children={
     *    @Apidoc\Param ("unit",type="string",desc="主单位"),
     *    @Apidoc\Param ("unit_position",type="string",desc="v1.0.3主货币显示位置 0-金额前 1-金额后"),
     *    @Apidoc\Param ("is_open",type="string",desc="是否开启副单位 0-关闭 1-开启"),
     *    @Apidoc\Param ("vices",type="object",desc="副单位",children={
     *       @Apidoc\Param ("vice_unit",type="string",desc="副单位"),
     *       @Apidoc\Param ("vice_unit_position",type="string",desc="v1.0.3副货币显示位置 0-金额前 1-金额后"),
     *       @Apidoc\Param ("unit_rate",type="float",desc="副单位比例"),
     *   }),
     * })
     * @Apidoc\Returned("menu", type="object", desc="电子菜单设置", children={
     *    @Apidoc\Param ("language",type="array",desc="常用语言，默认th, en, zh, zhtw"),
     *    @Apidoc\Param ("default_language",type="array",desc="默认语言，默认en"),
     * })
     * @Apidoc\Returned("cloud_basic", type="object", desc="v1.0.2云端基础信息", children={
     *   @Apidoc\Param("base", type="array", require=true, desc="云端基础设置", children={
     *     @Apidoc\Param("brand_name", type="string", require=true, desc="品牌名称"),
     *     @Apidoc\Param("brand_logo", type="string", require=true, desc="正方形品牌logo"),
     *     @Apidoc\Param("brand_logo_long", type="string", require=true, desc="长方形品牌logo"),
     *     @Apidoc\Param("browser_logo", type="string", require=true, desc="v1.0.4.1浏览器logo"),
     *     @Apidoc\Param("browser_title", type="string", require=true, desc="v1.0.4.1浏览器标题"),
     *     @Apidoc\Param("expiration_reminder", type="string", require=true, desc="剩余多少天到期提示"),
     *  }),
     * })
     * })
     */
    public function getInfo()
    {
        $baseInfo = SupplierModel::getMenuBaseInfo();
        return $this->renderSuccess('基础信息', compact('baseInfo'));
    }
}
