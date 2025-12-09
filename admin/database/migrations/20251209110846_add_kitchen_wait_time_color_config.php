<?php

use think\facade\Db;
use think\migration\Migrator;
use think\migration\db\Column;

class AddKitchenWaitTimeColorConfig extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     */
    public function change()
    {
        $db = Db::connect(Db::getConfig('default'), true);

        // 查询所有门店的厨显设置
        $settings = $db->name('setting')->where('key', 'kitchen')->select();

        foreach ($settings as $setting) {
            $values = json_decode($setting['values'], true);
            if (!$values) {
                continue;
            }

            // 如果已经存在 wait_time_color_ranges，跳过
            if (isset($values['wait_time_color_ranges']) && is_array($values['wait_time_color_ranges']) && count($values['wait_time_color_ranges']) > 0) {
                continue;
            }

            // 初始化默认配置
            $waitTimeColorRanges = [
                ['minute' => 0, 'color' => '#100A05'], // 第一区间：0分钟（黑色）
            ];

            // 从旧格式 wait_color 读取配置
            $waitColor = $values['wait_color'] ?? [];
            if (is_array($waitColor) && count($waitColor) > 0) {
                // 第二区间：10分钟
                $secondColor = $this->convertColorToRgb($waitColor[0] ?? 'yellow');
                $waitTimeColorRanges[] = ['minute' => 10, 'color' => $secondColor];

                // 第三区间：20分钟
                $thirdColor = $this->convertColorToRgb($waitColor[1] ?? 'red');
                $waitTimeColorRanges[] = ['minute' => 20, 'color' => $thirdColor];
            } else {
                // 如果没有旧配置，使用默认值
                $waitTimeColorRanges[] = ['minute' => 10, 'color' => '#FFBE00']; // 默认黄色
                $waitTimeColorRanges[] = ['minute' => 20, 'color' => '#E50028']; // 默认红色
            }

            // 更新 values
            $values['wait_time_color_ranges'] = $waitTimeColorRanges;
            $updatedValues = json_encode($values, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);

            // 更新数据库
            $db->name('setting')
                ->where('key', 'kitchen')
                ->update(['values' => $updatedValues, 'update_time' => time()]);
        }
    }

    /**
     * 转换颜色格式：red/yellow → RGB
     *
     * @param string $color 颜色值（red 或 yellow）
     * @return string RGB 格式颜色值
     */
    private function convertColorToRgb($color)
    {
        $colorMap = [
            'red' => '#E50028',
            'yellow' => '#FFBE00',
        ];

        $colorLower = strtolower(trim($color));
        return $colorMap[$colorLower] ?? '#FFBE00'; // 默认黄色
    }
}
