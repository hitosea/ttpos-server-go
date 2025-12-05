<?php

namespace app\admin\model\app;

use app\common\enum\settings\SettingEnum;
use app\common\model\settings\Setting;
use app\common\model\shop\SaasUser;
use PDO;
use think\facade\Cache;
use think\facade\Env;
use think\facade\Validate;
use app\admin\model\CompanyStaff;
use app\common\model\store\PayType;
use hg\apidoc\annotation as Apidoc;
use app\common\model\app\App as AppModel;
use app\common\enum\order\OrderPayTypeEnum;
use app\common\service\websocket\Websocket;
use app\common\model\shop\User as ShopStaffModel;
use app\admin\model\supplier\Supplier as SupplierModel;
use help\HttpHelp;

class App extends AppModel
{
    /**
     * 获取小程序列表
     * @Apidoc\field("app_id,auth_day,expire_time,status,create_time")
     * @Apidoc\AddField("id",type="int",default="0",desc="列表ID")
     * @Apidoc\AddField("suppliers_total",type="int",default="0",desc="子店铺总数, 商家类型通过这个来判断显示")
     * @Apidoc\AddField("suppliers",type="array",default="[]",desc="子列表")
     * @Apidoc\AddField("user_name",type="string",default="000000",desc="超管邮箱")
     * @Apidoc\AddField("shop_supplier_id",type="int",default="0",desc="线上店铺id")
     * @Apidoc\AddField("shop_supplier_name",type="int",default="0",desc="线上店铺名称")
     * @Apidoc\AddField("parent_id",type="int",default="0",desc="父id")
     * @Apidoc\AddField("link_phone",type="int",default="0",desc="联系电话")
     * @Apidoc\AddField("logo",type="int",default="0",desc="logo")
     * @Apidoc\AddField("level",type="int",default="0",desc="层级")
     * @Apidoc\AddField("sale_stock",type="int",default="0",desc="进销存: 0不开启, 1开启")
     * @Apidoc\AddField("reserve",type="int",default="0",desc="预订: 0不开启, 1开启")
     * @Apidoc\AddField("cash_limit",type="int",default="0",desc="收银机上限")
     * @Apidoc\AddField("kitchen_limit",type="int",default="0",desc="厨显上限")
     * @Apidoc\AddField("tablet_limit",type="int",default="0",desc="平板上限")
     * @Apidoc\AddField("address",type="int",default="0",desc="联系地址")
     * @Apidoc\AddField("coordinates",type="string",default="",desc="经纬度，如：13.721899,100.52900")
     * @Apidoc\AddField("store_type",type="int",default="0",desc="店铺类型10加盟20自营")
     * @Apidoc\AddField("category_set",type="int",default="0",desc="商品分类设置10同步 主店20分店创建")
     */
    public function getList($param)
    {
        $status = $param['status'] ?? 0;
        $keyword = $param['keyword'] ?? '';
        $appId = $param['app_id'] ?? 0;
        //
        $field = [
            "app.uuid as id",
            "app.uuid as app_id",
            "app.auth_day",
            "app.auth_start_time",
            "app.expire_time",
            "app.status",
            "app.create_time",
            "app.name as shop_supplier_name",
            "app.logo",
            //
            "user.username as user_name",
            //
            "su.link_phone",
            "su.sale_stock",
            "su.is_open_tax",
            "su.cash_limit",
            "su.kitchen_limit",
            "su.tablet_limit",
            "su.address",
            "su.assistant_limit",
            //
            "su.company_uuid as shop_supplier_id",
            "su.is_open_member",
            "su.is_open_tablet",
            "su.is_open_coupon",
            "su.is_open_marketing",
            "su.is_open_h5 as is_open_scan",
            "su.is_open_assistant",
            "su.is_open_kitchen_kds",
            "su.is_open_buffet",
            "su.is_open_h5_order as is_accept_scan_order",
            "su.is_open_local_print",
            "su.is_open_advanced_ticket_print",
            "su.table_limit",
            "su.printer_limit",
            "su.languages",
            "su.timezone",
            "su.enable_sms",
            "su.sms_quota",
            "su.coordinates",
            // v2.10.0
            "su.enable_table_map",
            "su.enable_data_management",
            // v2.11.0
            "su.enable_kiosk",
        ];
        //
        $countWhere = 'where 1 = 1';
        if ($keyword) {
            $countWhere .= ($countWhere ? ' and' : ' where') . ' (su.name like "%' . $keyword . '%" OR su.app_id like "%' . $keyword . '%")';
        }
        if ($status > 0) {
            $st = $status == 1 ? 1 : 0;
            $countWhere .= ($countWhere ? ' and' : ' where') . " (app.status = $st)";
        }
        //
        return $this->alias('app')
            ->field($field)
            ->leftJoin('company_staff user', "user.company_uuid = app.uuid and user.is_super = 1")
            ->leftJoin('company_setting su', "su.company_uuid = app.uuid")
            ->when($keyword, function ($q) use ($keyword) {
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('app.name', $keyword);
                    $qq->orLike('app.uuid', $keyword);
                });
            })
            ->when($status > 0, function ($q) use ($status) {
                $q->where(function ($q) use ($status) {
                    $q->where('app.status', '=', $status == 1 ? 1 : 0);
                });
            })
            ->when($appId, function ($q) use ($appId) {
                $q->where('app.uuid', '=', $appId);
            })
            ->order(["app.create_time" => 'desc'])
            ->group('app.uuid')
            ->paginate($param);
    }

    /**
     * 下拉选择列表
     */
    public function getSelectList($param)
    {
        $keyword = $param['keyword'] ?? '';
        //
        return $this->alias('app')
            ->field("su.shop_supplier_id as id,su.name")
            ->leftJoin('shop_user user', "user.app_id = app.app_id")
            ->leftJoin('supplier su', "su.app_id = app.app_id")
            ->where('user.is_super', '=', 1)
            ->where('su.parent_id', '=', 0)
            ->where('su.delete_time', '=', 0)
            ->where('app.delete_time', '=', 0)
            ->when($keyword, function ($q) use ($keyword) {
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('su.name', $keyword);
                    $qq->orLike('su.shop_supplier_id', $keyword);
                });
            })
            ->order(["su.create_time" => 'asc'])
            ->group('app.app_id')
            ->paginate($param)
            ->append([]);
    }


    /**
     * 外送商家列表
     */
    public function getDeliveryList($param)
    {
        $keyword = $param['keyword'] ?? '';
        $channel = $param['channel'] ?? '';
        $configured = $param['configured'] ?? false;

        return $this->alias('app')
            ->field("su.uuid, app.name, su.delivery_status, su.delivery_config")
            ->leftJoin('company_setting su', "su.company_uuid = app.uuid")
            ->where('su.delete_time', '=', 0)
            ->where('app.delete_time', '=', 0)
            ->when($configured, function ($q) {
                return $q->where('su.delivery_config', '<>', "");
            })
            ->where("app.uuid", "in", function ($q) {
                return $q->name('payment_app')->where("delete_time", "=", 0)->field('company_uuid');
            })
            ->when($keyword, function ($q) use ($keyword) { // 商家名称关键字
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('app.name', $keyword);
                    $qq->orLike('app.uuid', $keyword);
                });
            })
            ->when($channel && $configured, function ($q) use ($channel) { // 外送渠道
                $q->whereRaw('JSON_CONTAINS(su.delivery_config, ?)', [json_encode(['channel' => $channel])]);
            })
            ->order(["su.create_time" => 'asc'])
            ->group('app.uuid')
            ->paginate($param)
            ->append([]);
    }


    /**
     * ERPNext商家列表
     */
    public function getErpnextCompanyList($param)
    {
        $keyword = $param['keyword'] ?? '';
        $configured = $param['configured'] ?? false;

        return $this->alias('app')
            ->field("su.uuid, app.name, su.link_phone, su.real_name, su.erpnext_site_code, su.erpnext_company_abbr, su.erpnext_branch_name, su.erpnext_pos_profile_name")
            ->leftJoin('company_setting su', "su.company_uuid = app.uuid")
            ->where('su.delete_time', '=', 0)
            ->where('app.delete_time', '=', 0)
            ->when($configured, function ($q) {
                return $q->where('su.erpnext_site_code', '<>', "")
                    ->where('su.erpnext_company_abbr', '<>', "")
                    ->where('su.erpnext_branch_name', '<>', "")
                    ->where('su.erpnext_pos_profile_name', '<>', "");
            })
            ->when($keyword, function ($q) use ($keyword) { // 商家名称关键字
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('app.name', $keyword);
                    $qq->orLike('app.uuid', $keyword);
                });
            })
            ->when(!$configured, function ($q) {
                return $q->where(function ($qq) {
                    $qq->where('app.expire_time', '=', "0")->whereOr('app.expire_time', '>', time())->where('app.status', '=', 1);
                });
            })
            ->order(["su.create_time" => 'asc'])
            ->group('app.uuid')
            ->paginate($param)
            ->append([]);
    }

    /**
     * 新增记录
     */
    public function add($data)
    {
        if (CompanyStaff::checkPhoneExist($data['link_phone'])) {
            $this->error = '联系电话已存在';
            return false;
        }

        if (CompanyStaff::checkExist($data['user_name'])) {
            $this->error = '超管邮箱已存在';
            return false;
        }

        //
        $authDay = $data['auth_day'] ?? 0;
        if (!filter_var($authDay, FILTER_VALIDATE_INT) && $authDay != 0) {
            $this->error = '授权时间需为正整数';
            return false;
        }
        $data['auth_start_time'] = $authStartTime = strtotime($data['auth_start_time']) ?? 0;
        $data['expire_time'] = $authDay ? strtotime("+$authDay days", $authStartTime) : 0;
        $data['pay_type'] = json_encode(array_keys(OrderPayTypeEnum::pay()));

        //
        $this->startTrans();
        try {
            // 添加集团
            $data['uuid'] = createUuid();
            $this->allowField(['uuid', 'name', 'logo', 'auth_day', 'expire_time', 'auth_start_time'])->save($data);

            //新增门店
            $supplierModel = new SupplierModel;
            if (!$supplierModel->add($data, $this->app_id)) {
                $this->error = $supplierModel->error;
                $this->rollback();
                return false;
            }
            //
            $this->commit();
            //
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 修改记录
     */
    public function edit($data)
    {
        $authDay = $data['auth_day'] ?? 0;
        if (!filter_var($authDay, FILTER_VALIDATE_INT) && $authDay != 0) {
            $this->error = '授权时间需为正整数';
            return false;
        }
        $authStartTime = $data['auth_start_time'] = strtotime($data['auth_start_time']) ?? 0;
        $companyUuid = $this->uuid;
        //
        $save_data = [
            'name' => $data['name'] ?? '',
            'auth_day' => $authDay,
            'auth_start_time' => $authStartTime,
            'expire_time' => $authDay ? strtotime("+$authDay days", $authStartTime) : 0,
        ];
        //
        $this->startTrans();
        try {
            // 用户
            $user_data = [
                'username' => $data['user_name'],
                'phone' => $data['link_phone'] ?? '',
            ];
            if (!empty($data['password'])) {
                $validate = Validate::rule('password', 'checkPassword');
                if (!$validate->check($data)) {
                    $this->error = $validate->getError();
                    return false;
                }
                $user_data['password'] = salt_hash($data['password']);
            }
            //
            $shop_user = CompanyStaff::withoutGlobalScope()->where('company_uuid', '=', $this['uuid'])->order('is_super', 'desc')->find();
            if ($shop_user['phone'] != $data['link_phone']) {
                if (CompanyStaff::checkPhoneExist($data['link_phone'])) {
                    $this->error = '联系电话已存在';
                    return false;
                }
            }
            if ($shop_user['username'] != $data['user_name']) {
                if (CompanyStaff::checkExist($data['user_name'])) {
                    $this->error = '超管邮箱已存在';
                    return false;
                }
            }

            // 平台
            $this->save($save_data);

            // 商户
            if (isset($data['logo'])) unset($data['logo']);
            $data['is_open_h5'] = $data['is_open_scan'] ?? 0;
            $data['is_open_h5_order'] = $data['is_accept_scan_order'] ?? 0;
            if (($data['enable_sms'] ?? 0) == 0) {
                $data['sms_quota'] = 0;
            }
            $companySetting = SupplierModel::where('company_uuid', '=', $this['uuid'])->find();
            $companySetting?->save($data);

            // 用户
            $shop_user->save($user_data);

            // 同步商家会员信息
            $staff['username'] = $shop_user->getData('username');
            $staff['phone'] = $shop_user->getData('phone');
            if (isset($data['password']) && $data['password'] != '') {
                $staff['password'] = salt_hash($data['password']);
            }
            (new ShopStaffModel([], $companyUuid))->setAppId($companyUuid)->where('uuid', $shop_user->uuid)->find()?->save($staff);

            // 会员设置 - 支付方式关联
            if (isset($data['is_open_member'])) {
                (new PayType([], $companyUuid))->setAppId($companyUuid)->setStatus(PayType::BALANCE_VALUE, $data['is_open_member']);
            }


            // 修改总部供应商名称
            if (
                $this->is_enable_erp == 1 &&
                !($companySetting->erpnext_site_code == "1" || $companySetting->erpnext_site_code == "") &&
                $companySetting->erpnext_company_abbr == $companySetting->erpnext_headquarter_abbr
            ) {
                $res = HttpHelp::postRequest('http://nginx/api/v1/admin/erpnext/shop/update_headquarter_supplier_name', json_encode([
                    'site_code' => $companySetting->erpnext_site_code,
                    'company_abbr' => $companySetting->erpnext_company_abbr,
                    'branch' => $companySetting->erpnext_branch_name,
                    'supplier_name' => $data['name'],
                    'company_uuid' => $companySetting->company_uuid,
                ]), [
                    'X-API-KEY: ' . env('JWT_SECRET'),
                    'Accept-Language: ' . request()->header('language'),
                ], 60);

                if (!$res) {
                    $this->error = '请求失败';
                    return false;
                }
                $res = json_decode($res, true);
                if ($res['code'] != 0) {
                    $this->error = $res['message'];
                    return false;
                }
            }

            //
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 软删除
     */
    public function setDelete()
    {
        $deleteTime = time();
        $host = env('DB_HOST');
        $port = env('DB_PORT');
        $prefix = env('DB_PREFIX');
        $pdo = new PDO("mysql:host={$host};port={$port}", env('DB_USERNAME'), env('DB_PASSWORD'));
        $databaseName = 'shop' . $this->uuid;
        $pdo->exec("use {$databaseName}");
        $res = $pdo->exec("UPDATE {$prefix}company SET delete_time = $deleteTime WHERE uuid = {$this->uuid}");
        if ($res === false) {
            return false;
        }

        $this->supplier?->delete();

        // 软删除saas商家员工
        $shopUsers = SaasUser::where('company_uuid', '=', $this['uuid'])->select();
        foreach ($shopUsers as $shopUser) {
            $shopUser->delete();
        }

        // 推送配置更新
        Websocket::pushClient(
            $this->uuid,
            Websocket::SOURCE_All,
            Websocket::SOURCE_All,
            Websocket::UPDATE_CONFIG,
            0,
            ['update_time' => time()]
        );

        return $this->delete();
    }
}
