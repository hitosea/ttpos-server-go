# grab.Grab / NotifyMenuUpdate

```json
{
    "merchant_id": "GFSBPOS-616-088"
}

```

```json
{
    "code": "0",
    "message": "成功",
    "data": {
        "request_id": "61406e89-559b-4c7a-aea6-02ca93a5687a",
        "@type": "type.googleapis.com/grab.NotifyMenuUpdateResp"
    }
}

```

-----

# menu.MenuService / UpdateMenuItem

```json
{
    "merchant_id": "GFSBPOS-616-088",
    "item_id": "TTPOS-ITEM-3701988558241793",
    "available_status": "UNAVAILABLE"
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "merchant_id": "GFSBPOS-616-088",
        "record_id": "TTPOS-ITEM-3701988558241793",
        "record_type": "ITEM",
        "@type": "type.googleapis.com/menu.UpdateMenuItemResp"
    }
}
```


-----

# menu.MenuService / UpdateMenuModifier

```json
{
    "merchant_id": "GFSBPOS-616-088",
    "modifier_name":"Piece",
    "modifier_id": "TTPOS-FLAVOR-592",
    "available_status": "UNAVAILABLE",
    "is_free": false
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "merchant_id": "GFSBPOS-616-088",
        "record_id": "TTPOS-FLAVOR-592",
        "record_type": "MODIFIER",
        "@type": "type.googleapis.com/menu.UpdateMenuModifierResp"
    }
}
```