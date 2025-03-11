<?php 

namespace app\common\model\supplier;

use app\common\model\BaseModel;
use app\common\model\settings\Printer;
use think\model\concern\SoftDelete;

class PrintingItem extends BaseModel 
{
    use SoftDelete;

    protected $name = 'product_printer_item';
    protected $pk = 'id';
    protected $deleteTime = 'delete_time';
    protected $autoWriteTimestamp = true;
    protected $defaultSoftDelete = 0;

    /**
     * 关联打印机
     */
    public function printer()
    {
        return $this->belongsTo(Printer::class, 'printer_uuid', 'uuid');
    }
}