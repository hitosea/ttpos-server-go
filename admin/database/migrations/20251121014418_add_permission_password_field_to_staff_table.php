<?php
/**
 * 为员工表添加权限密码字段
 * 
 * 任务: story-shop-staff-permission-password Phase 1.1
 * 需求: R1.1
 */

use think\migration\Migrator;

class AddPermissionPasswordFieldToStaffTable extends Migrator
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
        // 检查表是否存在
        if ($this->hasTable('staff')) {
            $table = $this->table('staff');
            
            // 检查 permission_password 字段是否不存在，如果不存在则添加
            if (!$table->hasColumn('permission_password')) {
                // 计算默认值（666888）的加密值：md5(md5('666888') . 'jjjshop_salt_2020')
                $defaultPassword = '666888';
                $encryptedDefault = md5(md5($defaultPassword) . 'jjjshop_salt_2020');
                
                // 字段不允许为空，默认值为空字符串
                $table->addColumn('permission_password', 'string', ['limit' => 255, 'null' => false, 'default' => '', 'comment' => '权限密码（加密存储），默认值666888', 'after' => 'password']);
                
                $table->update();
                
                // 为已存在的员工记录设置默认权限密码（如果字段为空）
                $this->execute("UPDATE `ttpos_staff` SET `permission_password` = '" . $encryptedDefault . "' WHERE (`permission_password` IS NULL OR `permission_password` = '') AND `delete_time` = 0");
            }
        }
    }
}

