<?php

namespace app\shop\model\buffet;

use help\ValidateHelp;
use app\common\model\buffet\BuffetCustomer;
use app\common\model\store\MultiLanguageName;
use app\common\model\buffet\Buffet as BuffetModel;

/**
 * 自助餐模型
 */
class Buffet extends BuffetModel
{
    /**
     * 获取自助餐详情
     */
    public static function detail($buffet_id)
    {
        return self::with([
            'buffetProducts', 
            'buffetLimitProducts', 
            'buffetCustomerType', 
            'buffetTaxes',
            'saleOrderBuffetCustomerType' => [
                'saleOrder' => [
                    'saleBill'
                ]
            ],
            'multiLanguageName'
        ])->where('uuid', $buffet_id)->find();
    }

    /**
     * 获取自助餐列表
     */
    public function getShopBuffetList($params)
    {
        $model = $this;
        // 名称
        if (isset($params['name']) && $params['name'] != '') {
            $model = $model->jsonLike('name', $params['name']);
        }
        // 状态
        if (isset($params['status']) && $params['status'] > -1) {
            $model = $model->where('status', (int)$params['status']);
        }
        // 查询列表数据
        $list = $model->with([
                'buffetProducts' => ['product'], 
                'buffetLimitProducts', 
                'buffetCustomerType', 
                'buffetTaxes',
                'saleOrderBuffetCustomerType' => [
                    'saleOrder' => [
                        'saleBill'
                    ]
                ]
            ])
            ->order(['create_time' => 'desc'])
            ->paginate($params);
        foreach ($list as &$item) {
            $item['buy_limit_status'] = count($item['buffetLimitProducts']) > 0 ? 1 : 0;
            $item['can_delete'] = $this->getCanDelete($item);
            if (count($item['buffetTaxes']) > 0 && isset($item['buffetTaxes'][0])) {
                $item['buffetTaxes'][0]['buffet_tax_type'] = '1';
            }
        }
        return $list;
    }

    /**
     * 添加自助餐
     */
    public function add($data)
    {
        $data['name'] = $data['name'] ?? ''; // 自助餐名称
        $data['sort'] = $data['sort'] ?? 0; // 排序
        $data['price'] = $data['price'] ?? 0; // 价格
        $data['time_limit'] = $data['time_limit'] ?? 90; // 用餐时间
        $data['is_time_limit'] = $data['is_time_limit'] ?? 0; // 是否开启不限制时间 0-否 1-是
        $data['time_limit'] = $data['is_time_limit'] == 1 ? $data['time_limit'] : 0;
        $data['status'] = $data['status'] ?? 0; // 状态 0-未开启 1-已开启
        $data['is_comb'] = $data['is_comb'] ?? 0; // 是否组合 0-否 1-是
        $data['buy_limit_status'] = $data['buy_limit_status'] ?? 0; // 是否限购 0-否 1-是
        $data['is_remain_continue'] = $data['is_remain_continue'] ?? 0; // 平板是否可继续点餐开关 0-关闭 1-开启
        $data['remain_continue_time'] = $data['remain_continue_time'] ?? 20; // 剩余xx分不可继续点餐
        $data['remain_continue_notice_time'] = $data['remain_continue_notice_time'] ?? 5; // 剩余xx分提醒不可继续点餐
        $data['open_overall_discount'] = $data['open_overall_discount'] ?? 1; // 是否开启整单折扣 0-否 1-是
        $shop_supplier_id = $data['shop_supplier_id'] ?? 0; // 供应商id

        if (ValidateHelp::hasEmptyValue($data['name'])) {
            $this->error = '请输入自助餐名称';
            return false;
        }

        // 如果name为数组，则转为json_encode保存
        if (is_array($data['name'])) {
            $data['name'] = json_encode($data['name'], JSON_UNESCAPED_UNICODE);
        }

        // 验证平板时间参数
        if ($data['is_remain_continue'] == 1) {
            if ($data['remain_continue_time'] < 1 || $data['remain_continue_time'] > 999) {
                $this->error = '剩余不可继续点餐时间必须在1到999之间';
                return false;
            }
            if ($data['remain_continue_notice_time'] < 1 || $data['remain_continue_notice_time'] > 999) {
                $this->error = '剩余不可继续点餐提醒时间必须在1到999之间';
                return false;
            }
        } else {
            $data['remain_continue_time'] = 0;
            $data['remain_continue_notice_time'] = 0;
        }

        // 验证顾客类型数据
        $customerType = $data['customer_type'] ?? [];
        if (is_array($customerType)) {
            if (empty($customerType)) {
                $this->error = '请选择顾客类型';
                return false;
            }
            if (count($customerType) > 5) {
                $this->error = '顾客类型不能超过5个';
                return false;
            }
            // 判断customer_type_id有值相同时返回提醒
            $customerTypeIds = array_column($customerType, 'customer_type_id');
            if (count(array_unique($customerTypeIds)) < count($customerTypeIds)) {
                $this->error = '顾客类型不可重复选择';
                return false;
            }
            $newCustomerType = [];
            foreach ($customerType as &$customer) {
                // 处理负数
                $customer = $this->sanitizeFormData(['price'], $customer);
                // 处理范围值
                if ($text = $this->alertFormData($customer)) {
                    $this->error = $text;
                    return false;
                }
                $customer['shop_supplier_id'] = $shop_supplier_id;
                $newCustomerType[$customer['customer_type_id']] = $customer;
            }
            unset($customer);
            $customerType = $newCustomerType;
        }

        // 如果开启组合产品，则必须填写组合产品
        if (empty($data['products'])) {
            $this->error = '请选择组合产品';
            return false;
        }
        $products = $data['products'] ? $data['products'] : [];

        // 如果开启限购，则必须填写限购数量
        if ($data['buy_limit_status'] == 1 && ($data['buy_limit_products'] ?? []) == []) {
            $this->error = '请输入限购数量';
            return false;
        }
        $buyLimitProducts = $data['buy_limit_status'] == 1 ? $data['buy_limit_products'] : [];
        // 处理产品限购数据
        if (is_array($buyLimitProducts)) {
            $newBuyLimitProducts = [];
            foreach ($buyLimitProducts as &$limitProduct) {
                // 处理负数
                $limitProduct = $this->sanitizeFormData(['limit_num'], $limitProduct);
                // 处理范围值
                if ($text = $this->alertFormData($limitProduct)) {
                    $this->error = $text;
                    return false;
                }
                $newBuyLimitProducts[$limitProduct['product_id']] = $limitProduct;
            }
            unset($limitProduct);
            $buyLimitProducts = $newBuyLimitProducts;
        }

        // 验证主数据
        if (is_array($data)) {
            // 处理负数
            $data = $this->sanitizeFormData(['price', 'time_limit', 'sort'], $data);
            // 处理范围值
            if ($text = $this->alertFormData($data)) {
                $this->error = $text;
                return false;
            }
        }
        //
        if (is_array($data['buffetTaxes']) && count($data['buffetTaxes']) > 1) {
            $this->error = '自助餐税类只能设置1条';
            return false;
        }
        //
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($data['name']);
        $data['tax_uuid'] = $data['buffetTaxes'][0]['tax_category_id'] ?? 0; // 税收ID
        $data['is_limit_time'] = $data['is_time_limit']; // 是否限时, 0-否 1-是
        $data['limit_time'] = $data['time_limit']; // 限时时间(分钟)
        $data['can_combined'] = $data['is_comb']; // 是否可合并, 0-否 1-是
        $data['non_ordering_time'] = $data['remain_continue_time']; // 平板不可下单时间(分钟)
        $data['reminder_order_time'] = $data['remain_continue_notice_time']; // 平板提醒不可下单时间(分钟)
        // 开启事务
        $this->startTrans();
        try {
            $this->save($data);
            // 更新自助餐顾客类型
            $this->updateBuffetCustomer($this['uuid'], $customerType);
            // 更新自助餐关联产品
            $this->updateBuffetProduct($products, $buyLimitProducts);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 编辑自助餐
     */
    public function edit($data)
    {
        $data['name'] = $data['name'] ?? ''; // 自助餐名称
        $data['sort'] = $data['sort'] ?? 0; // 排序
        $data['price'] = $data['price'] ?? 0; // 价格
        $data['time_limit'] = $data['time_limit'] ?? 0; // 用餐时间
        $data['is_time_limit'] = $data['is_time_limit'] ?? 0; // 是否开启不限制时间 0-否 1-是
        $data['time_limit'] = $data['is_time_limit'] == 1 ? $data['time_limit'] : 0;
        $data['status'] = $data['status'] ?? 0; // 状态 0-未开启 1-已开启
        $data['is_comb'] = $data['is_comb'] ?? 0; // 是否组合 0-否 1-是
        $data['buy_limit_status'] = $data['buy_limit_status'] ?? 0; // 是否限购 0-否 1-是
        $data['is_remain_continue'] = $data['is_remain_continue'] ?? 0; // 平板是否可继续点餐开关 0-关闭 1-开启
        $data['remain_continue_time'] = $data['remain_continue_time'] ?? 0; // 剩余xx分不可继续点餐
        $data['remain_continue_notice_time'] = $data['remain_continue_notice_time'] ?? 0; // 剩余xx分提醒不可继续点餐
        $data['open_overall_discount'] = $data['open_overall_discount'] ?? 1; // 是否开启整单折扣 0-否 1-是

        if (ValidateHelp::hasEmptyValue($data['name'])) {
            $this->error = '请输入自助餐名称';
            return false;
        }
        // 如果name为数组，则转为json_encode保存
        if (is_array($data['name'])) {
            $data['name'] = json_encode($data['name'], JSON_UNESCAPED_UNICODE);
        }

        // 验证平板时间参数
        if ($data['is_remain_continue'] == 1) {
            if ($data['remain_continue_time'] < 1 || $data['remain_continue_time'] > 999) {
                $this->error = '剩余不可继续点餐时间必须在1到999之间';
                return false;
            }
            if ($data['remain_continue_notice_time'] < 1 || $data['remain_continue_notice_time'] > 999) {
                $this->error = '剩余不可继续点餐提醒时间必须在1到999之间';
                return false;
            }
        } else {
            $data['remain_continue_time'] = 0;
            $data['remain_continue_notice_time'] = 0;
        }

        // 验证顾客类型数据
        $customerType = $data['customer_type'] ?? [];
        if (is_array($customerType)) {
            if (empty($customerType)) {
                $this->error = '请选择顾客类型';
                return false;
            }
            if (count($customerType) > 5) {
                $this->error = '顾客类型不能超过5个';
                return false;
            }
            // 判断customer_type_id有值相同时返回提醒
            $customerTypeIds = array_column($customerType, 'customer_type_id');
            if (count(array_unique($customerTypeIds)) < count($customerTypeIds)) {
                $this->error = '顾客类型不可重复选择';
                return false;
            }
            $newCustomerType = [];
            foreach ($customerType as &$customer) {
                // 处理负数
                $customer = $this->sanitizeFormData(['price'], $customer);
                // 处理范围值
                if ($text = $this->alertFormData($customer)) {
                    $this->error = $text;
                    return false;
                }
                $customer['shop_supplier_id'] = $this->shop_supplier_id;
                $newCustomerType[$customer['customer_type_id']] = $customer;
            }
            unset($customer);
            $customerType = $newCustomerType;
        }

        // 如果开启组合产品，则必须填写组合产品
        if (empty($data['products'])) {
            $this->error = '请选择组合产品';
            return false;
        }
        $products = $data['products'] ? $data['products'] : [];

        // 如果开启限购，则必须填写限购数量
        if ($data['buy_limit_status'] == 1 && ($data['buy_limit_products'] ?? []) == []) {
            $this->error = '请输入限购数量';
            return false;
        }
        $buyLimitProducts = $data['buy_limit_status'] == 1 ? $data['buy_limit_products'] : [];
        // 处理产品限购数据
        if (is_array($buyLimitProducts)) {
            $newBuyLimitProducts = [];
            foreach ($buyLimitProducts as &$limitProduct) {
                // 处理负数
                $limitProduct = $this->sanitizeFormData(['limit_num'], $limitProduct);
                // 处理范围值
                if ($text = $this->alertFormData($limitProduct)) {
                    $this->error = $text;
                    return false;
                }
                $newBuyLimitProducts[$limitProduct['product_id']] = $limitProduct;
            }
            unset($limitProduct);
            $buyLimitProducts = $newBuyLimitProducts;
        }

        // 验证主数据
        if (is_array($data)) {
            // 处理负数
            $data = $this->sanitizeFormData(['price', 'time_limit', 'sort'], $data);
            // 处理范围值
            if ($text = $this->alertFormData($data)) {
                $this->error = $text;
                return false;
            }
        }
        //
        if (is_array($data['buffetTaxes']) && count($data['buffetTaxes']) > 1) {
            $this->error = '自助餐税类只能设置1条';
            return false;
        }
        //
        $data['multi_language_name_uuid'] = (new MultiLanguageName)->saveNames($data['name'], $this['multi_language_name_uuid']);
        $data['tax_uuid'] = $data['buffetTaxes'][0]['tax_category_id'] ?? 0; // 税收ID
        $data['is_limit_time'] = $data['is_time_limit']; // 是否限时, 0-否 1-是
        $data['limit_time'] = $data['time_limit']; // 限时时间(分钟)
        $data['can_combined'] = $data['is_comb']; // 是否可合并, 0-否 1-是
        $data['non_ordering_time'] = $data['remain_continue_time']; // 平板不可下单时间(分钟)
        $data['reminder_order_time'] = $data['remain_continue_notice_time']; // 平板提醒不可下单时间(分钟)
        // 开启事务
        $this->startTrans();
        try {
            unset($data['id']);
            $this->save($data);
            // 更新自助餐顾客类型
            $this->updateBuffetCustomer($this['uuid'], $customerType);
            // 更新自助餐关联产品
            $this->updateBuffetProduct($products, $buyLimitProducts);
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 更新自助餐关联产品
     */
    private function updateBuffetProduct($products, $buyLimitProducts)
    {
        $buffetProducts = new BuffetProduct;
        $buffetProducts->destroy(['buffet_package_uuid' => $this['uuid']]);
        // 新增关联产品
        $data = [];
        foreach ($products as $item) {
            $productId = $item['product_id'];
            $data[] = [
                'buffet_package_uuid' => $this['uuid'],
                'product_package_uuid' => $productId,
                'limit' => $buyLimitProducts[$productId]['limit_num'] ?? 0,
                'is_show_cashier' => $item['is_show_cashier'] ?? 0,
                'is_show_tablet' => $item['is_show_tablet'] ?? 0,
                'is_show_kitchen' => $item['is_show_kitchen'] ?? 0,
                'is_show_assistant' => $item['is_show_assistant'] ?? 0,
                'is_show_h5' => $item['is_show_h5'] ?? 0,
            ];
        }
        $buffetProducts->saveAll($data);
    }

    /**
     * 更新自助餐关联顾客类型
     */
    public function updateBuffetCustomer($buffet_id, $customers)
    {
        $delete = [];
        foreach ($customers as $item) {
            $customerTypeId = $item['customer_type_id'];
            $price = $item['price'] ?? 0;
            $buffetCustomerType = BuffetCustomer::where('buffet_package_uuid', $buffet_id)->where('customer_type_uuid', $customerTypeId)->find();
            if ($buffetCustomerType) {
                $buffetCustomerType->price = $price;
                $buffetCustomerType->save();
            } else {
                $buffetCustomerType = BuffetCustomer::create([
                    'buffet_package_uuid' => $buffet_id,
                    'customer_type_uuid' => $customerTypeId,
                    'price' => $price,
                ]);
            }
            $delete[] = $buffetCustomerType->uuid;
        }
        if (count($delete) > 0) {
            $list = BuffetCustomer::where('buffet_package_uuid', $buffet_id)->whereNotIn('uuid', $delete)->select();
            foreach ($list as $item) {
                $item->delete();
            }
        }
    }

    /**
     * 设置自助餐组合
     */
    public function setComb($is_comb)
    {
        return $this->save(['can_combined' => (int)$is_comb]);
    }

    /**
     * 设置自助餐状态
     */
    public function setStatus($state)
    {
        return $this->save(['status' => (int)$state]);
    }

    /**
     * 设置自助餐状态
     */
    public function setOverallDiscount($open_overall_discount)
    {
        return $this->save(['open_overall_discount' => (int)$open_overall_discount]);
    }

    /**
     * 删除自助餐-软删除
     */
    public function setDelete()
    {
        if ($this->getCanDelete($this) == 0) {
            $this->error = '自助餐正在使用中，不可删除';
            return false;
        }
        $this->startTrans();
        try {
            // 删除自助餐商品
            foreach ($this->buffetProducts as $buffetProduct) {
                $buffetProduct->delete();
            }
            // 删除顾客类型定价
            foreach ($this->buffetCustomerType as $cutomerType) {
                $cutomerType->delete();
            }
            // 删除对应多语言
            $this->multiLanguageName->delete();
            $this->multiLanguageName->clearCache($this->multi_language_name_uuid);
            // 删除自助餐
            $this->delete();

            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }

    /**
     * 处理数据超过范围值时，返回提示信息
     */
    private function alertFormData($data)
    {
        $limits = [
            'price' => ['range' => [0, 100000000], 'message' => '价格必须在0到100000000之间'],
            'time_limit' => ['range' => [0, 999], 'message' => '用餐时间必须在0到999之间'],
            'sort' => ['range' => [0, 999], 'message' => '排序必须在0到999之间'],
            'limit_num' => ['range' => [1, 1000], 'message' => '限购数量必须在1到1000之间'],
        ];
        foreach ($limits as $key => $value) {
            if (array_key_exists($key, $data) && ($data[$key] < $value['range'][0] || $data[$key] > $value['range'][1])) {
                return $value['message'];
            }
        }
        return '';
    }

    /**
     * 处理数据为负数时，自动转换为0
     */
    private function sanitizeFormData($keys, $data)
    {
        foreach ($keys as $key) {
            if (array_key_exists($key, $data)) {
                $data[$key] = max(0, $data[$key]);
            }
        }
        return $data;
    }
}
