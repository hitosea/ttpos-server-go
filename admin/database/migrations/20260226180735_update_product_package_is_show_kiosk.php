<?php
use think\migration\Migrator;

class UpdateProductPackageIsShowKiosk extends Migrator
{
    /**
     * 将商品包表所有记录的 is_show_kiosk 设置为 1（在自助点餐机显示）
     */
    public function change()
    {
        $this->execute("
            UPDATE ttpos_product_package
            SET is_show_kiosk = 1,
                update_time = UNIX_TIMESTAMP()
            WHERE delete_time = 0
        ");
    }
}
