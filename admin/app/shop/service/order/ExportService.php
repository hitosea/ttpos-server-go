<?php

namespace app\shop\service\order;

use app\common\library\helper;
use PhpOffice\PhpSpreadsheet\IOFactory;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

/**
 * 订单导出服务类
 */
class ExportService
{
    /**
     * 订单导出
     */
    public function orderList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽
        $sheet->getColumnDimension('B')->setWidth(30);
        $sheet->getColumnDimension('P')->setWidth(30);

        //设置工作表标题名称
        $sheet->setTitle(__('订单明细'));

        $sheet->setCellValue('A1', __('订单ID'));
        $sheet->setCellValue('B1', __('订单类型'));
        $sheet->setCellValue('C1', __('商品信息'));
        $sheet->setCellValue('D1', __('桌号/序号'));
        $sheet->setCellValue('E1', __('订单号'));
        $sheet->setCellValue('F1', __('状态'));
        $sheet->setCellValue('G1', __('支付时间'));
        $sheet->setCellValue('H1', __('订单金额'));
        $sheet->setCellValue('I1', __('服务费'));
        $sheet->setCellValue('J1', __('优惠折扣'));
        $sheet->setCellValue('K1', __('会员折扣'));
        $sheet->setCellValue('L1', __('实付金额'));
        $sheet->setCellValue('M1', __('退款金额'));
        $sheet->setCellValue('N1', __('会员'));
        $sheet->setCellValue('O1', __('支付方式'));
        $sheet->setCellValue('P1', __('用餐方式'));
        $sheet->setCellValue('Q1', __('收银员'));

        //填充数据
        $index = 0;
        foreach ($list as $order) {
            $sheet->setCellValue('A' . ($index + 2), $order['order_id']);
            $sheet->setCellValue('B' . ($index + 2), $order['bill_type']);
            $sheet->setCellValue('C' . ($index + 2), $this->filterProductInfo($order));
            $sheet->setCellValue('D' . ($index + 2), $order['serial_no'] ?? '');
            $sheet->setCellValue('E' . ($index + 2), "\t" . $order['order_no'] . "\t");
            $sheet->setCellValue('F' . ($index + 2), $order['status_text']);
            $sheet->setCellValue('G' . ($index + 2), $this->filterTime($order['finish_time']));
            $sheet->setCellValue('H' . ($index + 2), $order['order_amount']);
            $sheet->setCellValue('I' . ($index + 2), $order['service_fee']);
            $sheet->setCellValue('J' . ($index + 2), $order['discount_fee']);
            $sheet->setCellValue('K' . ($index + 2), $order['member_fee']);
            $sheet->setCellValue('L' . ($index + 2), Helper::number2($order['payment_amount']));
            $sheet->setCellValue('M' . ($index + 2), $order['refund_amount']);
            $sheet->setCellValue('N' . ($index + 2), $order['member_names'] ?? '');
            $sheet->setCellValue('O' . ($index + 2), rtrim($order['pay_type_name'], '+'));
            $sheet->setCellValue('P' . ($index + 2), ($order['dining_method'] ?? 0) == 1 ? __('打包带走') : __('店内就餐') );
            $sheet->setCellValue('Q' . ($index + 2), $order['cashier_name'] ?? '');

            $index++;
        }

        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('用餐订单数据') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 订单导出
     */
    public function deliverList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽
        $sheet->getColumnDimension('A')->setWidth(20);
        $sheet->getColumnDimension('B')->setWidth(10);

        //设置工作表标题名称
        $sheet->setTitle(__('订单明细'));

        $sheet->setCellValue('A1', __('订单号'));
        $sheet->setCellValue('B1', __('订单金额'));
        $sheet->setCellValue('C1', __('订单状态'));
        $sheet->setCellValue('D1', __('配送方式'));
        $sheet->setCellValue('E1', __('配送费'));
        $sheet->setCellValue('F1', __('配送状态'));
        $sheet->setCellValue('G1', __('配送时间'));
        $sheet->setCellValue('H1', __('送达时间'));
        $sheet->setCellValue('I1', __('收货人姓名'));
        $sheet->setCellValue('J1', __('联系电话'));
        $sheet->setCellValue('L1', __('收货人地址'));
        $sheet->setCellValue('L1', __('配送员'));
        $sheet->setCellValue('M1', __('配送员电话'));

        //填充数据
        $index = 0;
        foreach ($list as $order) {
            $address = $order['address'];
            $sheet->setCellValue('A' . ($index + 2), "\t" . $order['order_no'] . "\t");
            $sheet->setCellValue('B' . ($index + 2), $order['orders']['order_price']);
            $sheet->setCellValue('C' . ($index + 2), $order['orders']['order_status']['text']);
            $sheet->setCellValue('D' . ($index + 2), $order['deliver_source_text']);
            $sheet->setCellValue('E' . ($index + 2), $order['price']);
            $sheet->setCellValue('F' . ($index + 2), $order['deliver_status_text']);
            $sheet->setCellValue('G' . ($index + 2), $order['create_time']);
            $sheet->setCellValue('H' . ($index + 2), $this->filterTime($order['deliver_time']));
            $sheet->setCellValue('I' . ($index + 2), $order['orders']['address']['name']);
            $sheet->setCellValue('J' . ($index + 2), "\t" . $order['orders']['address']['phone'] . "\t");
            $sheet->setCellValue('K' . ($index + 2), $address ? $address->getFullAddress() : '');
            $sheet->setCellValue('L' . ($index + 2), $order['linkman']);
            $sheet->setCellValue('M' . ($index + 2), $order['phone']);
            $index++;
        }

        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('订单配送信息') . '-' . date('YmdHis') . '.xlsx';

        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 积分订单导出
     */
    public function pointsList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽
        $sheet->getColumnDimension('B')->setWidth(30);
        $sheet->getColumnDimension('P')->setWidth(30);

        //设置工作表标题名称
        $sheet->setTitle(__('积分订单明细'));

        $sheet->setCellValue('A1', __('订单号'));
        $sheet->setCellValue('B1', __('商品信息'));
        $sheet->setCellValue('C1', __('订单总额'));
        $sheet->setCellValue('D1', __('兑换积分'));
        $sheet->setCellValue('E1', __('配送费'));
        $sheet->setCellValue('F1', __('支付方式'));
        $sheet->setCellValue('G1', __('下单时间'));
        $sheet->setCellValue('H1', __('会员'));
        $sheet->setCellValue('I1', __('配送方式'));
        $sheet->setCellValue('J1', __('自提门店'));
        $sheet->setCellValue('K1', __('门店电话'));
        $sheet->setCellValue('L1', __('门店地址'));
        $sheet->setCellValue('M1', __('收货人姓名'));
        $sheet->setCellValue('N1', __('联系电话'));
        $sheet->setCellValue('O1', __('收货人地址'));
        $sheet->setCellValue('P1', __('付款状态'));
        $sheet->setCellValue('Q1', __('付款时间'));
        $sheet->setCellValue('R1', __('核销时间'));
        $sheet->setCellValue('S1', __('订单状态'));
        $sheet->setCellValue('T1', __('支付交易号'));

        //填充数据
        $index = 0;
        foreach ($list as $order) {
            $address = $order['address'];
            $sheet->setCellValue('A' . ($index + 2), "\t" . $order['order_no'] . "\t");
            $sheet->setCellValue('B' . ($index + 2), $order['product_name']);
            $sheet->setCellValue('C' . ($index + 2), $order['pay_price']);
            $sheet->setCellValue('D' . ($index + 2), $order['points_num']);
            $sheet->setCellValue('E' . ($index + 2), $order['express_price']);
            $sheet->setCellValue('F' . ($index + 2), $order['pay_type']['text']);
            $sheet->setCellValue('G' . ($index + 2), $order['create_time']);
            $sheet->setCellValue('H' . ($index + 2), $order['user']['nickname']);
            $sheet->setCellValue('I' . ($index + 2), $order['delivery_type']['text']);
            $sheet->setCellValue('J' . ($index + 2), $order['store'] ? $order['store']['name'] : '');
            $sheet->setCellValue('K' . ($index + 2), $order['store'] ? "\t" . $order['store']['link_phone'] . "\t" : '');
            $sheet->setCellValue('L' . ($index + 2), $order['store'] ? $order['store']['address'] : '');
            $sheet->setCellValue('M' . ($index + 2), $order['address'] ? $order['address']['name'] : '');
            $sheet->setCellValue('N' . ($index + 2), $order['address'] ? "\t" . $order['address']['phone'] . "\t" : '');
            $sheet->setCellValue('O' . ($index + 2), $order['address'] ? $order['address']['detail'] : '');
            $sheet->setCellValue('P' . ($index + 2), $order['pay_status']['text']);
            $sheet->setCellValue('Q' . ($index + 2), $this->filterTime($order['pay_time']));
            $sheet->setCellValue('R' . ($index + 2), $this->filterTime($order['receipt_time']));
            $sheet->setCellValue('S' . ($index + 2), $order['state_text']);
            $sheet->setCellValue('T' . ($index + 2), $order['transaction_id']);
            $index++;
        }

        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('积分订单') . '-' . date('YmdHis') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 格式化商品信息
     */
    private function filterProductInfo($order)
    {
        $content = '';
        $key = 1;
        // 商品
        foreach ($order['product'] ?? [] as $product) {
            $total_price = Helper::number2($product['total_price']);
            // 商品名称
            $content .= $key . "." . __("商品名称") . "：{$product['name']}\n";
            // 商品规格
            if (!empty($product['attr_name'])) {
                $content .= "　" . __("商品规格") . "：{$product['attr_name']}\n";
            }
            // 购买数量
            $content .= "　" . __("购买数量") . "：{$product['num']}\n";
            // 商品总价
            $content .= "　" . __("商品总价") . "：{$total_price}\n\n";
            $key++;
        }
        return $content;
    }

    /**
     * 订单导出
     */
    public function financeOrderList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();
        //列宽
        $sheet->getColumnDimension('A')->setWidth(40);
        $sheet->getColumnDimension('B')->setWidth(30);
        $sheet->getColumnDimension('C')->setWidth(20);
        $sheet->getColumnDimension('D')->setWidth(20);
        $sheet->getColumnDimension('E')->setWidth(20);
        $sheet->getColumnDimension('F')->setWidth(20);
        $sheet->getColumnDimension('G')->setWidth(20);
        //设置工作表标题名称
        $sheet->setTitle(__('订单明细'));
        $sheet->setCellValue('A1', __('订单号'));
        $sheet->setCellValue('B1', __('订单来源'));
        $sheet->setCellValue('C1', __('应收金额'));
        $sheet->setCellValue('D1', __('优惠金额'));
        $sheet->setCellValue('E1', __('实付金额'));
        $sheet->setCellValue('F1', __('预计收入'));
        $sheet->setCellValue('G1', __('订单状态'));
        //填充数据
        $index = 0;
        foreach ($list as $order) {
            $sheet->setCellValue('A' . ($index + 2), "\t" . $order['order_no'] . "\t");
            $sheet->setCellValue('B' . ($index + 2), $order['order_type_text']);
            $sheet->setCellValue('C' . ($index + 2), $order['order_price']);
            $sheet->setCellValue('D' . ($index + 2), $order['order_price'] - $order['pay_price']);
            $sheet->setCellValue('E' . ($index + 2), $order['pay_price']);
            $sheet->setCellValue('F' . ($index + 2), $order['pay_price'] - $order['refund_money']);
            $sheet->setCellValue('G' . ($index + 2), $order['order_status']['text']);
            $index++;
        }
        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('订单') . '-' . date('YmdHis') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 订单导出
     */
    public function recordOrderList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();
        //列宽
        $sheet->getColumnDimension('A')->setWidth(40);
        $sheet->getColumnDimension('B')->setWidth(30);
        $sheet->getColumnDimension('C')->setWidth(20);
        $sheet->getColumnDimension('D')->setWidth(20);
        $sheet->getColumnDimension('E')->setWidth(20);
        $sheet->getColumnDimension('F')->setWidth(20);
        $sheet->getColumnDimension('G')->setWidth(20);
        $sheet->getColumnDimension('H')->setWidth(20);
        $sheet->getColumnDimension('I')->setWidth(20);
        //设置工作表标题名称
        $sheet->setTitle(__('订单交易明细'));
        $sheet->setCellValue('A1', __('订单号'));
        $sheet->setCellValue('B1', __('订单类型'));
        $sheet->setCellValue('C1', __('总金额'));
        $sheet->setCellValue('D1', __('优惠金额'));
        $sheet->setCellValue('E1', __('实付金额'));
        $sheet->setCellValue('F1', __('实际到账'));
        $sheet->setCellValue('G1', __('支付方式'));
        $sheet->setCellValue('H1', __('订单状态'));
        $sheet->setCellValue('I1', __('下单时间'));
        //填充数据
        $index = 0;
        foreach ($list as $order) {
            $sheet->setCellValue('A' . ($index + 2), "\t" . $order['order_no'] . "\t");
            $sheet->setCellValue('B' . ($index + 2), $order['order_type_text']);
            $sheet->setCellValue('C' . ($index + 2), $order['order_price']);
            $sheet->setCellValue('D' . ($index + 2), $order['order_price'] - $order['pay_price']);
            $sheet->setCellValue('E' . ($index + 2), $order['pay_price']);
            $sheet->setCellValue('F' . ($index + 2), $order['pay_price'] - $order['refund_money']);
            $sheet->setCellValue('G' . ($index + 2), $order['pay_type']['text']);
            $sheet->setCellValue('H' . ($index + 2), $order['order_status']['text']);
            $sheet->setCellValue('I' . ($index + 2), $order['create_time']);
            $index++;
        }
        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename =  __('订单') . '-' . date('YmdHis') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 订单导出
     */
    public function ProductRank($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();
        //列宽
        $sheet->getColumnDimension('A')->setWidth(40);
        $sheet->getColumnDimension('B')->setWidth(30);
        $sheet->getColumnDimension('C')->setWidth(20);
        $sheet->getColumnDimension('D')->setWidth(20);
        $sheet->getColumnDimension('E')->setWidth(20);
        //设置工作表标题名称
        $sheet->setTitle(__('商品销量明细'));
        $sheet->setCellValue('A1', __('排名'));
        $sheet->setCellValue('B1', __('商品名称'));
        $sheet->setCellValue('C1', __('商品价格'));
        $sheet->setCellValue('D1', __('销量'));
        $sheet->setCellValue('E1', __('销售额'));

        //填充数据
        $index = 0;
        foreach ($list as $key => $item) {
            $sheet->setCellValue('A' . ($index + 2), "\t" . ($key + 1) . "\t");
            $sheet->setCellValue('B' . ($index + 2), $item['product_name']);
            $sheet->setCellValue('C' . ($index + 2), $item['product_price']);
            $sheet->setCellValue('D' . ($index + 2), $item['total_num']);
            $sheet->setCellValue('E' . ($index + 2), $item['total_price']);
            $index++;
        }
        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('订单') . '-' . date('YmdHis') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 订单导出
     */
    public function groupOrder($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();
        //列宽
        $sheet->getColumnDimension('A')->setWidth(20);
        $sheet->getColumnDimension('B')->setWidth(30);

        $sheet->getColumnDimension('G')->setWidth(15);
        $sheet->getColumnDimension('J')->setWidth(20);
        $sheet->getColumnDimension('K')->setWidth(20);
        $sheet->getColumnDimension('L')->setWidth(20);
        $sheet->getColumnDimension('M')->setWidth(20);
        $sheet->getColumnDimension('N')->setWidth(20);
        //设置工作表标题名称
        $sheet->setTitle(__('团购订单明细'));
        $sheet->setCellValue('A1', __('订单号'));
        $sheet->setCellValue('B1', __('团购名称'));
        $sheet->setCellValue('C1', __('团购价格'));
        $sheet->setCellValue('D1', __('团购数量'));
        $sheet->setCellValue('E1', __('订单总额'));
        $sheet->setCellValue('F1', __('实付款金额'));
        $sheet->setCellValue('G1', __('支付方式'));
        $sheet->setCellValue('H1', __('下单时间'));
        $sheet->setCellValue('I1', __('会员'));
        $sheet->setCellValue('J1', __('会员电话'));
        $sheet->setCellValue('K1', __('付款时间'));
        $sheet->setCellValue('L1', __('核销时间'));
        $sheet->setCellValue('M1', __('订单状态'));
        $sheet->setCellValue('N1', __('微信支付交易号'));

        //填充数据
        $index = 0;
        foreach ($list as $key => $item) {
            $sheet->setCellValue('A' . ($index + 2), "\t" . $item['order_no'] . "\t");
            $sheet->setCellValue('B' . ($index + 2), $item['product'][0]['group_name']);
            $sheet->setCellValue('C' . ($index + 2), $item['product'][0]['group_price']);
            $sheet->setCellValue('D' . ($index + 2), $item['product'][0]['total_num']);
            $sheet->setCellValue('E' . ($index + 2), $item['total_price']);
            $sheet->setCellValue('F' . ($index + 2), $item['pay_price']);
            $sheet->setCellValue('G' . ($index + 2), $item['pay_type']['text']);
            $sheet->setCellValue('H' . ($index + 2), $item['create_time']);
            $sheet->setCellValue('I' . ($index + 2), $item['user']['nickname']);
            $sheet->setCellValue('J' . ($index + 2), $item['user']['mobile']);
            $sheet->setCellValue('K' . ($index + 2), $this->filterTime($item['pay_time']));
            $sheet->setCellValue('L' . ($index + 2), $this->filterTime($item['settled_time']));
            $sheet->setCellValue('M' . ($index + 2), $item['order_status']['text']);
            $sheet->setCellValue('N' . ($index + 2), $item['transaction_id']);
            $index++;
        }
        //保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('团购订单明细') . '-' . date('YmdHis') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 余额提现订单导出
     */
    public function userCashList($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽
        $sheet->getColumnDimension('I')->setWidth(50);

        //设置工作表标题名称
        $sheet->setTitle(__('余额提现明细'));

        $sheet->setCellValue('A1', 'ID');
        $sheet->setCellValue('B1', __('用户ID'));
        $sheet->setCellValue('C1', __('微信昵称'));
        $sheet->setCellValue('D1', __('手机号'));
        $sheet->setCellValue('E1', __('提现金额'));
        $sheet->setCellValue('F1', __('实际到账'));
        $sheet->setCellValue('G1', __('提现比例'));
        $sheet->setCellValue('H1', __('提现方式'));
        $sheet->setCellValue('I1', __('提现信息'));
        $sheet->setCellValue('J1', __('审核状态'));
        $sheet->setCellValue('K1', __('申请时间'));
        $sheet->setCellValue('L1', __('审核时间'));
        //填充数据
        $index = 0;
        foreach ($list as $cash) {
            $sheet->setCellValue('A' . ($index + 2), $cash['id']);
            $sheet->setCellValue('B' . ($index + 2), $cash['user_id']);
            $sheet->setCellValue('C' . ($index + 2), $cash['nickname']);
            $sheet->setCellValue('D' . ($index + 2), "\t" . $cash['mobile'] . "\t");
            $sheet->setCellValue('E' . ($index + 2), $cash['money']);
            $sheet->setCellValue('F' . ($index + 2), $cash['real_money']);
            $sheet->setCellValue('G' . ($index + 2), $cash['cash_ratio'] . '%');
            $sheet->setCellValue('H' . ($index + 2), $cash['pay_type']['text']);
            $sheet->setCellValue('I' . ($index + 2), $this->cashInfo($cash));
            $sheet->setCellValue('J' . ($index + 2), $cash['apply_status']['text']);
            $sheet->setCellValue('K' . ($index + 2), $cash['create_time']);
            $sheet->setCellValue('L' . ($index + 2), $cash['audit_time']);
            $index++;
        }
        //保存文件
        $filename = __('余额提现明细') . '-' . date('YmdHis') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }

    /**
     * 格式化提现信息
     */
    private function cashInfo($cash)
    {
        $content = '';
        if ($cash['pay_type']['value'] == 20) {
            $content .= __("支付宝姓名") . "：{$cash['alipay_name']}\n";
            $content .= __("支付宝账号") . "  ：{$cash['alipay_account']}\n";
        } elseif ($cash['pay_type']['value'] == 30) {
            $content .= __("银行名称") . "：{$cash['bank_name']}\n";
            $content .= __("开户名") . "  ：{$cash['bank_account']}\n";
            $content .= __("银行卡号") . "  ：{$cash['bank_card']}\n";
        }
        return $content;
    }

    /**
     * 日期值过滤
     */
    private function filterTime($value)
    {
        if (!$value) return '';
        return date('Y-m-d H:i:s', $value);
    }
}
