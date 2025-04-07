<?php


namespace app\common\model\store;

use app\common\model\BaseModel;
use app\common\model\order\Order;
use app\common\enum\order\OrderStatusEnum;
use app\common\service\websocket\Websocket;

/**
 * 桌位模型
 */
class Table extends BaseModel
{
    protected $name = 'desk';
    protected $pk = 'id';

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['table_id', 'table_no', 'switch_status', 'is_bind'];

    /**
     * 分类更新后推送通知
     */
    public static function onAfterWrite(Table $model)
    {
        $msgData = [
            'type' => 'update',
            'desk_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_DESK, 0, $msgData);
    }

    /**
     * 分类删除后推送通知
     */
    public static function onAfterDelete(Table $model)
    {
        $msgData = [
            'type' => 'delete',
            'desk_uuid' => $model->uuid,
            'update_time' => time()
        ];
        Websocket::pushClient(request()->appId, Websocket::SOURCE_All, Websocket::SOURCE_All, Websocket::UPDATE_DESK, 0, $msgData);
    }

    /**
     * 兼容字段
     */
    public function getTableIdAttr($value, $data)
    {
        return $this->uuid ?: 0;
    }
    public function getTableNoAttr($value, $data)
    {
        return $this->desk_no ?: '';
    }
    public function getSwitchStatusAttr($value, $data)
    {
        return $this->is_disable == 1 ? 0 : 1;
    }
    public function getIsBindAttr($value, $data)
    {
        return $this->device_uuid > 0 ? 1 : 0;
    }
    public function getStatusAttr($value, $data)
    {
        return $this->sale_bill_uuid > 0 ? 30 : $value;
    }

    /**
     * 关联进行中订单
     */
    public function underwayOrder()
    {
        return $this->hasOne('app\\common\\model\\order\\Order', 'desk_uuid', 'uuid')->where('status', 0)->order('id desc');
    }

    /**
     * 桌位详情
     */
    public static function detail($where)
    {
        $filter = is_array($where) ? $where : ['uuid' => $where];
        return static::with(['underwayOrder'])->where($filter)->find();
    }

    /**
     * 获取桌台分类
     * @param $shop_supplier_id
     * @return array
     */
    public static function getAreaType($shop_supplier_id)
    {
        $areaList = (new TableArea)->select();
        $typeList = (new TableType)->select();
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
        return self::where('uuid', '=', $table_id)->update(['status' => 0]);
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
            'desk_no' => $name,
        ];
        if (!is_null($id) && $id != 0) {
            $filter[] = ['uuid', '<>', $id];
        }
        return static::where($filter)->value('id') ? true : false;
    }
}
