<?php

namespace app\shop\service\order;

use app\common\library\helper;
use PhpOffice\PhpSpreadsheet\IOFactory;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

/**
 * 订单导出服务类
 */
class MemberOrderExportService
{
    /**
     * 订单导出
     */
    public function export($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽设置
        // $sheet->getColumnDimension('B')->setWidth(20);
        // $sheet->getColumnDimension('E')->setWidth(30);
        // $sheet->getColumnDimension('M')->setWidth(20);

        //设置工作表标题名称
        $sheet->setTitle(__('外送订单'));

        // 设置表头
        $sheet->setCellValue('A1', __('外卖序号'));
        $sheet->setCellValue('B1', __('订单号'));
        $sheet->setCellValue('C1', __('状态'));
        $sheet->setCellValue('D1', __('下单时间'));
        $sheet->setCellValue('E1', __('支付时间'));
        $sheet->setCellValue('F1', __('订单金额'));
        $sheet->setCellValue('G1', __('实付金额'));
        $sheet->setCellValue('H1', __('用户信息'));
        $sheet->setCellValue('I1', __('配送费'));
        $sheet->setCellValue('J1', __('支付方式'));


        $statusMap = [
            0 => __('选购中'),
            1 => __('待付款'),
            2 => __('待商家接单'),
            3 => __('商家备餐中'),
            4 => __('待骑手接单'),
            5 => __('骑手正在赶往商家'),
            6 => __('骑手配送中'),
            7 => __('已完成'),
            8 => __('已取消'),
        ];

        //填充数据
        $index = 0;
        foreach ($list as $order) { 

            $userInfo = $order['contact']['name'] ?? '';
            if ($order['contact']['phone']) {
                $userInfo .= '(' . $order['contact']['phone'] . ')';
            }

            $sheet->setCellValue('A' . ($index + 2), $order['serial_number']); // 外卖序号
            $sheet->setCellValue('B' . ($index + 2), $order['order_no']); // 订单号
            $sheet->setCellValue('C' . ($index + 2), $statusMap[$order['status']]); // 状态
            $sheet->setCellValue('D' . ($index + 2), date('Y-m-d H:i:s', $order['create_time'])); // 下单时间
            $sheet->setCellValue('E' . ($index + 2), date('Y-m-d H:i:s', $order['pay_time'])); // 支付时间
            $sheet->setCellValue('F' . ($index + 2), $order['origin_amount']); // 订单金额
            $sheet->setCellValue('G' . ($index + 2), $order['pay_amount']); // 实付金额 
            $sheet->setCellValue('H' . ($index + 2), $userInfo); // 用户信息
            $sheet->setCellValue('I' . ($index + 2), $order['delivery_fee']); // 配送费
            $sheet->setCellValue('J' . ($index + 2), $order['pay_type']); // 支付方式

            $index++;
        }

        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('外送订单数据') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }
}
