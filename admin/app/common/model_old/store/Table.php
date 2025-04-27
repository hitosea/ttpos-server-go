<?php


namespace app\common\model_old\store;

use app\common\model_old\BaseModel;
use app\common\model_old\order\Order;
use app\common\enum\settings\SettingEnum;
use app\common\enum\order\OrderStatusEnum;
use app\tablet\model\order\Order as OrderModel;
use app\common\model_old\settings\Setting as SettingModel;

/**
 * 桌位模型
 */
class Table extends BaseModel
{
    protected $pk = 'table_id';
    protected $name = 'table';

    /**
     * 关联门店
     */
    public function supplier()
    {
        return $this->BelongsTo('app\\common\\model\\supplier\\Supplier', 'shop_supplier_id', 'shop_supplier_id');
    }

    /**
     * 关联进行中订单
     */
    public function underwayOrder()
    {
        return $this->hasOne('app\\common\\model\\order\\Order', 'table_id', 'table_id')->where('is_merge', 0)->where('order_status', 10)->order('order_id desc');
    }

    /**
     * 桌位详情
     */
    public static function detail($where)
    {
        $filter = is_array($where) ? $where : ['table_id' => $where];
        return static::with(['underwayOrder'])->where($filter)->find();
    }

    /**
     * 查询桌台是否进行中
     * @param $table_id
     * @return bool
     */
    public static function isOpen($table_id)
    {
        $order = OrderModel::where([['table_id', '=', $table_id], ['order_status', '=', OrderStatusEnum::NORMAL]])->field('table_id')->find();
        $tableInfo = static::field('status, shop_supplier_id, app_id')->where('table_id', '=', $table_id)->find();
        $tableStatus = $tableInfo['status'];
        if (!$order) {
            if ($tableStatus == 30) {
                // 支付后是否清台 0-清台 1-不清台
                $store = SettingModel::getSupplierItem(SettingEnum::BUSINESS, $tableInfo['shop_supplier_id'], $tableInfo['app_id']);
                if ($store['no_clear_table'] == 1) {
                    return true;
                }
                static::where('table_id', '=', $table_id)->update(['status' => 10]);
            }
            return false;
        }
        return $tableStatus == 30;
    }

    /**
     * 查询桌台是否绑定中 1:绑定  0:未绑定
     * @param $table_id
     * @return bool
     */
    public static function isBind($table_id)
    {
        return static::where('table_id', '=', $table_id)->value('is_bind') == 1;
    }

    /**
     * 获取桌台分类
     * @param $shop_supplier_id
     * @return array
     */
    public static function getAreaType($shop_supplier_id)
    {
        $areaList = (new TableArea)->where('shop_supplier_id', '=', $shop_supplier_id)->select();
        $typeList = (new TableType)->where('shop_supplier_id', '=', $shop_supplier_id)->select();
        return compact('areaList', 'typeList');
    }

    /**
     * 获取已开台桌台
     * @param $table_id     // 当前桌台ID
     * @param $merge_id     // 订单合并ID
     * @return array
     */
    public static function getTablesMergeList($table_id, $merge_id)
    {
        $mergeTableIds = $merge_id ? Order::where('merge_id', $merge_id)->column('table_id') : [];
        //
        $openTableList =  (new Order)->alias('o')
            ->field([
                't.table_id',
                't.table_no',
                't.status',
                't.sort',
                'o.order_id',
                'o.merge_id',
                'o.order_status',
                'o.user_id',
                'o.is_change_price',
            ])
            ->join('table t', 't.table_id = o.table_id')
            ->with(['userBaseInfo'])
            ->where('t.status', '=', 30)
            ->where('o.order_status', '=', OrderStatusEnum::NORMAL)
            ->where(function ($q) use ($merge_id) {
                $q->where('o.merge_id', '')->whereOr('o.merge_id', $merge_id);
            })
            ->group('t.table_id')
            ->order('o.create_time', 'desc')
            ->select()->toArray();

        foreach ($openTableList as $k => &$table) {
            if (in_array($table['table_id'], $mergeTableIds)) {
                $table['is_merge'] = 1;
            } else {
                $table['is_merge'] = 0;
            }
            // 当前桌
            if ($table['table_id'] == $table_id) {
                $table['is_current_table'] = 1;
            } else {
                $table['is_current_table'] = 0;
            }
        }

        // 移除多余键
        $keysToRemove = ["state_text", "order_source_text", "order_type_text", "deliver_text", "elapsed_time", "pay_time_text", "buffet_remaining_time", "actual_price", "actual_receive_price"];
        $openTableList = array_map(function ($item) use ($keysToRemove) {
            return array_filter($item, function ($key) use ($keysToRemove) {
                return !in_array($key, $keysToRemove);
            }, ARRAY_FILTER_USE_KEY);
        }, $openTableList);

        //
        $isMergeColumn = array_column($openTableList, 'is_merge');  // 已合并的优先显示
        $sortColumn = array_column($openTableList, 'sort'); // 按sort的升序
        // 使用array_multisort进行多字段排序
        array_multisort($isMergeColumn, SORT_DESC, $sortColumn, SORT_ASC, $openTableList);

        return $openTableList;
    }


    /**
     * 获取已开台桌台  1.0.9
     * @param $table_id     // 当前桌台ID
     * @return array
     */
    public static function getTableMergeList($table_id)
    {
        //
        $openTableList =  (new Order)->alias('o')
            ->field([
                't.table_id',
                't.table_no',
                't.status',
                't.sort',
                'o.order_id',
                'o.merge_id',
                'o.order_status',
                'o.user_id',
                'o.is_change_price',
            ])
            ->join('table t', 't.table_id = o.table_id')
            ->with(['userBaseInfo'])
            ->where('t.status', '=', 30)
            ->where('o.order_status', '=', OrderStatusEnum::NORMAL)
            ->where('o.is_buffet', '=', 0)
            ->group('t.table_id')
            ->order('o.create_time', 'desc')
            ->select()->toArray();

        foreach ($openTableList as $k => &$table) {
            // 当前桌
            if ($table['table_id'] == $table_id) {
                $table['is_current_table'] = 1;
            } else {
                $table['is_current_table'] = 0;
            }
        }

        // 移除多余键
        $keysToRemove = ["state_text", "order_source_text", "order_type_text", "deliver_text", "elapsed_time", "pay_time_text", "buffet_remaining_time", "actual_price", "actual_receive_price"];
        $openTableList = array_map(function ($item) use ($keysToRemove) {
            return array_filter($item, function ($key) use ($keysToRemove) {
                return !in_array($key, $keysToRemove);
            }, ARRAY_FILTER_USE_KEY);
        }, $openTableList);

        //
        $sortColumn = array_column($openTableList, 'sort'); // 按sort的升序
        // 使用array_multisort进行多字段排序
        array_multisort($sortColumn, SORT_ASC, $openTableList);

        return $openTableList;
    }

    /**
     * 开台
     * @param $table_id
     * @return Table
     */
    public static function open($table_id)
    {
        return self::where('table_id', '=', $table_id)->update(['status' => 30]);
    }

    /**
     * 关台
     * @param $table_id
     * @return Table
     */
    public static function close($table_id)
    {
        return self::where('table_id', '=', $table_id)->update(['status' => 10]);
    }

    /**
     * 获取已开台桌台
     * @param $table_id     // 当前桌台ID
     * @return array
     */
    public static function getMoveTablesList($table_id)
    {
        $openTableList = (new Order)->alias('o')
            ->field([
                't.table_id',
                't.table_no',
                't.status',
                't.sort',
                'o.order_id',
                'o.merge_id',
                'o.order_status',
                'o.user_id',
                'o.is_change_price',
            ])
            ->join('table t', 't.table_id = o.table_id')
            ->with(['userBaseInfo'])
            ->where('t.status', '=', 30)
            ->where('t.table_id', '<>', $table_id)
            ->where('o.order_status', '=', OrderStatusEnum::NORMAL)
            ->group('t.table_id')
            ->order('o.create_time', 'desc')
            ->select()->toArray();

        // 移除多余键
        $keysToRemove = ["state_text", "order_source_text", "order_type_text", "deliver_text", "elapsed_time", "pay_time_text", "buffet_remaining_time", "actual_price", "actual_receive_price"];
        $openTableList = array_map(function ($item) use ($keysToRemove) {
            return array_filter($item, function ($key) use ($keysToRemove) {
                return !in_array($key, $keysToRemove);
            }, ARRAY_FILTER_USE_KEY);
        }, $openTableList);
        //
        $sortColumn = array_column($openTableList, 'sort'); // 按sort的升序
        // 使用array_multisort进行多字段排序
        array_multisort($sortColumn, SORT_ASC, $openTableList);

        return $openTableList;
    }

    /**
     * 检查名称唯一性
     */
    public function checkNameExist($name, $shop_supplier_id, $id = null)
    {
        $filter = [
            'table_no' => $name,
            'shop_supplier_id' => $shop_supplier_id
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['table_id', '<>', $id];
        }
        return static::where($filter)->value('table_id') ? true : false;
    }
}
