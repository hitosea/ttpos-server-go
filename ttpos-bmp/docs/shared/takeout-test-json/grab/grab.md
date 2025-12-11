# Grab 测试 JSON

## 获取自助激活链接

```json
{
    "provider_name": "grab",
    "shop_uuid": "8619817471094482",
    "request_id": "{{$number.bigInt}}"
}
```



```json
{
    "provider_name": "grab",
    "self_serve_url": "https://developer.grab.com/self-serve-activation/NzI4NTRjOGItM2VlMy00ZjUyLWFiZjYtNzg2MTVlZTI3MmMx/activate",
    "request_id": "112828527335683"
}
```


----

## 获取激活状态

```json
{
    "shop_uuid": 8619817471094482,
    "provider_name": "grab"
}
```


```json
{
    "shop_uuid": "8619817471094482",
    "provider_name": "grab",
    "provider_merchant_id": "GFSBPOS-255-417",
    "provider_shop_status": "ACTIVE",
    "updated_at": "1765457422"
}
```
