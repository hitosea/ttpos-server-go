# 开帐

```json
{
    "pos_profile_name": "美宜佳-2号店 - POS",
    "cashier_email": "7171433574400000@ttpos-user.com",
    "period_start_date": {{$date.timestamp}},
    "company_abbr": "myj02d",
    "open_pos_entry_detail": [
        {
            "mode_of_payment": "Cash",
            "opening_amount": 0
        },
        {
            "mode_of_payment": "Balance",
            "opening_amount": 0
        }
    ]
}
```

```json
{
    "code": "0",
    "message": "创建开帐成功",
    "data": {
        "open_pos_entry_info": {
            "open_pos_entry_detail": [
                {
                    "mode_of_payment": "Cash",
                    "opening_amount": 0
                },
                {
                    "mode_of_payment": "Balance",
                    "opening_amount": 0
                }
            ],
            "open_pos_entry_name": "POS-OPE-2025-00326",
            "pos_profile_name": "美宜佳-2号店 - POS",
            "cashier_email": "7171433574400000@ttpos-user.com",
            "company_abbr": "myj02d",
            "branch": ""
        },
        "@type": "type.googleapis.com/selling.OpenPosEntryResp"
    }
}
```