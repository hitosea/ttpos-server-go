<?php

use think\migration\Migrator;
use think\facade\Db;

class FixGrabOrderStateFollowRider extends Migrator
{
    /**
     * 修复 Grab 旧订单状态：未备餐完成时订单也跟随骑手状态变动
     *
     * 任务 #39635: 收银机接单-未备餐完成时订单也跟随状态变动（Grab相关）
     *
     * 处理逻辑：
     * 1. 对于 order_state = 10（已接单配餐中/备餐中）的订单
     *    - 如果 platform_order_state 已是骑手相关状态，视作已完成备餐
     *    - 按 Grab 骑手状态映射到 待配送(20)/配送中(30)/已完成(40)
     * 2. 订单时间超过 1 天的，直接处理为已完成(40)
     * 3. completed_time 为空（0）的，设置为 accepted_time
     *
     * Grab 平台状态映射：
     * - DRIVER_ARRIVED, DRIVER_ALLOCATED → 20 (待配送/待骑手接单)
     * - COLLECTED → 30 (配送中)
     * - DELIVERED, COMPLETED → 40 (已完成)
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);
        $currentTime = time();
        $oneDayAgo = $currentTime - 86400; // 1天前的时间戳

        // 1. 骑手已接单/到店 → 待配送 (order_state = 20)
        $db->name('takeout_order')
            ->where('platform', 'grab')
            ->where('order_state', 10)
            ->whereRaw("UPPER(platform_order_state) IN ('DRIVER_ARRIVED', 'DRIVER_ALLOCATED')")
            ->where('delete_time', 0)
            ->update([
                'order_state' => 20,
                'update_time' => $currentTime,
            ]);

        // 2. 骑手正在配送 → 配送中 (order_state = 30)
        $db->name('takeout_order')
            ->where('platform', 'grab')
            ->where('order_state', 10)
            ->whereRaw("UPPER(platform_order_state) = 'COLLECTED'")
            ->where('delete_time', 0)
            ->update([
                'order_state' => 30,
                'update_time' => $currentTime,
            ]);

        // 3. 骑手已配送完成 → 已完成 (order_state = 40)
        $db->name('takeout_order')
            ->where('platform', 'grab')
            ->where('order_state', 10)
            ->whereRaw("UPPER(platform_order_state) IN ('DELIVERED', 'COMPLETED')")
            ->where('delete_time', 0)
            ->update([
                'order_state' => 40,
                'update_time' => $currentTime,
            ]);

        // 4. 订单时间超过1天且仍处于备餐中(10)的订单 → 已完成 (order_state = 40)
        $db->name('takeout_order')
            ->where('platform', 'grab')
            ->where('order_state', 10)
            ->where('order_time', '<=', $oneDayAgo)
            ->where('order_time', '>', 0)
            ->where('delete_time', 0)
            ->update([
                'order_state' => 40,
                'update_time' => $currentTime,
            ]);

        // 5. completed_time 为空（0）的已完成订单，设置为 accepted_time
        $db->name('takeout_order')
            ->where('platform', 'grab')
            ->where('order_state', 40)
            ->where('completed_time', 0)
            ->where('accepted_time', '>', 0)
            ->where('delete_time', 0)
            ->update([
                'completed_time' => Db::raw('accepted_time'),
                'update_time' => $currentTime,
            ]);
    }
}
