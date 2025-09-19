<?php

namespace app\common\model\erp;

use app\common\model\BaseModel;
use app\common\model\shop\User;
use think\model\concern\SoftDelete;

/**
 * 供应商模型
 */
class ErpSupplier extends BaseModel
{
    use SoftDelete;
    protected $name = 'supplier';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;

    // 定义全局的查询范围
    protected $globalScope = ['erp_code'];

    public function scopeErp_code($query)
    {
        $query->where('erp_code', '');
    }

    /**
     * 追加字段
     * @var string[]
     */
    protected $append = ['contact_person'];

    /**
     * 兼容字段
     */
    public function getContactPersonAttr($value, $data)
    {
        return $this->contact_name ?: '';
    }

    /**
     * 采购负责人
     */
    public function purchaser()
    {
        return $this->belongsTo(User::class, 'staff_uuid', 'uuid');
    }

    /**
     * 获取列表
     *
     * @param [type] $params
     * @return object
     */
    public function getList($params)
    {
        $model = new self;
        if (isset($params['name']) && $params['name']) {
            $model = $model->like('name', $params['name']);
        }
        $list = $model->with(['purchaser'])->order('create_time desc')->paginate($params);

        foreach ($list as &$item) {
            $item['id'] = $item['uuid'];
            $item['purchaser_id'] = $item['staff_uuid'];
            $item['shop_supplier_id'] = $item['purchaser']['company_uuid'];
            $item['purchaser']['shop_user_id'] = $item['purchaser']['uuid'];
        }
        
        return $list;
    }

    public function getSelectList()
    {
        $model = new self;
        $list = $model->with(['purchaser'])->withAttr('id', function ($value, $data) {
            return $data['uuid'];
        })->order('create_time desc')->select();
        return $list;
    }

    /**
     * 详情
     *
     * @param [type] $erp_supplier_id
     * @return self
     */
    public function detail($erp_supplier_id)
    {
        $model = new self;
        $info = $model->with(['purchaser'])->where('uuid', $erp_supplier_id)->find();
        return $info;
    }

    /**
     * 新增
     */
    public function add($data)
    {
        $model = new self;
        // 验证供应商名称唯一性
        $exist = $model->where('name', $data['name'])->find();
        if ($exist) {
            $this->error = '该供应商名称已存在';
            return false;
        }
        //
        $data['contact_name'] = $data['contact_person'] ?? '';
        $data['staff_uuid'] = $data['purchaser_id'] ?? 0;
        $model->save($data);
        return $model->uuid;
    }

    /**
     * 编辑
     */
    public function edit($data)
    {
        // 验证供应商名称唯一性
        $exist = (new self)->where('name', $data['name'])->where('uuid', '<>', $this->uuid)->find();
        if ($exist) {
            $this->error = '该供应商名称已存在';
            return false;
        }
        //
        $data['contact_name'] = $data['contact_person'] ?? '';
        $data['staff_uuid'] = $data['purchaser_id'] ?? 0;
        $this->save($data);
        return $this->uuid;
    }

    /**
     * 删除
     */
    public function del()
    {
        return $this->delete();
    }
}
