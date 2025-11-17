# 门店间直接调拨

## case1 门店间直接调拨
```json
{
    "from_company_abbr": "TH0133",
    "from_branch": "133门店",
    "to_company_abbr": "TH0121",
    "to_branch": "121门店",
    "from_warehouse": "133门店-Normal-Default - TH0133",
    "to_warehouse": "121门店-Normal-Default - TH0121",
    "items": [
        {
            "item_code": "WPR3689220113825793",
            "qty": 6,
            "uom": "蓝象总店-piece"
        }
    ]
}
```
### 响应

```json

{
    "code": "0",
    "message": "材料调拨成功",
    "data": {
        "from_receipt": {
            "po_no": "PUR-ORD-2025-00447",
            "so_no": "SAL-ORD-2025-00186",
            "from_company_abbr": "TH0133",
            "to_company_abbr": "TH0121"
        },
        "to_receipt": {
            "po_no": "PUR-ORD-2025-00447",
            "so_no": "SAL-ORD-2025-00186",
            "from_company_abbr": "TH0133",
            "to_company_abbr": "TH0121"
        },
        "audit_receipt": {
            "po_no": "PUR-ORD-2025-00447",
            "so_no": "SAL-ORD-2025-00186",
            "from_company_abbr": "TH0133",
            "to_company_abbr": "TH0121"
        },
        "@type": "type.googleapis.com/material_transfer.MaterialTransferResp"
    }
}
```
----

## case2 门店跨一级调拨

```json
{
  "from_company_abbr": "TH0133",
  "from_branch": "133门店",
  "to_company_abbr": "TH0121",
  "to_branch": "121门店",
  "from_warehouse": "133门店-Normal-Default - TH0133",
  "to_warehouse": "121门店-Normal-Default - TH0121",
  "from_parent_company_abbr":"FF",
  "from_parent_branch": "wallace-FF",

  "items": [
    {
      "item_code": "WPR3689220113825793",
      "qty": 6,
      "uom": "蓝象总店-piece"
    }
  ]
}
```

### 响应

```json
{
    "code": "0",
    "message": "材料调拨成功",
    "data": {
        "from_receipt": {
            "po_no": "PUR-ORD-2025-00457",
            "so_no": "SAL-ORD-2025-00204",
            "from_company_abbr": "TH0133",
            "to_company_abbr": "FF"
        },
        "to_receipt": {
            "po_no": "PUR-ORD-2025-00458",
            "so_no": "SAL-ORD-2025-00205",
            "from_company_abbr": "FF",
            "to_company_abbr": "TH0121"
        },
        "audit_receipt": {
            "po_no": "PUR-ORD-2025-00458",
            "so_no": "SAL-ORD-2025-00205",
            "from_company_abbr": "FF",
            "to_company_abbr": "TH0121"
        },
        "@type": "type.googleapis.com/material_transfer.MaterialTransferResp"
    }
}

```

```json

{
    "from_company_abbr": "TH0133",
    "from_branch": "133门店",
    "to_company_abbr": "TH0121",
    "to_branch": "121门店",
    "from_warehouse": "133门店-Normal-Default - TH0133",
    "to_warehouse": "121门店-Normal-Default - TH0121",

    "to_parent_company_abbr": "JJ",
    "to_parent_branch": "wallace-JJ",
    "items": [
        {
            "item_code": "WPR3689220113825793",
            "qty": 6,
            "uom": "蓝象总店-piece"
        }
    ]
}
```

```json

{
    "code": "0",
    "message": "材料调拨成功",
    "data": {
        "from_receipt": {
            "po_no": "PUR-ORD-2025-00455",
            "so_no": "SAL-ORD-2025-00202",
            "from_company_abbr": "TH0133",
            "to_company_abbr": "JJ"
        },
        "to_receipt": {
            "po_no": "PUR-ORD-2025-00456",
            "so_no": "SAL-ORD-2025-00203",
            "from_company_abbr": "JJ",
            "to_company_abbr": "TH0121"
        },
        "audit_receipt": {
            "po_no": "PUR-ORD-2025-00456",
            "so_no": "SAL-ORD-2025-00203",
            "from_company_abbr": "JJ",
            "to_company_abbr": "TH0121"
        },
        "@type": "type.googleapis.com/material_transfer.MaterialTransferResp"
    }
}
```

-----
## case3 调入调出审核节点都不一样
```json
{
  "from_company_abbr": "TH0133",
  "from_branch": "133门店",
  "to_company_abbr": "TH0121",
  "to_branch": "121门店",
  "from_warehouse": "133门店-Normal-Default - TH0133",
  "to_warehouse": "121门店-Normal-Default - TH0121",
  "from_parent_company_abbr":"FF",
  "from_parent_branch": "wallace-FF",
  "to_parent_company_abbr": "JJ",
  "to_parent_branch": "wallace-JJ",
  "items": [
    {
      "item_code": "WPR3689220113825793",
      "qty": 6,
      "uom": "蓝象总店-piece"
    }
  ]
}

```

### 响应

```json
{
  "code": "0",
  "message": "材料调拨成功",
  "data": {
    "from_receipt": {
      "po_no": "PUR-ORD-2025-00450",
      "so_no": "SAL-ORD-2025-00197",
      "from_company_abbr": "TH0133",
      "to_company_abbr": "FF"
    },
    "to_receipt": {
      "po_no": "PUR-ORD-2025-00452",
      "so_no": "SAL-ORD-2025-00199",
      "from_company_abbr": "JJ",
      "to_company_abbr": "TH0121"
    },
    "audit_receipt": {
      "po_no": "PUR-ORD-2025-00451",
      "so_no": "SAL-ORD-2025-00198",
      "from_company_abbr": "FF",
      "to_company_abbr": "JJ"
    },
    "@type": "type.googleapis.com/material_transfer.MaterialTransferResp"
  }
}

```