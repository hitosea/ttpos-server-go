<?php

use think\facade\Db;
use think\facade\Config;
use think\migration\Migrator;

class AddIsSuperToCompanyStaff extends Migrator
{

    // 迁移目标
    const TARGET = 'main';

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
        $table = $this->table('company_staff');
        if (!$table->hasColumn('is_super')) {
            $table->addColumn('is_super', 'integer', ['default' => 0, 'comment' => '是否超级管理员', 'after' => 'phone']);
            $table->update();
        }
        // 遍历修复旧数据
        $default = config('database.default');
        $config = $oldConfig = config('database');
        $mysql = $config['connections'][$default];
        foreach (Db::name('company')->where('delete_time', 0)->column('uuid') as $appid) {
            $mysql['database'] = 'shop' . $appid;
            $mysql['username'] = env('DB_USERNAME');
            $mysql['password'] = env('DB_PASSWORD');
            $config['connections'][$default] = $mysql;
            Config::set($config, 'database');
            $dbs = Db::connect(Db::getConfig('default'), true);
            $superUuid = $dbs->name('staff')->where('company_uuid', $appid)->where('is_super', 1)->value('uuid');
            // 恢复
            $config['connections'][$default] = $oldConfig['connections'][$default];
            Config::set($config, 'database');
            $dbs = Db::connect(Db::getConfig('default'), true);
            // 更新
            Db::name('company_staff')->where('company_uuid', $appid)->where('uuid', $superUuid)->update(['is_super' => 1]);
        }
    }
}
