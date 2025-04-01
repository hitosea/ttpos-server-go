<?php

namespace app\shop\controller\setting;

use help\HttpHelp;
use app\shop\controller\Controller;
use hg\apidoc\annotation as Apidoc;
use app\common\model\shop\BindRecord;
use app\common\enum\settings\SettingEnum;
use app\shop\model\settings\Printer as PrinterModel;
use app\shop\model\settings\Setting as SettingModel;

/**
 * 加密设置
 * @Apidoc\Group("supplier")
 * @Apidoc\Sort(6)
 */
class EncryptionConfig extends Controller
{

    /**
     * @Apidoc\Title("获取加密设置")
     * @Apidoc\Method ("GET")
     * @Apidoc\Url ("/index.php/shop/setting.EncryptionConfig/index")
     * @Apidoc\Param("uuid", type="string", require=true, default="", desc="uuid")
     * @Apidoc\Param("cash_box_id", type="string", require=true, default="", desc="收银机箱标识")
     * @Apidoc\Param("access_token", type="string", require=true, default="", desc="收银机访问令牌")
     * @Apidoc\Param("queue_url", type="string", require=true, default="", desc="收银机队列地址")
     * @Apidoc\Returned("data", type="array", desc="授权信息", children={
     *    @Apidoc\Returned("cash_sn", type="string", desc="收银机序列号"),
     *    @Apidoc\Returned("brand", type="string", desc="品牌"),
     *    @Apidoc\Returned("cash_sign", type="string", desc="收银机标识"),
     *    @Apidoc\Returned("cash_box_id", type="string", desc="收银机箱标识"),
     *    @Apidoc\Returned("access_token", type="string", desc="收银机访问令牌"),
     *    @Apidoc\Returned("queue_url", type="string", desc="收银机队列地址"),
     *    @Apidoc\Returned("uuid", type="string", desc="uuid"),
     * })
     */
    public function index()
    {
        $cashierList = BindRecord::where('source', BindRecord::SOURCE_CASHIER)
            ->field("device_id as cash_sn")
            ->field("if(brand='', '-', brand) as brand")
            ->field("cash_sign, cash_box_id, access_token, queue_url")
            ->field("uuid")
            ->order('id')
            ->select()
            ->toArray();
        // 生成cash_sign
        foreach ($cashierList as &$cashier) {
            if (empty($cashier['cash_sign'])) {
                $cashier['cash_sign'] = 'POS' . substr(md5($cashier['cash_sn']), 0, 16);
                BindRecord::where('device_id', $cashier['cash_sn'])->update(['cash_sign' => $cashier['cash_sign']]);
            }
        }
        return $this->renderSuccess('', $cashierList);
    }

    /**
     * @Apidoc\Title("保存加密设置")
     * @Apidoc\Method ("POST")
     * @Apidoc\Url ("/index.php/shop/setting.EncryptionConfig/save")
     * @Apidoc\Param("uuid", type="string", require=true, default="", desc="uuid")
     * @Apidoc\Param("cash_box_id", type="string", require=true, default="", desc="收银机箱标识")
     * @Apidoc\Param("access_token", type="string", require=true, default="", desc="收银机访问令牌")
     * @Apidoc\Param("queue_url", type="string", require=true, default="", desc="收银机队列地址")
     * @Apidoc\Returned()
     */
    public function save()
    {
        $postData = $this->postData();
        // 
        if (!$postData['cash_box_id'] || !$postData['access_token'] || !$postData['queue_url']) {
            return $this->renderError('参数错误');
        }
        // 
        $cashier = BindRecord::where('source', BindRecord::SOURCE_CASHIER)
            ->where("uuid", $postData['uuid'])
            ->order('id')
            ->find();
        if (!$cashier) {
            return $this->renderError('参数错误');
        }
        if ($cashier->cash_box_id) {
            return $this->renderError('不能重覆设置');
        }
        // 
        $cashier->cash_box_id = $postData['cash_box_id'];
        $cashier->access_token = $postData['access_token'];
        $cashier->queue_url = $postData['queue_url'];
        // 绑定
        $res = HttpHelp::postRequest($cashier->queue_url . '/json/Echo', [], array(
            'cashboxid' => $cashier->cash_box_id, 
            'accesstoken' => $cashier->access_token
        ));
        if ($res === false) {
            return $this->renderError('绑定失败，配置错误');
        }
        // 保存
        $cashier->save();
        // 
        return $this->renderSuccess('操作成功');
    }
}
