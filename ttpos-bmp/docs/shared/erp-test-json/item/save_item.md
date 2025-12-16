# 保存物品 item.ItemService / SaveItem

## 允许负库存

```json
{
      "item_name": "WRP大白兔-{{$string.alpha(length=5)}}",
    "item_group": "RawMaterial",
    "stock_uom": "Ampere-Hour",
    "branch": "大富翁桌游馆",
    "company_abbr": "大",
    "item_specification":"small",
    "classification": "6",
    "classification_code" :"66",
    "internal_code": "6666-{{$string.alpha(length=5)}}",
    "uoms" :[
        {
            "uom":"Meter",
            "conversion_factor": 1.15
        }
    ],
    "allow_negative_stock": true
}
```

```json
{
    "code": "0",
    "message": "保存物品成功",
    "data": {
        "uoms": [
            {
                "uom": "Meter",
                "conversion_factor": 1.15
            }
        ],
        "attributes": [],
        "item_name": "WRP大白兔-hKJue",
        "item_group": "RawMaterial",
        "stock_uom": "Ampere-Hour",
        "item_code": "WPR3702175054301185",
        "valuation_rate": 0,
        "barcode": "",
        "opening_stock": 0,
        "is_stock_item": false,
        "template_item_code": "",
        "branch": "大富翁桌游馆",
        "company_abbr": "大",
        "company": "大富翁桌游馆",
        "item_specification": "small",
        "disabled": false,
        "classification": "6",
        "classification_code": "66",
        "internal_code": "6666-PypHK",
        "not_for_sale": false,
        "purchase_uom": "",
        "has_variants": false,
        "item_group_name": "",
        "variant_of": "",
        "allow_negative_stock": true,
        "@type": "type.googleapis.com/item.ItemInfo"
    }
}
```

----

## 不允许负库存

```json

{
      "item_name": "WRP大白兔-{{$string.alpha(length=5)}}",
    "item_group": "RawMaterial",
    "stock_uom": "Ampere-Hour",
    "branch": "大富翁桌游馆",
    "company_abbr": "大",
    "item_specification":"small",
    "classification": "6",
    "classification_code" :"66",
    "internal_code": "6666-{{$string.alpha(length=5)}}",
    "uoms" :[
        {
            "uom":"Meter",
            "conversion_factor": 1.15
        }
    ],
    "allow_negative_stock": false
}
```

```json
{
    "code": "0",
    "message": "保存物品成功",
    "data": {
        "uoms": [
            {
                "uom": "Meter",
                "conversion_factor": 1.15
            }
        ],
        "attributes": [],
        "item_name": "WRP大白兔-WSlLy",
        "item_group": "RawMaterial",
        "stock_uom": "Ampere-Hour",
        "item_code": "WPR3702175245142017",
        "valuation_rate": 0,
        "barcode": "",
        "opening_stock": 0,
        "is_stock_item": false,
        "template_item_code": "",
        "branch": "大富翁桌游馆",
        "company_abbr": "大",
        "company": "大富翁桌游馆",
        "item_specification": "small",
        "disabled": false,
        "classification": "6",
        "classification_code": "66",
        "internal_code": "6666-CSyWM",
        "not_for_sale": false,
        "purchase_uom": "",
        "has_variants": false,
        "item_group_name": "",
        "variant_of": "",
        "allow_negative_stock": false,
        "@type": "type.googleapis.com/item.ItemInfo"
    }
}

```


---

## 不传负库存
```json
{
      "item_name": "WRP大白兔-{{$string.alpha(length=5)}}",
    "item_group": "RawMaterial",
    "stock_uom": "Ampere-Hour",
    "branch": "大富翁桌游馆",
    "company_abbr": "大",
    "item_specification":"small",
    "classification": "6",
    "classification_code" :"66",
    "internal_code": "6666-{{$string.alpha(length=5)}}",
    "uoms" :[
        {
            "uom":"Meter",
            "conversion_factor": 1.15
        }
    ]
}
```

```json
{
    "code": "0",
    "message": "保存物品成功",
    "data": {
        "uoms": [
            {
                "uom": "Meter",
                "conversion_factor": 1.15
            }
        ],
        "attributes": [],
        "item_name": "WRP大白兔-xGEUV",
        "item_group": "RawMaterial",
        "stock_uom": "Ampere-Hour",
        "item_code": "WPR3702175324833793",
        "valuation_rate": 0,
        "barcode": "",
        "opening_stock": 0,
        "is_stock_item": false,
        "template_item_code": "",
        "branch": "大富翁桌游馆",
        "company_abbr": "大",
        "company": "大富翁桌游馆",
        "item_specification": "small",
        "disabled": false,
        "classification": "6",
        "classification_code": "66",
        "internal_code": "6666-SbUZI",
        "not_for_sale": false,
        "purchase_uom": "",
        "has_variants": false,
        "item_group_name": "",
        "variant_of": "",
        "@type": "type.googleapis.com/item.ItemInfo"
    }
}

```


----

## 修改允许负库存

```json

{
    "item_code": "WPR3702175442274305",
    "item_name": "WRP大白兔-TCiuE",
    "item_group": "RawMaterial",
    "stock_uom": "Ampere-Hour",
    "branch": "大富翁桌游馆",
    "company_abbr": "大",
    "item_specification":"small",
    "classification": "6",
    "classification_code" :"66",
    "uoms" :[
        {
            "uom":"Meter",
            "conversion_factor": 1.15
        }
    ],
     "allow_negative_stock": true
}

```

```json

{
    "code": "0",
    "message": "保存物品成功",
    "data": {
        "uoms": [
            {
                "uom": "Meter",
                "conversion_factor": 1.15
            }
        ],
        "attributes": [],
        "item_name": "WRP大白兔-TCiuE",
        "item_group": "RawMaterial",
        "stock_uom": "Ampere-Hour",
        "item_code": "WPR3702175442274305",
        "valuation_rate": 0,
        "barcode": "",
        "opening_stock": 0,
        "is_stock_item": false,
        "template_item_code": "",
        "branch": "大富翁桌游馆",
        "company_abbr": "大",
        "company": "",
        "item_specification": "small",
        "disabled": false,
        "classification": "6",
        "classification_code": "66",
        "internal_code": "",
        "not_for_sale": false,
        "purchase_uom": "",
        "has_variants": false,
        "item_group_name": "",
        "variant_of": "",
        "allow_negative_stock": true,
        "@type": "type.googleapis.com/item.ItemInfo"
    }
}
```

----

## 存在负库存时不允许修改

```json

{
    "item_code": "WPR3685909115568129",
    "item_name": "Bread dough",
    "item_group": "RawMaterial",
    "stock_uom": "个ชิ้น",
     "allow_negative_stock": false
}
```

```json
{
    "code": "1",
    "message": "物料已产生负库存，请修正库存后再进行操作",
    "data": null
}
```