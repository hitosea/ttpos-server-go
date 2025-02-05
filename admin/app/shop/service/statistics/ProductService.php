<?php

namespace app\shop\service\statistics;

use app\common\library\helper;
use PhpOffice\PhpSpreadsheet\IOFactory;
use PhpOffice\PhpSpreadsheet\Style\Fill;
use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;
use PhpOffice\PhpSpreadsheet\Cell\Coordinate;

/**
 * 商品销售统计
 */
class ProductService
{

    /**
     * 列表导出
     */
    public function productExport($list)
    {
        $spreadsheet = new Spreadsheet();
        $sheet = $spreadsheet->getActiveSheet();

        //列宽
        $sheet->getColumnDimension('B')->setWidth(30);
        $sheet->getColumnDimension('P')->setWidth(30);

        //设置工作表标题名称
        $sheet->setTitle(__('商品销售统计'));

        $sheet->setCellValue('A1', __('序号'));
        $sheet->setCellValue('B1', __('商品名称'));
        $sheet->setCellValue('C1', __('商品分类'));
        $sheet->setCellValue('D1', __('销售数量'));
        $sheet->setCellValue('E1', __('原价销售额'));
        $sheet->setCellValue('F1', __('赠菜'));
        $sheet->setCellValue('G1', __('实际销售额'));
        $sheet->setCellValue('H1', __('营业收入'));

        //填充数据
        $index = 0;
        foreach ($list as $key => $product) {
            $sheet->setCellValue('A' . ($index + 2), $key + 1);
            $sheet->setCellValue('B' . ($index + 2), $product['product_name_text']);
            $sheet->setCellValue('C' . ($index + 2), $product['category_name_text']);
            $sheet->setCellValue('D' . ($index + 2), $product['product_num']);
            $sheet->setCellValue('E' . ($index + 2), $product['sales_price']);
            $sheet->setCellValue('F' . ($index + 2), $product['free_product_num']);
            $sheet->setCellValue('G' . ($index + 2), $product['received_price']);
            $sheet->setCellValue('H' . ($index + 2), $product['business_price']);

            $index++;
        }

        // 保存文件
        $writer = new Xlsx($spreadsheet);
        $filename = __('商品销售统计') . '.xlsx';
        header('Content-Type: application/vnd.ms-excel');
        header('Content-Disposition: attachment;filename="' . $filename . '"');
        header('Cache-Control: max-age=0');
        $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
        $writer->save('php://output');
        exit();
    }
}
