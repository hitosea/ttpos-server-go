<?php
/**
 * 添加分批类型颜色到设置表
 */

use think\migration\Migrator;

class AddBatchColorToSettingTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-abstractmigration-class
     *
     * The following commands can be used in this method and Phinx will
     * automatically reverse them when rolling back:
     *
     *    createTable
     *    renameTable
     *    addColumn
     *    renameColumn
     *    addIndex
     *    addForeignKey
     *
     * Remember to call "create()" or "update()" and NOT "save()" when working
     * with the Table class.
     */
    public function change()
    {
    
        // 检查并插入分批类型颜色设置（只有当key不存在时才插入）
        $existing = $this->fetchRow("SELECT `key` FROM `ttpos_setting` WHERE `key` = 'batch_color'");
        if (!$existing) {
            $this->execute("INSERT INTO `ttpos_setting` (`key`, `describe`, `values`, `create_time`, `update_time`) VALUES ('batch_color', '分批类型颜色', '[\"#FF585B\", \"#FC0169\", \"#FF9900\", \"#BC3BBB\", \"#7A55D4\", \"#97B92D\", \"#006E5E\", \"#C18000\", \"#8C5A3F\"]', " . time() . ", " . time() . ")");
        }
    }
}
