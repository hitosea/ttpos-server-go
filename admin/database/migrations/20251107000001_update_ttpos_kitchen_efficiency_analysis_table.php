<?php

use think\migration\Migrator;
use think\migration\worker\Incr;
use Phinx\Db\Adapter\MysqlAdapter;

class UpdateTTPOSKitchenEfficiencyAnalysisTable extends Migrator
{
    /**
     * Change Method.
     *
     * Write your reversible migrations using this method.
     *
     * More information on writing migrations is available here:
     * http://docs.phinx.org/en/latest/migrations.html#the-change-method
     *
     * @return void
     */
    public function change()
    {
        $table = $this->table('kitchen_efficiency_analysis');

        // 定义表字段
        if ($table->hasColumn('min')) {
            $table->changeColumn('min', 'decimal', ['comment' => '最短出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        }
        if ($table->hasColumn('max')) {
            $table->changeColumn('max', 'decimal', ['comment' => '最长出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        }
        if ($table->hasColumn('avg')) {
            $table->changeColumn('avg', 'decimal', ['comment' => '平均出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        }
        if ($table->hasColumn('total')) {
            $table->changeColumn('total', 'decimal', ['comment' => '总出品时长', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        }
        if ($table->hasColumn('count')) {
            $table->changeColumn('count', 'decimal', ['comment' => '出品次数', 'default' => 0, 'signed' => false, 'precision' => 22, 'scale' => 4]);
        }
        $table->update();
    }
    
}
