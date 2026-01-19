// 查询没有估值率或成本价的物品
// 
// 注意：此脚本需要通过 ERPNext API 查询，因为估值率信息存储在 ERPNext 中
// 建议使用 SQL 查询脚本（check_items_without_valuation_rate.sql）或直接通过 ERPNext 界面查询
//
// 使用方法：
//   1. 通过 ERPNext API 查询所有物品的估值率信息
//   2. 或者使用 SQL 脚本查询（如果 ERPNext 数据库可访问）
//   3. 或者通过 ERPNext 界面：Items > Reports > Item Price List
//
// 通过盘点解决估值率问题：
//   在创建盘点单时，为每个物品明细指定 valuation_rate 字段
//   系统会优先使用请求中的估值率，如果未指定则从 Item 获取
//
// 估值率获取优先级：
//   1. 请求值（如果 > 0）
//   2. Item.ValuationRate（如果 > 0）
//   3. Item.StandardRate（如果 > 0）
//   4. Item.LastPurchaseRate（如果 > 0）
//   5. 如果所有价格都是 0，返回 0 并设置 AllowZeroValuationRate = 1
package main

// 此文件保留作为参考，实际查询建议使用 SQL 脚本或 ERPNext API

