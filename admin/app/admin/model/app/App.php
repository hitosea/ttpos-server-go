<?php

namespace app\admin\model\app;

use think\facade\Env;
use think\facade\Validate;
use app\common\model\store\PayType;
use hg\apidoc\annotation as Apidoc;
use app\admin\model\Shop as ShopUser;
use app\common\model\app\App as AppModel;
use app\common\enum\order\OrderPayTypeEnum;
use app\admin\model\supplier\Supplier as SupplierModel;

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
     * @Apidoc\AddField("store_type",type="int",default="0",desc="店铺类型10加盟20自营")
     * @Apidoc\AddField("category_set",type="int",default="0",desc="商品分类设置10同步 主店20分店创建")
     */
    public function getList($param)
    {
        $prefix = Env::get('DB_PREFIX');
        $shopType = $param['shop_type'] ?? 0;
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
            "su.cash_limit",
            "su.kitchen_limit",
            "su.tablet_limit",
            "su.address",
            "su.assistant_limit",
            //
            "su.is_open_member",
            "su.is_open_tablet",
            "su.is_open_scan",
            "su.is_open_assistant",
            "su.is_open_kitchen_kds",
            "su.is_open_buffet",
            "su.is_accept_scan_order",
            "su.is_open_local_print",
            "su.table_limit",
            "su.printer_limit",
            "su.languages",
            "su.timezone",
        ];
        //
        $countWhere = 'where su.is_delete = 0';
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
            ->leftJoin('company_staff user', "user.company_uuid = app.uuid")
            ->leftJoin('company_setting su', "su.company_uuid = app.uuid")

            ->when($keyword, function ($q) use ($keyword) {
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('su.name', $keyword);
                    $qq->orLike('su.company_uuid', $keyword);
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
            ->where('su.is_delete', '=', 0)
            ->where('app.is_delete', '=', 0)
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
     * 新增记录
     */
    public function add($data)
    {
        if (ShopUser::checkPhoneExist($data['link_phone'])) {
            $this->error = '联系电话已存在';
            return false;
        }

        if (ShopUser::checkExist($data['user_name'])) {
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
        $appId = $this->uuid;
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
            $shop_user = ShopUser::withoutGlobalScope()->where('company_uuid', '=', $this['uuid'])->find();
            //
            if ($shop_user['phone'] != $data['link_phone']) {
                if (ShopUser::checkPhoneExist($data['link_phone'])) {
                    $this->error = '联系电话已存在';
                    return false;
                }
            }
            if ($shop_user['username'] != $data['user_name']) {
                if (ShopUser::checkExist($data['user_name'])) {
                    $this->error = '超管邮箱已存在';
                    return false;
                }
            }

            // 平台
            $this->save($save_data);

            // 用户
            $shop_user->save($user_data);

            // 商户
            if (isset($data['logo'])) unset($data['logo']);
            SupplierModel::where('company_uuid', '=', $this['uuid'])->find()?->save($data);

            // 会员设置 - 支付方式关联
            // if (isset($data['is_open_member'])) {
            //     (new PayType([], $appId))->setStatus(PayType::BALANCE_VALUE, $data['is_open_member']);
            // }

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
        $this->supplier?->delete();
        return $this->delete();
    }
}
