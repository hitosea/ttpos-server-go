<?php
namespace db;
use help\StringHelp;
use think\db\Query as thinkQuery;

class Query extends  thinkQuery
{
	//克隆当前实例
    public function copy() {
    	return clone $this;
    }

    //克隆当前实例
    public function clone() {
    	return clone $this;
    }
    
    //分页-添加默认值
    public function pages($page=1,$limit=20) {
    	return $this->page( input('pageIndex', $page) ,  input('pageSize',$limit) );
    }
    
    /**
     * 根据查询条件获取当前模型的 分页列表
     */
    public function getPageList($page=1,$limit=20,$order=["id" =>"desc"]) {
        $page = request()->param("pageIndex",$page);
        $limit = request()->param("pageSize",$limit);
        $list = $this->order($order)->paginate(['page' => $page,'list_rows'=> $limit]);
        return $list;
    }

    /**
     * 自动赋值
     */
    public function fill(array $params) {
        $that =  $this;
        $fields = array_keys($that->getFields());
        foreach($fields as $field){
            foreach($params as $key => $param){
                if($field == $key){
                    $that->$field = $param;
                }
            }
        }
        return $that;
    }

    /**
     * 转成数组
     */
    public function toArray() {
        $that =  $this->clone();
        $fields = array_keys($that->getFields());
        $params = get_object_vars($that);
        $array = [];
        foreach($fields as $field){
            foreach($params as $key => $param){
                if($field == $key){
                    $array[$key] = $param;
                }
            }
        }
        return $array;
    }
    
    // 模糊查询
    public function like($column,$value) {
    	$value = StringHelp::escapeLikeStr($value);
    	return clone $this->where($column, 'like', '%' . trim($value) . '%');
    }

    // 模糊查询
    public function orLike($column,$value) {
        $value = StringHelp::escapeLikeStr($value);
    	return clone $this->whereOr($column, 'like', '%' . trim($value) . '%');
    }

    // 模糊查询
    public function notLike($column,$value) {
        $value = StringHelp::escapeLikeStr($value);
    	return clone $this->where($column, 'not like', '%' . trim($value) . '%');
    }

    // JSON_EXTRACT 模糊查询
    public function jsonLike($column, $value, $key = "") {
        if (!$key) {
            $lang = checkDetect();
            $languages = getSettingLanguages();
            $langKey = 'en';
            foreach ($languages as $language) {
                $name = $language['name'] ?? '';
                if ($name == $lang) {
                    $langKey = $language['key'];
                }
            }
            $key = $langKey;
        }
    	$value = StringHelp::escapeLikeStr($value);
    	$value = trim($value);
        $value = str_replace('\"', '\\\\\\\\"', $value);
        return clone $this->where("JSON_EXTRACT($column, '$.\"$key\"') LIKE '%$value%'");
    }

    // JSON_EXTRACT 等于查询
    public function jsonEqual($column, $value, $key = "") {
        if (!$key) {
            $lang = checkDetect();
            $languages = getSettingLanguages();
            $langKey = 'en';
            foreach ($languages as $language) {
                $name = $language['name'] ?? '';
                if ($name == $lang) {
                    $langKey = $language['key'];
                }
            }
            $key = $langKey;
        }
    	$value = trim($value);
        return clone $this->where("JSON_UNQUOTE(JSON_EXTRACT($column, '$.\"$key\"')) = '$value'");
    }
    
}
