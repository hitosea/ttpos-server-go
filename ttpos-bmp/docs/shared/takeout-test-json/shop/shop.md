## shop.Shop / ActivateShop 激活门店

### Grab
```json

{
    "provider_name": "grab",
      "shop_uuid": "{{$number.bigInt}}",
    "request_id": "{{$number.bigInt}}"
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "shop_uuid": "909704028624238",
        "provider_name": "grab",
        "self_serve_url": "https://developer.grab.com/self-serve-activation/NzFjMTU0NGQtYWNhMC00ZWRiLWI3MGQtMDgyMmFmYmJhMGI0/activate",
        "request_id": "812114163753860",
        "updated_at": "1767775553",
        "@type": "type.googleapis.com/shop.ActivateShopResp"
    }
}
```

------

### LINEMAN

```json
{
    "provider_name": "lineman",
      "shop_uuid": "{{$number.bigInt}}",
    "request_id": "{{$number.bigInt}}"
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "shop_uuid": "616218476261511",
        "provider_name": "lineman",
        "self_serve_url": "",
        "request_id": "215986844631031",
        "updated_at": "1767775845",
        "@type": "type.googleapis.com/shop.ActivateShopResp"
    }
}
```



------------
## shop.Shop / GetShopProviderCfg 获取门店外卖配置

```json
{
    "shop_uuid": "909704028624238"
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "providers": [
            {
                "provider_name": "grab",
                "provider_merchant_id": "",
                "provider_shop_status": "SYNCING",
                "updated_at": "1767775553"
            }
        ],
        "shop_uuid": "909704028624238",
        "@type": "type.googleapis.com/shop.GetShopProviderCfgResp"
    }
}
```

----------

```json
{
    "shop_uuid": "909704028624238",
    "provider_name": "lineman"
}
```


```json
{
    "code": "0",
    "message": "success",
    "data": {
        "providers": [
            {
                "provider_name": "lineman",
                "provider_merchant_id": "",
                "provider_shop_status": "INACTIVE",
                "updated_at": "0"
            }
        ],
        "shop_uuid": "909704028624238",
        "@type": "type.googleapis.com/shop.GetShopProviderCfgResp"
    }
}
```