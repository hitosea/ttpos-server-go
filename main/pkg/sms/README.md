### 短信发送平台

#### 请求参数说明
##### 请求头
| 字段      | 类型     | 是否必填   |  说明|
|---------|--------|------|------|
| api-key | string |  是|    唯一密钥 |

##### 请求体
| 字段      | 类型             | 是否必填 | 说明            |
|---------|----------------|------|---------------|
| template_id | string         | 是    | 模板ID          |
| phone | string         | 是    | 手机号，需要加上地区号前缀 |
| language | string         | 是    | 语言            |
| params | map[string]any | 是    | 替换参数          |

#### 响应说明
形如：
```json
{"code":401,"msg":"Invalid API Key"}
```
code 为0表成功，非0表示失败

##### 模板说明
###### 1、会员消费
| 字段          | 值            | 
|-------------|--------------| 
| template_id | member_consumption         | 
| language    | 可选zh、en | 

###### 1.1、会员消费替换参数
| 字段      | 类型      | 是否必填 | 说明     |
|---------|---------|------|--------|
| company | string  | 是    | 商家名称   |
| consumption | float64 | 是    | 消费金额   |
| member_pay | float64 | 是    | 会员支付   |
| increase_points | float64 | 是    | 获得积分   |
| balance | float64 | 是    | 当前会员余额 |
| points_balance | float64 | 是    | 积分余额 |

###### 2、会员充值
| 字段          | 值            | 
|-------------|--------------| 
| template_id | member_recharge         | 
| language    | 可选zh、en |

###### 2.1、会员充值替换参数
| 字段      | 类型      | 是否必填 | 说明     |
|---------|---------|------|--------|
| company | string  | 是    | 商家名称   |
| recharge | float64 | 是    | 充值金额   |
| bonus_money | float64 | 是    | 赠送金额   |
| bonus_points | float64 | 是    | 赠送会员积分   |
| balance | float64 | 是    | 当前会员余额 |
| points_balance | float64 | 是    | 积分余额   |


###### 3、会员充值退款
| 字段          | 值                      | 
|-------------|------------------------| 
| template_id | member_recharge_refund | 
| language    | 可选zh、en                |

###### 3.1、会员充值退款替换参数
| 字段      | 类型      | 是否必填 | 说明     |
|---------|---------|------|--------|
| company | string  | 是    | 商家名称   |
| recharge_refund | float64 | 是    | 退款金额   |
| balance | float64 | 是    | 当前会员余额 |
| points_balance | float64 | 是    | 积分余额   |

###### 4、会员用餐订单退款
| 字段          | 值                      | 
|-------------|------------------------| 
| template_id | member_order_refund | 
| language    | 可选zh、en                |

###### 4.1、会员充值退款替换参数
| 字段      | 类型      | 是否必填 | 说明     |
|---------|---------|------|--------|
| company | string  | 是    | 商家名称   |
| order_refund | float64 | 是    | 退款金额   |
| balance | float64 | 是    | 当前会员余额 |
| points_balance | float64 | 是    | 积分余额   |



#### 请求示例

###### 1、发送短信
```curl
curl --request POST \
  --url http://192.168.100.245:8787/api/sms/send \
  --header 'api-key: 6Q76zPZMrc6KRhGLt4ilBJCoBEKqJSpZT74K3sIXJvz4gpD33Wsd78V7rxq72j8R' \
  --header 'content-type: application/json' \
  --data '{
  "template_id":"member_recharge",
  "phone":"+8617777777777",
  "language":"zh",
  "params":{
    "company":"向西餐馆",
    "recharge":100,
    "bonus_money":50,
    "bonus_points":100,
    "balance":100,
    "points_balance":100
  }
}'
```

###### 2、查询短信状态，仅限容器内部，message_id可传递多个，以英文逗号分隔
```curl
curl -XGET "http://127.0.0.1:8080/api/sms/query?message_id=G183FE43F437E343EHMDP1009"
```

#### 其他说明

用户发送短信后，如果未收到回执，没5分钟回自动查询短信状态