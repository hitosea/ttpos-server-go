<?php

namespace app\shop\service\order;

use app\common\library\helper;
use PhpOffice\PhpSpreadsheet\IOFactory;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

/**
 * 订单导出服务类
 */
class UserRechargeExportService
{
    /**
     * 订单导出
     */
    public function orderList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽设置
        $sheet->getColumnDimension('B')->setWidth(20);
        $sheet->getColumnDimension('E')->setWidth(30);
        $sheet->getColumnDimension('M')->setWidth(20);

        //设置工作表标题名称
        $sheet->setTitle(__('充值订单'));

        // 设置表头
        $sheet->setCellValue('A1', __('订单类型'));
        $sheet->setCellValue('B1', __('充值金额'));
        $sheet->setCellValue('C1', __('赠送金额'));
        $sheet->setCellValue('D1', __('赠送积分'));
        $sheet->setCellValue('E1', __('订单号'));
        $sheet->setCellValue('F1', __('状态'));
        $sheet->setCellValue('G1', __('支付时间'));
        $sheet->setCellValue('H1', __('订单金额'));
        $sheet->setCellValue('I1', __('实付金额'));
        $sheet->setCellValue('J1', __('退款金额'));
        $sheet->setCellValue('K1', __('会员'));
        $sheet->setCellValue('L1', __('支付方式'));
        $sheet->setCellValue('M1', __('收银员'));

        //填充数据
        $index = 0;
        foreach ($list as $order) {
            $payType = '';
            foreach ($order['payment_methods'] as $pay) {
                $payType .= $pay . '、';
            }
            $order['order_status_text'] = [
                0 => __('待付款'),
                1 => __('已完成'),
                2 => __('已取消')
            ][$order['status']];
            $order['pay_time_text'] = $order['payment_time'] ? date('Y-m-d H:i:s', $order['payment_time']) : '';
            $sheet->setCellValue('A' . ($index + 2), __('充值订单')); //订单类型
            $sheet->setCellValue('B' . ($index + 2), $order['recharge_amount']); //充值金额
            $sheet->setCellValue('C' . ($index + 2), $order['gift_amount']); //赠送金额
            $sheet->setCellValue('D' . ($index + 2), $order['gift_point']); //赠送积分
            $sheet->setCellValue('E' . ($index + 2), "\t" . $order['order_no'] . "\t"); //订单号
            $sheet->setCellValue('F' . ($index + 2), $order['order_status_text']); //状态
            $sheet->setCellValue('G' . ($index + 2), $order['pay_time_text']); //支付时间
            $sheet->setCellValue('H' . ($index + 2), $order['recharge_amount']); //订单金额
            $sheet->setCellValue('I' . ($index + 2), Helper::number2($order['amount'])); //实付金额
            $sheet->setCellValue('J' . ($index + 2), $order['refund_money'] ?? 0); //退款金额
            $sheet->setCellValue('K' . ($index + 2), $order['member_uuid'] ? __('会员') . 'ID' . '(' . $order['member_uuid'] .  ')' : ''); //会员
            $sheet->setCellValue('L' . ($index + 2), trim($payType, '、')); //支付方式
            $sheet->setCellValue('M' . ($index + 2), $order['cashier'] ? $order['cashier']['real_name'] : ''); //收银员

            $index++;
        }

        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('充值订单数据') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }
}
