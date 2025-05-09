<?php

use think\migration\Migrator;
use think\migration\db\Column;

class AddIndexToStatistics extends Migrator
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
        try {
            $table = $this->table('statistics_sale');
            if (!$table->hasIndex('idx_duty_no')) {
                $table->addIndex(['duty_no'], ['name' => 'idx_duty_no']);
            }
            if (!$table->hasIndex('idx_desk_uuid')) {
                $table->addIndex(['desk_uuid'], ['name' => 'idx_desk_uuid']);
            }
            if (!$table->hasIndex('idx_complete_time')) {
                $table->addIndex(['complete_time'], ['name' => 'idx_complete_time']);
            }
            $table->update();

            // 
            $table = $this->table('statistics_product');
            if (!$table->hasIndex('idx_duty_no')) {
                $table->addIndex(['duty_no'], ['name' => 'idx_duty_no']);
            }
            if (!$table->hasIndex('idx_desk_uuid')) {
                $table->addIndex(['desk_uuid'], ['name' => 'idx_desk_uuid']);
            }
            if (!$table->hasIndex('idx_complete_time')) {
                $table->addIndex(['complete_time'], ['name' => 'idx_complete_time']);
            }
            $table->update();

            // 
            $table = $this->table('statistics_payment');
            if (!$table->hasIndex('idx_duty_no')) {
                $table->addIndex(['duty_no'], ['name' => 'idx_duty_no']);
            }
            if (!$table->hasIndex('idx_desk_uuid')) {
                $table->addIndex(['desk_uuid'], ['name' => 'idx_desk_uuid']);
            }
            if (!$table->hasIndex('idx_complete_time')) {
                $table->addIndex(['complete_time'], ['name' => 'idx_complete_time']);
            }
            $table->update();

            // 
            $table = $this->table('statistics_member_payment');
            if (!$table->hasIndex('idx_duty_no')) {
                $table->addIndex(['duty_no'], ['name' => 'idx_duty_no']);
            }
            if (!$table->hasIndex('idx_payment_method_uuid')) {
                $table->addIndex(['payment_method_uuid'], ['name' => 'idx_payment_method_uuid']);
            }
            if (!$table->hasIndex('idx_complete_time')) {
                $table->addIndex(['complete_time'], ['name' => 'idx_complete_time']);
            }
            $table->update();

            // 
            $table = $this->table('statistics_member');
            if (!$table->hasIndex('idx_duty_no')) {
                $table->addIndex(['duty_no'], ['name' => 'idx_duty_no']);
            }
            if (!$table->hasIndex('idx_complete_time')) {
                $table->addIndex(['complete_time'], ['name' => 'idx_complete_time']);
            }
            $table->update();
        } catch (\Exception $e) {
            trace($e->getMessage(), 'error');
        }
    }
}
