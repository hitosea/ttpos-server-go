<?php

namespace app\common\service\websocket;

class Websocket
{
    // 源类型常量
    const SOURCE_All = '*';                 // 全部
    const SOURCE_SHOP = 'shop';             // 商家
    const SOURCE_CASHIER = 'cashier';       // 收银机
    const SOURCE_TABLET = 'tablet';         // 平板端
    const SOURCE_KITCHEN = 'kitchen';       // 厨显端
    const SOURCE_ASSISTANT = 'assistant';   // 点餐助手
    const SOURCE_H5 = 'H5';                 // H5

    // 消息类型常量
    const UPDATE_ORDER = 'update_order';                            // 更新订单
    const CUSTOMER_CALL = 'customer_call';                          // 客户呼叫
    const PRINT_DATA = 'print_data';                                // 打印数据
    const H5_ORDER = 'h5_order';                                    // H5订单
    const UPDATE_CONFIG = 'update_config';                          // 更新配置
    const UPDATE_PERMISSION = 'update_permission';                  // 更新权限
    const UPDATE_USER = 'update_user';                              // 更新用户
    const UPDATE_PRODUCT = 'update_product';                        // 更新商品
    const UPDATE_CATEGORY = 'update_category';                      // 更新分类
    const UPDATE_SELECTED_PRINTER = 'update_selected_printer';      // 更新打印机
    const UPDATE_BUFFET = 'update_buffet';                          // 更新自助餐
    const UPDATE_DESK = 'update_desk';                              // 更新桌台
    const UPDATE_DESK_TYPE = 'update_desk_type';                    // 更新桌台类型

    /**
     * 推送消息到WebSocket服务器
     *
     * @param int $companyUuid 公司UUID
     * @param string $sourceClient 源客户端
     * @param string $notDeviceId 排除的设备ID
     * @param string $messageType 消息类型
     * @param array $data 消息数据
     * @return bool
     */
    public static function pushClient(
        int $companyUuid, 
        string $sourceClient, 
        string $notDeviceId, 
        string $messageType, 
        int $staffUuid, 
        array $data
    ) {

        try {
            // 计算包含关键参数的MD5值
            $jsonData = '';
            if ($messageType === self::UPDATE_ORDER && isset($data['sale_bill_uuid'])) {
                $jsonData = (string)$data['sale_bill_uuid'];
            }

            // 创建缓存键
            $cacheKey = '';
            if ($messageType !== self::PRINT_DATA) {
                $key = sprintf("%d:%s:%s:%s", $companyUuid, $sourceClient, $notDeviceId, $messageType);
                $md5Sum = md5($key . $jsonData);
                $cacheKey = sprintf("ws_msg:%s", $md5Sum);
            }

            // 判断当前是否在容器内执行
            $url = sprintf("http://127.0.0.1:%s/ws/push", getenv('NGINX_PORT'));
            if (file_exists('/.dockerenv')) {
                $url = "http://nginx/ws/push";
            }

            // 构建请求体
            $payload = [
                'company_uuid' => $companyUuid,
                'source_client' => $sourceClient,
                'staff_uuid' => $staffUuid,
                'device_id' => "*",
                'not_device_id' => $notDeviceId,
                'message_type' => $messageType,
                'message_key' => $cacheKey,
                'data' => $data,
            ];

            // 发送POST请求
            $ch = curl_init($url);
            curl_setopt_array($ch, [
                CURLOPT_POST => true,
                CURLOPT_POSTFIELDS => json_encode($payload),
                CURLOPT_RETURNTRANSFER => true,
                CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
            ]);

            $response = curl_exec($ch);
            $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            curl_close($ch);

            if ($httpCode !== 200) {
                error_log(sprintf("Received non-OK response: %d %s", $httpCode, $url));
                return false;
            }

            return true;
        } catch (\Exception $e) {
            error_log(sprintf("Failed to push websocket message: %s", $e->getMessage()));
            return false;
        }
    }


    /**
     * 推送消息到指定设备
     *
     * @param int $companyUuid 公司UUID
     * @param string $sourceClient 源客户端
     * @param string $deviceId 设备ID
     * @param string $messageType 消息类型
     * @param array $data 消息数据
     * @return bool
     */
    public static function pushAppointClient(
        int $companyUuid, 
        string $sourceClient, 
        string $deviceId, 
        string $messageType, 
        int $staffUuid, 
        array $data
    ) {

        try {
            // 计算包含关键参数的MD5值
            $jsonData = '';
            if ($messageType === self::UPDATE_ORDER && isset($data['sale_bill_uuid'])) {
                $jsonData = (string)$data['sale_bill_uuid'];
            }

            // 创建缓存键
            $cacheKey = '';
            if ($messageType !== self::PRINT_DATA) {
                $key = sprintf("%d:%s:%s:%s", $companyUuid, $sourceClient, $deviceId, $messageType);
                $md5Sum = md5($key . $jsonData);
                $cacheKey = sprintf("ws_msg:%s", $md5Sum);
            }

            // 判断当前是否在容器内执行
            $url = sprintf("http://127.0.0.1:%s/ws/push", getenv('NGINX_PORT'));
            if (file_exists('/.dockerenv')) {
                $url = "http://nginx/ws/push";
            }

            // 构建请求体
            $payload = [
                'company_uuid' => $companyUuid,
                'source_client' => $sourceClient,
                'staff_uuid' => $staffUuid,
                'device_id' => $deviceId,
                'message_type' => $messageType,
                'message_key' => $cacheKey,
                'data' => $data,
            ];

            // 发送POST请求
            $ch = curl_init($url);
            curl_setopt_array($ch, [
                CURLOPT_POST => true,
                CURLOPT_POSTFIELDS => json_encode($payload),
                CURLOPT_RETURNTRANSFER => true,
                CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
            ]);

            $response = curl_exec($ch);
            $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
            curl_close($ch);

            if ($httpCode !== 200) {
                error_log(sprintf("Received non-OK response: %d %s", $httpCode, $url));
                return false;
            }

            return true;
        } catch (\Exception $e) {
            error_log(sprintf("Failed to push websocket message: %s", $e->getMessage()));
            return false;
        }
    }
}
