<?php
use think\migration\Migrator;
use think\migration\db\Column;

class AddSourceToProductFlavor extends Migrator
{
    /**
     * 为商品规格表添加来源标识字段
     * 用于追踪外卖平台商品规格，避免重复创建
     */
    public function change()
    {
        // 1. 外卖商品表添加 grab_product_id 字段
        if ($this->hasTable('product_package_takeout')) {
            $table = $this->table('product_package_takeout');
            
            if (!$table->hasColumn('grab_product_id')) {
                $table->addColumn('grab_product_id', 'string', ['limit' => 500, 'default' => '', 'comment' => 'Grab商品ID（用于去重）', 'after' => 'image_file_uuid']);
            }
            
            // 添加唯一索引（company_uuid + grab_product_id 确保不重复）
            if (!$table->hasIndexByName('idx_grab_product_id')) {
                $table->addIndex(['grab_product_id'], ['name' => 'idx_grab_product_id']);
            }
            
            $table->update();
        }

        // 2. 外卖规格表添加 grab_modifier_id 字段
        if ($this->hasTable('product_bom_takeout')) {
            $table = $this->table('product_bom_takeout');
            
            if (!$table->hasColumn('grab_modifier_id')) {
                $table->addColumn('grab_modifier_id', 'string', ['limit' => 500, 'default' => '', 'comment' => 'Grab修饰符ID（规格/属性/加料）', 'after' => 'product_package_takeout_uuid']);
            }
            
            if (!$table->hasIndexByName('idx_grab_modifier_id')) {
                $table->addIndex(['grab_modifier_id'], ['name' => 'idx_grab_modifier_id']);
            }
            
            $table->update();
        }

        // 分类表添加 source 字段标记来源
        if ($this->hasTable('product_category')) {
            $table = $this->table('product_category');
            
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'string', ['limit' => 50, 'default' => '', 'comment' => '来源标记: grab, manual等', 'after' => 'name']);
            }
            
            if (!$table->hasColumn('source_id')) {
                $table->addColumn('source_id', 'string', ['limit' => 500, 'default' => '', 'comment' => '来源平台的分类ID', 'after' => 'source']);
            }
            
            if (!$table->hasIndexByName('idx_source_id')) {
                $table->addIndex(['source', 'source_id'], ['name' => 'idx_source_id']);
            }
            
            $table->update();
        }

        // 商品规格表添加 source 和 source_id 字段
        if ($this->hasTable('product_flavor')) {
            $table = $this->table('product_flavor');
            
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'string', [
                    'limit' => 50, 
                    'default' => '', 
                    'comment' => '来源标记(grab/manual等)', 
                    'after' => 'name'
                ]);
            }
            
            if (!$table->hasColumn('source_id')) {
                $table->addColumn('source_id', 'string', [
                    'limit' => 500, 
                    'default' => '', 
                    'comment' => '来源平台的规格ID', 
                    'after' => 'source'
                ]);
            }
            
            // 添加索引以优化查询性能
            if (!$table->hasIndexByName('idx_source_id')) {
                $table->addIndex(['source_id'], ['name' => 'idx_source_id']);
            }
            
            // 添加联合索引用于查询是否已存在该平台的规格
            if (!$table->hasIndexByName('idx_source_source_id')) {
                $table->addIndex(['source', 'source_id'], ['name' => 'idx_source_source_id']);
            }
            
            $table->update();
        }

        // 商品单位表添加 source 和 source_id 字段
        if ($this->hasTable('product_unit')) {
            $table = $this->table('product_unit');
            
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'string', [
                    'limit' => 50, 
                    'default' => '', 
                    'comment' => '来源标记(grab/manual等)', 
                    'after' => 'name'
                ]);
            }
            
            if (!$table->hasColumn('source_id')) {
                $table->addColumn('source_id', 'string', [
                    'limit' => 191, 
                    'default' => '', 
                    'comment' => '来源平台的单位ID', 
                    'after' => 'source'
                ]);
            }
            
            // 添加索引以优化查询性能
            if (!$table->hasIndexByName('idx_unit_source_id')) {
                $table->addIndex(['source_id'], ['name' => 'idx_unit_source_id']);
            }
            
            // 添加联合索引用于查询是否已存在该平台的单位
            if (!$table->hasIndexByName('idx_unit_source_source_id')) {
                $table->addIndex(['source', 'source_id'], ['name' => 'idx_unit_source_source_id']);
            }
            
            $table->update();
        }

        // 商品属性组表添加 source 和 source_id 字段
        if ($this->hasTable('product_attribute_group')) {
            $table = $this->table('product_attribute_group');
            
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'string', [
                    'limit' => 50, 
                    'default' => '', 
                    'comment' => '来源标记(grab/manual等)', 
                    'after' => 'name'
                ]);
            }
            
            if (!$table->hasColumn('source_id')) {
                $table->addColumn('source_id', 'string', [
                    'limit' => 500, 
                    'default' => '', 
                    'comment' => '来源平台的属性组ID', 
                    'after' => 'source'
                ]);
            }
            
            // 添加索引以优化查询性能
            if (!$table->hasIndexByName('idx_attr_group_source_id')) {
                $table->addIndex(['source_id'], ['name' => 'idx_attr_group_source_id']);
            }
            
            // 添加联合索引用于查询是否已存在该平台的属性组
            if (!$table->hasIndexByName('idx_attr_group_source_source_id')) {
                $table->addIndex(['source', 'source_id'], ['name' => 'idx_attr_group_source_source_id']);
            }
            
            $table->update();
        }

        // 商品属性表添加 source 和 source_id 字段
        if ($this->hasTable('product_attribute')) {
            $table = $this->table('product_attribute');
            
            if (!$table->hasColumn('source')) {
                $table->addColumn('source', 'string', [
                    'limit' => 50, 
                    'default' => '', 
                    'comment' => '来源标记(grab/manual等)', 
                    'after' => 'name'
                ]);
            }
            
            if (!$table->hasColumn('source_id')) {
                $table->addColumn('source_id', 'string', [
                    'limit' => 500, 
                    'default' => '', 
                    'comment' => '来源平台的属性ID', 
                    'after' => 'source'
                ]);
            }
            
            if (!$table->hasColumn('product_attribute_group_uuid')) {
                $table->addColumn('product_attribute_group_uuid', 'biginteger', [
                    'signed' => false,
                    'default' => 0, 
                    'comment' => '属性组UUID（用于关联查询）', 
                    'after' => 'attribute_group_uuid'
                ]);
            }
            
            // 添加索引以优化查询性能
            if (!$table->hasIndexByName('idx_attr_source_id')) {
                $table->addIndex(['source_id'], ['name' => 'idx_attr_source_id']);
            }
            
            // 添加联合索引用于查询是否已存在该平台的属性
            if (!$table->hasIndexByName('idx_attr_group_source')) {
                $table->addIndex(['product_attribute_group_uuid', 'source', 'source_id'], ['name' => 'idx_attr_group_source']);
            }
            
            $table->update();
        }
    }
}

