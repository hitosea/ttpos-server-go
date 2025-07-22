<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddVisitorFieldsToMember extends Migrator
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
        if ($this->hasTable('member')) {
            // 修改会员表，添加游客相关字段
            $table = $this->table('member');
            
            // 添加是否游客字段
            if (!$table->hasColumn('is_visitor')) {
                $table->addColumn('is_visitor', 'integer', [
                    'default' => 0, 
                    'null' => false, 
                    'comment' => '是否游客,0-否 1-是',
                    'after' => 'phone'
                ]);
            }
            
            // 添加设备ID字段
            if (!$table->hasColumn('device_id')) {
                $table->addColumn('device_id', 'string', [
                    'limit' => 255, 
                    'default' => '', 
                    'null' => false, 
                    'comment' => '设备ID,用于标识游客',
                    'after' => 'is_visitor'
                ]);
                
                // 添加设备ID索引
                $table->addIndex(['device_id'], [
                    'name' => 'idx_device_id',
                    'unique' => false
                ]);
            }
            
            $table->update();
        }
    }
} 