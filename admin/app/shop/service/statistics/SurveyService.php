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
        $isOpenGrabDelivery = request()->licenses['enable_grab_delivery'] ?? 0;

        // 所有区域数据
        $allRegionData = [];
        $paymentType = [];
        $list = $list['data'];
        foreach ($list as $date => $data) {
            foreach ($data['area_list'] as $region) {
                $allRegionData[$region['area_id']] = $region;
            }
            foreach ($data['payment_list'] as $payment) {
                $paymentType[$payment['payment_code']] = $payment;
            }
        }

        // 支付方式排序，先ID=0排在最后, 再按Sort升序排序，再按CreateTime降序排序， ID降序排序
        $paymentType = array_values($paymentType);
        usort($paymentType, function ($a, $b) {
            if ($a['id'] == 0) {
                return 1;
            }
            if ($b['id'] == 0) {
                return -1;
            }
            if ($a['sort'] == $b['sort']) {
                if ($a['create_time'] == $b['create_time']) {
                    return $a['id'] <=> $b['id'];
                }
                return $b['create_time'] <=> $a['create_time'];
            }
            return $a['sort'] <=> $b['sort'];
        });

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
        $sheet->setCellValue('A' . ($index = $index + 1), __('充值金额'));
        // v2.3.0新增 充值、外送数据
        if (request()->licenses['is_open_delivery'] == 1) { 
            $sheet->setCellValue('A' . ($index = $index + 1), __('外送销售'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('外送营收'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('外送退款'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('外送配送费'));
        }
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
        // v2.3.0新增 外送数据 
        if (request()->licenses['is_open_delivery'] == 1) {
            $sheet->setCellValue('A' . ($index = $index + 1), __('外送点餐'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('订单数'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('最小订单金额'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('最大订单金额'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('平均订单金额'));
        }
        // v2.15.0新增 Grab 数据
        if ($isOpenGrabDelivery) {
            $sheet->setCellValue('A' . ($index = $index + 1), __('Grab'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('订单数'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('最小订单金额'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('最大订单金额'));
            $sheet->setCellValue('A' . ($index = $index + 1), __('平均订单金额'));
        }
        $sheet->setCellValue('A' . ($index = $payRow = $index + 1), __('支付数据'));
        // 纵向的支付数据
        foreach ($paymentType as $value) {
            $sheet->setCellValue('A' . ($index + 1), $value['payment_name']);
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
            $sheet->setCellValue($columnLetter . $index, $data['day'])->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT); // 设置日期
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_sale_amount']); //总销售额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_business_amount']); //营业收入
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_service_fee']); //服务费
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_payment_fee']); //支付手续费
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_tax']); //税费
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_product_num']); //商品数量
            if ($isOpenMember) {
                // 新增会员数/会员折扣  -  优惠折扣/优惠占比
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_member_num'] . "/" . Helper::number2($data['total_discount_member']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
                $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['total_discount']) . "/" . Helper::number2($data['total_discount_ratio']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            } else {
                // 优惠折扣/优惠占比
                $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['total_discount']) . "/" . Helper::number2($data['total_discount_ratio']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            }
            // 赠菜折扣/赠菜数量 - 免单折扣/免单数量
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_refund_amount']); //退款金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['total_gift_amount']) . "/" . Helper::number2($data['total_gift_num']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            $sheet->setCellValue($columnLetter . ($index = $index + 1), Helper::number2($data['total_free_amount']) . "/" . Helper::number2($data['total_free_num']))->getStyle($columnLetter . $index)->getAlignment()->setHorizontal(\PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_RIGHT);
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_received_amount']); //实收金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_recharge_amount']); // 充值
            // v2.3.0新增 充值、外送数据 
            if (request()->licenses['is_open_delivery'] == 1) { 
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_takeout_sale_amount']); // 外送销售
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_takeout_business_amount']); // 外送营收
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_takeout_refund_amount']); // 外送退款
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_takeout_delivery_fee']); // 外送配送费
            }
            // ------区域数据--------
            $sheet->setCellValue($columnLetter . ($index = $index + 1), ''); // 区域数据 - 空白行
            foreach ($allRegionData as $region) {
                $salesPrice = 0;
                $businessPrice = 0;
                $productNum = 0;
                foreach ($data['area_list'] as $regionData) {
                    if ($region['area_id'] == $regionData['area_id']) {
                        $salesPrice = $regionData['area_sale_amount'];
                        $businessPrice = $regionData['area_business_amount'];
                        $productNum = $regionData['area_product_num'];
                    }
                }
                $sheet->setCellValue($columnLetter . ($index = $index + 2), $salesPrice); //总销售额
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $businessPrice); //营业收入
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $productNum); //商品数量
            }
            // ------订单数据--------
            $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['total_order_num']); //合计订单数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['min_order_amount']); //最小订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['max_order_amount']); //最大订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['avg_order_amount']); //平均订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['total_desk_num']); //桌数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['total_meal_num']); //人数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['min_desk_order_amount']); //最小订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['max_desk_order_amount']); //最大订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['avg_desk_order_amount']); //平均订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['total_instant_order_num']); //订单数
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['min_instant_order_amount']); //最小订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['max_instant_order_amount']); //最大订单金额
            $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['avg_instant_order_amount']); //平均订单金额
            // v2.3.0新增 外送订单 
            if (request()->licenses['is_open_delivery'] == 1) {
                $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['total_takeout_order_num']); // 外送订单数
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['min_takeout_order_amount']); // 外送最小订单金额
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['max_takeout_order_amount']); // 外送最大订单金额
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['avg_takeout_order_amount']); // 外送平均订单金额
            }
            // v2.15.0新增 Grab 数据
            if ($isOpenGrabDelivery) {
                $sheet->setCellValue($columnLetter . ($index = $index + 2), $data['total_grab_order_num']); // Grab订单数
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['min_grab_order_amount']); // Grab最小订单金额
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['max_grab_order_amount']); // Grab最大订单金额
                $sheet->setCellValue($columnLetter . ($index = $index + 1), $data['avg_grab_order_amount']); // Grab平均订单金额
            }
            // 支付数据
            // 支付金额应该从 $payRow + 1 开始（与支付方式名称的行号对应）
            $payColumnIndex = $payRow + 1;
            foreach ($paymentType as $value) {
                $totalPaymentAmount = 0;
                foreach ($data['payment_list'] as $payment) {
                    if ($value['payment_code'] == $payment['payment_code']) {
                        $totalPaymentAmount = $payment['total_payment_amount'];
                        break; // 找到匹配的支付方式后跳出循环
                    }
                }
                $sheet->setCellValue($columnLetter . $payColumnIndex, $totalPaymentAmount);
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
