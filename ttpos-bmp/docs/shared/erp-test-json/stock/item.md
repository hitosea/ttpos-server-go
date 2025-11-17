

## 创建带属性商品 item.ItemService / SaveItem

```json
{
    "item_name": "SP大白兔-{{$string.alpha(length=5)}}",
    "item_group": "Products",
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
    "has_variants":1,
    "attributes":[
        {
            "attribute_name":"Color"
        },
         {
            "attribute_name":"CFG-Specifications"
        }
    ],
    "pos_attribute_group" :[
        {
            "item_group":"SX3688457203482625",
            "group_name":"大-SP大白兔-温度",
            "min_select":1,
            "max_select":1,
            "attribute_list" :[
                {
                    "item":"SXVfff112312312312444"
                }
            ]
        }
    ]
}
```