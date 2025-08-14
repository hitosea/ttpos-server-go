<?php

namespace app\common\model\pay;

use help\HttpHelp;
use think\facade\Log;
use help\RSAEncryptorHelp;
use app\common\model\BaseModel;
use app\common\model\store\PayType;
use app\common\model\app\App;

/**
 * 支付应用模型
 */
class PaymentApp extends BaseModel
{
    protected $name = 'payment_app';
    protected $pk = 'id';

    /**
     * 追加字段
     */
    protected $append = [];

    /**
     * 详情
     */
    public static function detail($shop_supplier_id)
    {
        return self::where('company_uuid', '=', $shop_supplier_id)->find() ?: ['ll_white_ip' => env('PAY_SERVICE_IP') ?: ''];
    }

    /**
     * 支付
     * @param mixed $param
     * @return bool
     */
    public function payment($param)
    {
        // 获取支付服务环境配置
        $ip = env('PAY_SERVICE_IP');
        $url = env('PAY_SERVICE_URL') . '/api/platform/add';
        $pbk = env('PAY_SERVICE_RSA_PUBLIC_KEY');

        // 检查环境配置
        if (!$ip || !$url || !$pbk) {
            $this->error = '支付服务环境配置错误';
            return false;
        }

        // 加密数据
        try {
            $encryptedData = $this->encryptData($param, $pbk, $ip);
            if (!$encryptedData) {
                $this->error = '数据加密失败(01): 检查 PAY_SERVICE_RSA_PUBLIC_KEY 配置是否正确';
                return false;
            }
        } catch (\Exception $e) {
            $this->error = '数据加密失败(02): 检查 PAY_SERVICE_RSA_PUBLIC_KEY 配置是否正确';
            return false;
        }
        // 发送支付请求
        $response = $this->sendPaymentRequest($url, $encryptedData);
        if (!$response || $response['code'] != 1) {
            $this->error = isset($response['msg']) ? __($response['msg']) . json_encode($response['data']) : '支付服务请求失败';
            return false;
        }


        $param['ll_sign_salt'] = $response['data']['sign_salt'] ?? '';
        // 更新或新增记录
        return $this->savePaymentData($param, $ip);
    }

    /**
     * 加密数据
     * @param array $param
     * @param string $publicKey
     * @param string $ip
     * @return string|false
     */
    private function encryptData($param, $publicKey, $ip)
    {
        $encryptor = new RSAEncryptorHelp();
        $publicKeyPem = $encryptor->toPemPublicKey($publicKey);

        $data = [
            'll_white_ip' => $ip,
            'll_merchant_id' => $param['ll_merchant_id'] ?? '',
            'll_public_key' => $param['ll_public_key'] ?? '',
            'll_merchant_private_key' => $param['ll_merchant_private_key'] ?? '',
            'll_token' => $param['ll_token'] ?? '',
            'll_store_id' => $param['ll_store_id'] ?? '',
            'shop_supplier_id' => $param['shop_supplier_id'] ?? 0,
        ];

        return $encryptor->pubEncrypt(json_encode($data), $publicKeyPem);
    }

    /**
     * 发送支付请求
     * @param string $url
     * @param string $encryptedData
     * @return array|false
     */
    private function sendPaymentRequest($url, $encryptedData)
    {
        $post = ['encrypt_data' => $encryptedData];
        $response = HttpHelp::postRequest($url, $post);
        return json_decode($response, true);
    }

    /**
     * 保存支付数据
     * @param array $param
     * @param string $ip
     * @return bool
     */
    private function savePaymentData($param, $ip)
    {
        $model = new self();
        $param['ll_white_ip'] = $ip;
        $param['company_uuid'] = $param['shop_supplier_id'] ?? 0;

        // 开启事务
        $model->startTrans();
        try {
            $existingRecord = $model->where('company_uuid', $param['company_uuid'])->find();
            // 确保 create_time 是时间戳
            $create_time = $existingRecord['create_time'] ?? time();
            if (!is_numeric($create_time)) {
                $create_time = strtotime($create_time);
            }
            
            if ($existingRecord) {
                if (!$existingRecord->save($param)) {
                    $model->rollback();
                    $this->error = '更新失败';
                    return false;
                }
            } else {
                if (!$model->save($param)) {
                    $model->rollback();
                    $this->error = '新增失败';
                    return false;
                }
            }
            
            // 更新lianlianpay支付方式时间
            $app_id = $param['app_id'] ?? 0;
            if ($app_id) {
                $result = (new PayType([], $app_id))->where('source', 2)->update(['create_time' => $create_time, 'update_time' => $create_time]);
                if ($result === false) {
                    $model->rollback();
                    $this->error = '更新支付方式时间失败';
                    return false;
                }
            }

            // 调用erpnext支付方式添加接口
            $company = (new App())->where('uuid', $param['company_uuid'])->find();
            if ($company->is_enable_erp) {
                $res = HttpHelp::postRequest('http://nginx/api/v1/admin/erpnext/lianlian/payment/add', json_encode($param), [
                    'X-API-KEY: ' . env('JWT_SECRET'),
                    'Accept-Language: ' . request()->header('language'),
                ]);
                if (!$res) {
                    Log::error('调用erpnext支付方式添加接口失败', $res);
                    $this->error = '调用erpnext支付方式添加接口失败';
                    return false;
                }
                $res = json_decode($res, true);
                if ($res['code'] != 0) {
                    $this->error = $res['message'];
                    return false;
                }
            }
            

            // 提交事务
            $model->commit();
        } catch (\Exception $e) {
            // 回滚事务
            $model->rollback();
            $this->error = '操作失败：' . $e->getMessage();
            return false;
        }

        return true;
    }
}
