#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
按门店对比两天的实盘数量（仅统计已审核的盘点单）

使用方法：
  python3 ttpos-scripts/db_generator/compare_daily_stock.py <day1> <day2> [input.csv]

参数：
  day1      - 第一天日期（如 24 或 2026-01-24）
  day2      - 第二天日期（如 25 或 2026-01-25）
  input.csv - 输入文件（可选，默认 ttpos-scripts/db_generator/output/data.csv）

示例：
  python3 ttpos-scripts/db_generator/compare_daily_stock.py 24 25
  python3 ttpos-scripts/db_generator/compare_daily_stock.py 24 25 ttpos-scripts/db_generator/output/data.csv
  python3 ttpos-scripts/db_generator/compare_daily_stock.py 2026-01-24 2026-01-25 ttpos-scripts/db_generator/output/data.csv

输出：CSV 和 Markdown 表格文档
"""

import csv
import sys
import os
import json
from collections import defaultdict
from datetime import datetime

def extract_en_name(name_json):
    """从多语言 JSON 中提取英文名称"""
    if not name_json:
        return ""
    try:
        name_obj = json.loads(name_json)
        return name_obj.get('en', '') or name_obj.get('zh', '') or str(name_obj)
    except (json.JSONDecodeError, TypeError):
        return name_json

def parse_day(day_str):
    """解析日期参数，返回用于匹配的字符串"""
    # 如果是完整日期格式 2026-01-24，提取日部分
    if '-' in day_str:
        return day_str  # 完整日期直接返回
    # 如果只是数字（如 24），转为匹配格式
    return f"01-{day_str.zfill(2)}"

def main():
    # 解析参数
    if len(sys.argv) < 3:
        print("错误：请指定两个日期参数")
        print()
        print("使用方法：")
        print("  python3 ttpos-scripts/db_generator/compare_daily_stock.py <day1> <day2> [input.csv]")
        print()
        print("示例：")
        print("  python3 ttpos-scripts/db_generator/compare_daily_stock.py 24 25")
        print("  python3 ttpos-scripts/db_generator/compare_daily_stock.py 24 26 ttpos-scripts/db_generator/output/data.csv")
        sys.exit(1)

    day1_input = sys.argv[1]
    day2_input = sys.argv[2]
    input_file = sys.argv[3] if len(sys.argv) > 3 else "ttpos-scripts/db_generator/output/data.csv"

    # 解析日期
    day1_match = parse_day(day1_input)
    day2_match = parse_day(day2_input)

    # 用于显示的标签
    day1_label = day1_input if '-' not in day1_input else day1_input.split('-')[-1]
    day2_label = day2_input if '-' not in day2_input else day2_input.split('-')[-1]

    # 生成输出文件名
    output_dir = os.path.dirname(input_file) or "ttpos-scripts/db_generator/output"
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_md = os.path.join(output_dir, f"compare_{day1_label}_{day2_label}_{timestamp}.md")
    output_csv = os.path.join(output_dir, f"compare_{day1_label}_{day2_label}_{timestamp}.csv")

    print("=" * 60)
    print(f"按门店对比 {day1_label} 号和 {day2_label} 号实盘数量")
    print("=" * 60)
    print(f"输入文件: {input_file}")
    print(f"对比日期: {day1_match} vs {day2_match}")
    print(f"输出 MD: {output_md}")
    print(f"输出 CSV: {output_csv}")
    print(f"统计条件: 仅统计已审核的盘点单")
    print()

    # 数据结构：{shop: {item_code: {day1: counted, day2: counted, name: xxx}}}
    # 用 None 表示未盘点，区分"盘了0"和"没盘"
    data = defaultdict(lambda: defaultdict(lambda: {"day1": None, "day2": None, "name": ""}))

    with open(input_file, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            status = row.get('状态', '')
            if status != '已审核':
                continue

            shop = row.get('门店名称', '')
            item_code = row.get('物品code', '')
            item_name = row.get('物品名称', '')
            create_time = row.get('创建时间', '')
            date = create_time.split(' ')[0] if create_time else ''

            counted = float(row.get('实盘数量', 0) or 0)

            # 匹配指定的两天
            if day1_match in date:
                data[shop][item_code]["day1"] = counted
                data[shop][item_code]["name"] = item_name
            elif day2_match in date:
                data[shop][item_code]["day2"] = counted
                if not data[shop][item_code]["name"]:
                    data[shop][item_code]["name"] = item_name

    if not data:
        print("没有找到符合条件的数据")
        return

    # 写入 CSV 文件（Excel 可直接打开）
    with open(output_csv, 'w', encoding='utf-8-sig', newline='') as f:
        writer = csv.writer(f)
        writer.writerow(['门店名称', '物品code', '物品名称', f'{day1_label}号实盘', f'{day2_label}号实盘', f'差异({day2_label}-{day1_label})'])

        for shop in sorted(data.keys()):
            items = data[shop]
            for item_code in sorted(items.keys()):
                item = items[item_code]
                item_name = extract_en_name(item["name"])
                counted_day1 = item["day1"]
                counted_day2 = item["day2"]

                # 格式化：None 显示为 "-"
                str_day1 = f"{counted_day1:.2f}" if counted_day1 is not None else "-"
                str_day2 = f"{counted_day2:.2f}" if counted_day2 is not None else "-"

                # 差异：任一天没盘则显示 "-"
                if counted_day1 is not None and counted_day2 is not None:
                    diff = counted_day2 - counted_day1
                    str_diff = f"{diff:.2f}"
                else:
                    str_diff = "-"

                writer.writerow([shop, item_code, item_name, str_day1, str_day2, str_diff])

    # 写入 Markdown 文件
    with open(output_md, 'w', encoding='utf-8') as f:
        f.write(f"# 门店盘点对比报表（{day1_label}号 vs {day2_label}号）\n\n")
        f.write(f"> 生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
        f.write(f"> 统计条件: 仅统计已审核的盘点单\n\n")

        for shop in sorted(data.keys()):
            f.write(f"## {shop}\n\n")
            f.write(f"| 物品code | 物品名称 | {day1_label}号实盘 | {day2_label}号实盘 | 差异({day2_label}-{day1_label}) |\n")
            f.write("|----------|----------|----------|----------|-------------|\n")

            items = data[shop]
            for item_code in sorted(items.keys()):
                item = items[item_code]
                item_name = extract_en_name(item["name"])
                counted_day1 = item["day1"]
                counted_day2 = item["day2"]

                # 格式化：None 显示为 "-"
                str_day1 = f"{counted_day1:.2f}" if counted_day1 is not None else "-"
                str_day2 = f"{counted_day2:.2f}" if counted_day2 is not None else "-"

                # 差异：任一天没盘则显示 "-"
                if counted_day1 is not None and counted_day2 is not None:
                    diff = counted_day2 - counted_day1
                    if diff > 0:
                        diff_str = f"+{diff:.2f}"
                    elif diff < 0:
                        diff_str = f"{diff:.2f}"
                    else:
                        diff_str = "0.00"
                else:
                    diff_str = "-"

                f.write(f"| {item_code} | {item_name} | {str_day1} | {str_day2} | {diff_str} |\n")

            f.write("\n")

    print(f"门店数量: {len(data)}")
    print(f"✓ 完成")
    print(f"  CSV (Excel): {output_csv}")
    print(f"  Markdown: {output_md}")

if __name__ == "__main__":
    main()
