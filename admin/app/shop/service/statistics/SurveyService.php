<?php

namespace app\shop\service\statistics;

use app\common\library\helper;
use PhpOffice\PhpSpreadsheet\IOFactory;
use PhpOffice\PhpSpreadsheet\Style\Fill;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;
use PhpOffice\PhpSpreadsheet\Cell\Coordinate;

/**
 * 店内概况
 */
class SurveyService
{

    /**
     * 列表导出
     */
    public function surveyExport($list)
    {
        // 授权无会员权限
        $isOpenMember = request()->licenses['is_open_member'] ?? 0;

        // 所有区域数据
        $allRegionData = [];
        foreach ($list as $date => $data) {
            foreach ($data['regionData'] as $region) {
                $allRegionData[$region['area_id']] = $region;
            }
        }

        //
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        // 列宽
        $sheet->getColumnDimension('A')->setWidth(18);
        for ($col = 2; $col <= 32; $col++) { // 从第2列(B)到第32列(AF)
            $columnLetter = Coordinate::stringFromColumnIndex($col);
            $sheet->getColumnDimension($columnLetter)->setWidth(12);
        }
        // 设置工作表标题名称
        $sheet->setTitle(__('店内概况'));
        //
        $sheet->setCellValue('A1', date('Y/m/d'));
        $sheet->setCellValue('B1', __('营业天数'));
        $sheet->setCellValue('C1', count($list));
        //
        $index = 2;
        $sheet->setCellValue('A' . $index, __('日期'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('总销售额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('营业收入'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('服务费'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('支付手续费'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('税费'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('商品数量'));
        if ($isOpenMember) {
            $sheet->setCellValue('A' . ($index = $index + 1), __('新增会员数/会员折扣'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('优惠折扣/优惠占比'));
        } else {
            $sheet->setCellValue('A' . ($index = $index + 1), __('优惠折扣/优惠占比'));
        }
        $sheet->setCellValue('A' . ($index = $index + 1), __('退款金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('赠菜折扣/赠菜数量'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('免单折扣/免单数量'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('实收金额'));
        // ------区域数据--------
        $sheet->setCellValue('A' . ($index = $regionRow = $index + 1), __('区域数据'));
        foreach ($allRegionData as $data) {
            $sheet->setCellValue('A' . ($index = $index + 1), $data['area_name']);
            $sheet->setCellValue('A' . ($index = $index + 1), __('总销售额'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('营业收入'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('商品数量'));
        }
        // ------订单数据--------
        $sheet->setCellValue('A' . ($index = $orderRow = $index + 1), __('订单数据'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('合计订单数'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('最小订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('最大订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('平均订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('桌台方式'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('桌数'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('人数'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('最小订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('最大订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('平均订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('点餐方式'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('订单数'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('最小订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('最大订单金额'));
        $sheet->setCellValue('A' . ($index = $index + 1), __('平均订单金额'));
        $sheet->setCellValue('A' . ($index = $payRow = $index + 1), __('支付数据'));
        // 纵向的支付数据
        $paymentType = array_values($list)[0]['incomes'];
        foreach ($paymentType as $value) {
            $sheet->setCellValue('A' . ($index + 1), $value['pay_type_name']);
            $index++;
        }
        // 合并单元格并设置背景颜色为浅灰色
        foreach ([$regionRow, $orderRow, $payRow] as $line) {
            $sheet->mergeCells("A$line:AF$line");
            $sheet->getStyle("A$line")->getFill()->setFillType(Fill::FILL_SOLID)->getStartColor()->setARGB('FFD3D3D3');
        }

        // 横向的日期数据
        $columnIndex = 2; // 从B列开始
        foreach ($list as $date => $data) {
            $index = 2;
            $columnLetter = Coordinate::stringFromColumnIndex($columnIndex);
            $sheet->setCellValue($columnLetter . $index, $date)->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT); // 设置日期
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['receivable_price']); //总销售额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['business_price']); //营业收入
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['service_money']); //服务费
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['pay_fee_money']); //支付手续费
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['consumption_tax_money']); //税费
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['product_num']); //商品数量
            if ($isOpenMember) {
                // 新增会员数/会员折扣  -  优惠折扣/优惠占比
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['user_count'] . "/" . Helper::number2($data['user_discount_money']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
                $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['discount_money']) . "/" . Helper::number2($data['discount_ratio']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            } else {
                // 优惠折扣/优惠占比
                $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['discount_money']) . "/" . Helper::number2($data['discount_ratio']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            }
            // 赠菜折扣/赠菜数量 - 免单折扣/免单数量
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['refund_money']); //退款金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['free_product_price']) . "/" . Helper::number2($data['free_product_num']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['free_order_price']) . "/" . Helper::number2($data['free_order_num']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['received_price']); //实收金额
            // ------区域数据--------
            $sheet->setCellValue($columnLetter . ($index = $index + 1), ''); // 区域数据 - 空白行
            foreach ($allRegionData as $region) {
                $salesPrice = 0;
                $businessPrice = 0;
                $productNum = 0;
                foreach ($data['regionData'] as $regionData) {
                    if ($region['area_id'] == $regionData['area_id']) {
                        $salesPrice = $regionData['sales_price'];
                        $businessPrice = $regionData['business_price'];
                        $productNum = $regionData['product_num'];
                    }
                }
                $sheet->setCellValue($columnLetter . ($index = $index + 2), $salesPrice); //总销售额
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $businessPrice); //营业收入
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $productNum); //商品数量
            }
            // ------订单数据--------
            $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['total_order_num']); //合计订单数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['min_order_price']); //最小订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['max_order_price']); //最大订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['avg_order_price']); //平均订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['table_order_num']); //桌数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['table_people_num']); //人数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['table_min_order_price']); //最小订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['table_max_order_price']); //最大订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['table_avg_order_price']); //平均订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['cashier_order_num']); //订单数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['cashier_min_order_price']); //最小订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['cashier_max_order_price']); //最大订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['cashier_avg_order_price']); //平均订单金额
            // 支付数据
            $payColumnIndex = $index + 1;
            foreach ($data['incomes'] as $value) {
                $sheet->setCellValue($columnLetter . ($payColumnIndex + 1), $value['price']);
                $payColumnIndex++;
            }
            $columnIndex++;
        }

        // 保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('店内概况数据') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }
}
