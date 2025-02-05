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
            "app.app_id as id",
            "app.app_id",
            "app.auth_day",
            "app.auth_start_time",
            "app.expire_time",
            "app.status",
            "app.create_time",
            //
            "user.user_name",
            //
            "su.shop_supplier_id",
            "su.name as shop_supplier_name",
            "su.parent_id",
            "su.link_phone",
            "su.logo",
            "su.level",
            "su.sale_stock",
            "su.reserve",
            "su.cash_limit",
            "su.kitchen_limit",
            "su.tablet_limit",
            "su.address",
            "su.store_type",
            "su.is_main",
            "su.category_set",
            "su.mac_addr",
            "su.serial_number",
            "su.deploy_mode",
            "su.assistant_limit",
            "su.chain_number",
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
            ->field('ifnull(children_apps.count,0) as suppliers_total')
            ->with(['suppliers' => function ($q)  use ($field, $status, $keyword) {
                $q->alias('su')->field($field)
                    ->leftJoin('app app', "app.app_id = su.app_id")
                    ->leftJoin('shop_user user', "user.app_id = su.app_id")
                    ->where('user.is_super', '=', 1)
                    ->where('su.is_delete', '=', 0)
                    ->where('app.is_delete', '=', 0)
                    ->when($keyword, function ($q) use ($keyword) {
                        $q->where(function ($qq) use ($keyword) {
                            $qq->like('su.name', $keyword);
                            $qq->orLike('su.app_id', $keyword);
                        });
                    })
                    ->when($status > 0, function ($q) use ($status) {
                        $q->where('app.status', '=', $status == 1 ? 1 : 0);
                    })
                    ->group('app.app_id');
            }])
            ->leftJoin('shop_user user', "user.app_id = app.app_id")
            ->leftJoin('supplier su', "su.app_id = app.app_id")
            ->leftJoin(
                "
                (
                    SELECT su.parent_id,COUNT(*) as count
                    FROM {$prefix}supplier as su
                    left join {$prefix}app as app on su.app_id = app.app_id
                    $countWhere
                    GROUP BY su.parent_id
                ) as children_apps",
                "su.shop_supplier_id = children_apps.parent_id"
            )
            ->where('user.is_super', '=', 1)
            ->where('app.is_delete', '=', 0)
            ->where('su.parent_id', '=', 0)
            ->when($keyword, function ($q) use ($keyword) {
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('su.name', $keyword);
                    $qq->orLike('su.app_id', $keyword);
                    $qq->whereOr('children_apps.count', '>', 0);
                });
            })
            ->when($status > 0, function ($q) use ($status) {
                $q->where(function ($q) use ($status) {
                    $q->where('app.status', '=', $status == 1 ? 1 : 0);
                    $q->whereOr('children_apps.count', '>', 0);
                });
            })
            ->when($shopType == 1, function ($q) {
                $q->where('children_apps.count', '>', 0);
            })
            ->when($shopType == 2, function ($q) {
                $q->whereNull('children_apps.count');
            })
            ->when($appId, function ($q) use ($appId) {
                $q->where('app.app_id', '=', $appId);
            })
            ->order(["app.create_time" => 'desc'])
            ->group('app.app_id')
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
        $level = $data['level'] ?? 1;
        $authDay = $data['auth_day'] ?? 0;
        if (!filter_var($authDay, FILTER_VALIDATE_INT) && $authDay != 0) {
            $this->error = '授权时间需为正整数';
            return false;
        }
        $data['auth_start_time'] = $authStartTime = strtotime($data['auth_start_time']) ?? 0;
        $data['expire_time'] = $authDay ? strtotime("+$authDay days", $authStartTime) : 0;
        $data['pay_type'] = json_encode(array_keys(OrderPayTypeEnum::pay()));
        // 验证不能大于父级时间
        if (isset($data['parent_id']) && $level == 2) {
            $parentSupplier = SupplierModel::find($data['parent_id']);
            $parentAuthDay = $parentSupplier?->app?->auth_day ?? 0;
            if ($parentAuthDay > 0 && ($authDay == 0 || $parentSupplier?->app?->expire_time < $data['expire_time'])) {
                $this->error = "不能大于总店的过期时间";
                return false;
            }
        }
        //
        $this->startTrans();
        try {
            // 添加应用
            $this->allowField(['auth_day', 'expire_time', 'app_name', 'auth_start_time'])->save($data);
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
        $shopSupplierId = $this->shop_supplier_id = $this->supplier?->shop_supplier_id;
        $appId = $this->app_id;
        //
        if (($data['level'] ?? 0) != $this->supplier->level) {
            $this->error = "不能编辑商家等级";
            return false;
        }
        $save_data = [
            'auth_day' => $authDay,
            'auth_start_time' => $authStartTime,
            'expire_time' => $authDay ? strtotime("+$authDay days", $authStartTime) : 0,
        ];
        //
        $level = $this->supplier->level;
        if (isset($data['parent_id']) && $level == 2) {
            // 验证上级不能等于自己
            if ($data['parent_id'] == $shopSupplierId) {
                $this->error = "上级不能等于自己";
                return false;
            }
            // 验证不能大于父级时间
            $parentSupplier = SupplierModel::find($data['parent_id']);
            $parentAuthDay = $parentSupplier?->app?->auth_day ?? 0;
            if ($parentAuthDay > 0 && ($authDay == 0 || $parentSupplier?->app?->expire_time < $save_data['expire_time'])) {
                $this->error = "不能大于总店的过期时间";
                return false;
            }
        } else if ($level == 1 && $authDay > 0) {
            // 验证不能小于子级时间
            $maxExpireTime = SupplierModel::alias('s')
                ->leftJoin('app', 'app.app_id = s.app_id')
                ->where('s.is_delete', 0)
                ->where('s.parent_id', $shopSupplierId)
                ->value('max(app.expire_time)');
            if ($authDay > 0 && $save_data['expire_time'] < $maxExpireTime) {
                $this->error = "不能小于分店的过期时间";
                return false;
            }
        }
        //
        $this->startTrans();
        try {
            // 用户
            $user_data = [
                'user_name' => $data['user_name'],
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
            $shop_user = ShopUser::withoutGlobalScope()->where('app_id', '=', $this['app_id'])
                ->where('is_delete', '=', 0)
                ->where('is_super', '=', 1)
                ->where('user_type', '=', 0)
                ->find();
            //
            if ($shop_user['phone'] != $data['link_phone']) {
                if (ShopUser::checkPhoneExist($data['link_phone'])) {
                    $this->error = '联系电话已存在';
                    return false;
                }
            }
            if ($shop_user['user_name'] != $data['user_name']) {
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
            $data['is_main'] = ($data['parent_id'] ?? 0) > 0 ? 0 : 1;
            SupplierModel::where('app_id', '=', $this['app_id'])->find()?->save($data);

            // 会员设置 - 支付方式关联
            if (isset($data['is_open_member'])) {
                (new PayType([], $appId))->setStatus(PayType::BALANCE_VALUE, $data['is_open_member']);
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
     * 移入移出回收站
     */
    public function recycle($is_recycle = true)
    {
        return $this->save(['is_recycle' => (int)$is_recycle]);
    }

    /**
     * 软删除
     */
    public function setDelete()
    {
        $this->supplier?->save(['is_delete' => 1]);
        return $this->save(['is_delete' => 1]);
    }
}