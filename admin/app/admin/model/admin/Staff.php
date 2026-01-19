<?php

namespace app\admin\model\admin;

use app\common\model\BaseModel;
use think\facade\Db;
use think\model\concern\SoftDelete;
use think\facade\Config;

/**
 * 统一账号模型（saas.ttpos_staff）
 */
class Staff extends BaseModel
{
    use SoftDelete;

    protected $name = 'staff';
    protected $table = 'ttpos_staff'; // 指定完整表名，不添加前缀
    protected $pk = 'uuid';
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    
    /**
     * 获取账号列表（包含关联门店和角色信息）
     */
    public function getList($param)
    {
        $keyword = $param['keyword'] ?? '';
        
        $list = self::field('uuid, email, phone, real_name, is_disable, last_company_uuid, create_time, update_time')
            ->when($keyword, function ($q) use ($keyword) {
                $q->where(function ($qq) use ($keyword) {
                    $qq->like('email', $keyword);
                    $qq->orLike('phone', $keyword);
                    $qq->orLike('uuid', $keyword);
                });
            })
            ->where('delete_time', 0)
            // ->where('uuid', 2791510642688000)
            ->order('create_time', 'desc')
            ->paginate($param);
        
        // 为每个账号添加关联门店和角色信息
        foreach ($list as &$item) {
            $item['company_list'] = $this->getCompanyListWithRoles($item['uuid']);
        }
        
        return $list;
    }
    
    /**
     * 获取账号关联的门店列表和角色信息（使用 PDO）
     */
    private function getCompanyListWithRoles($staffUuid)
    {
        // 获取 saas 数据库的 PDO 连接
        $saasPdo = $this->getSaasPdoConnection();
        if (!$saasPdo) {
            return [];
        }
        
        // 使用 PDO 从 saas 数据库查询账号关联的门店
        $sql = "SELECT company_uuid, is_super FROM ttpos_company_staff WHERE uuid = :uuid AND delete_time = 0";
        $stmt = $saasPdo->prepare($sql);
        $stmt->execute([':uuid' => $staffUuid]);
        $companyStaffList = $stmt->fetchAll(\PDO::FETCH_ASSOC);
        
        $companyList = [];
        
        foreach ($companyStaffList as $cs) {
            $companyUuid = $cs['company_uuid'];

            // 获取门店名称（从门店数据库的 ttpos_company 表）
            $companyName = $this->getCompanyName($companyUuid);
            
            // 获取该账号在该门店下的角色（从门店数据库的 ttpos_staff_role 表）
            $roles = $this->getStaffRoles($staffUuid, $companyUuid);

            if ($cs['is_super'] == 1) {
                $roles[] = [
                    'role_uuid' => 0,
                    'role_name' => '超管',
                ];
            }
            
            $companyList[] = [
                'company_uuid' => $companyUuid,
                'company_name' => $companyName,
                'roles' => $roles,
            ];
        }
        
        return $companyList;
    }
    
    /**
     * 获取门店名称（使用 PDO）
     */
    private function getCompanyName($companyUuid)
    {
        $pdo = $this->getShopPdoConnection($companyUuid);
        if (!$pdo) {
            return '';
        }
        
        // 使用 PDO 查询门店名称
        $sql = "SELECT name FROM ttpos_company WHERE uuid = :uuid AND delete_time = 0 LIMIT 1";
        $stmt = $pdo->prepare($sql);
        $stmt->execute([':uuid' => $companyUuid]);
        $company = $stmt->fetch(\PDO::FETCH_ASSOC);
        
        return $company ? $company['name'] : '';
    }
    
    /**
     * 获取员工在门店下的角色列表（使用 PDO）
     */
    private function getStaffRoles($staffUuid, $companyUuid)
    {
        $pdo = $this->getShopPdoConnection($companyUuid);
        if (!$pdo) {
            return [];
        }
        
        // 使用 PDO 查询员工角色关系表
        $sql = "SELECT r.uuid as role_uuid, r.name as role_name 
                FROM ttpos_staff_role sr
                LEFT JOIN ttpos_role r ON sr.role_uuid = r.uuid
                WHERE sr.staff_uuid = :staff_uuid 
                AND sr.delete_time = 0 
                AND r.delete_time = 0";
        $stmt = $pdo->prepare($sql);
        $stmt->execute([':staff_uuid' => $staffUuid]);
        $staffRoles = $stmt->fetchAll(\PDO::FETCH_ASSOC);
        
        return $staffRoles ?: [];
    }
    
    /**
     * 获取 saas 数据库的 PDO 连接
     */
    private function getSaasPdoConnection()
    {
        try {
            $config = Config::get('database.connections.mysql');
            
            // 构建 PDO DSN
            $dsn = sprintf(
                'mysql:host=%s;port=%s;dbname=%s;charset=%s',
                $config['hostname'] ?? '127.0.0.1',
                $config['hostport'] ?? 3306,
                $config['database'] ?? 'saas',
                $config['charset'] ?? 'utf8mb4'
            );
            
            // 创建 PDO 连接
            $pdo = new \PDO(
                $dsn,
                $config['username'] ?? 'root',
                $config['password'] ?? '',
                [
                    \PDO::ATTR_ERRMODE => \PDO::ERRMODE_EXCEPTION,
                    \PDO::ATTR_DEFAULT_FETCH_MODE => \PDO::FETCH_ASSOC,
                    \PDO::ATTR_EMULATE_PREPARES => false,
                ]
            );
            
            return $pdo;
        } catch (\Exception $e) {
            return null;
        }
    }
    
    /**
     * 获取门店数据库的 PDO 连接
     */
    private function getShopPdoConnection($companyUuid)
    {
        try {
            // 门店数据库连接格式：shop{company_uuid}
            $dbName = 'shop' . $companyUuid;
            
            // 获取数据库配置
            $config = Config::get('database.connections.mysql');
            
            // 构建 PDO DSN
            $dsn = sprintf(
                'mysql:host=%s;port=%s;dbname=%s;charset=%s',
                $config['hostname'] ?? '127.0.0.1',
                $config['hostport'] ?? 3306,
                $dbName,
                $config['charset'] ?? 'utf8mb4'
            );
            
            // 创建 PDO 连接
            $pdo = new \PDO(
                $dsn,
                $config['username'] ?? 'root',
                $config['password'] ?? '',
                [
                    \PDO::ATTR_ERRMODE => \PDO::ERRMODE_EXCEPTION,
                    \PDO::ATTR_DEFAULT_FETCH_MODE => \PDO::FETCH_ASSOC,
                    \PDO::ATTR_EMULATE_PREPARES => false,
                ]
            );
            
            return $pdo;
        } catch (\Exception $e) {
            return null;
        }
    }
    
    /**
     * 添加账号
     */
    public function add($data)
    {
        $this->startTrans();
        try {
            // 验证邮箱唯一性
            $exists = self::where('email', $data['email'])
                ->where('delete_time', 0)
                ->find();
            if ($exists) {
                $this->error = '该邮箱已在平台注册';
                $this->rollback();
                return false;
            }
            
            // 验证手机号唯一性（只有非空手机号才验证）
            if (!empty($data['phone'])) {
                $exists = self::where('phone', $data['phone'])
                    ->where('phone', '<>', '')
                    ->where('delete_time', 0)
                    ->find();
                if ($exists) {
                    $this->error = '该手机号已在平台注册';
                    $this->rollback();
                    return false;
                }
            }
            
            // 创建统一账号
            $uuid = createUuid();
            $currentTime = time();
            $res = self::create([
                'uuid' => $uuid,
                'email' => trim($data['email']),
                'phone' => trim($data['phone'] ?? ''),
                'real_name' => trim($data['real_name'] ?? ''),
                'password' => hash_password_bcrypt($data['password']),
                'password_change_count' => 0,
                'password_change_time' => 0,
                'is_disable' => 0,
                'last_company_uuid' => $data['company_uuid'] ?? 0,
                'create_time' => $currentTime,
                'update_time' => $currentTime,
            ]);
            
            // 绑定门店（如果提供了 company_uuid）
            if (!empty($data['company_uuid'])) {
                Db::connect('saas')->table('ttpos_company_staff')->insert([
                    'uuid' => $uuid,
                    'company_uuid' => $data['company_uuid'],
                    'username' => trim($data['email']),
                    'phone' => trim($data['phone'] ?? ''),
                    'is_super' => 0,
                    'create_time' => $currentTime,
                    'update_time' => $currentTime,
                ]);
                
                // 如果提供了角色，设置角色
                if (!empty($data['role_uuids']) && is_array($data['role_uuids'])) {
                    $this->updateStaffRoles($uuid, $data['company_uuid'], $data['role_uuids']);
                }
            }
            
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }
    
    /**
     * 编辑账号（支持设置关联门店和角色）
     */
    public function edit($data)
    {
        $this->startTrans();
        try {
            $where = ['uuid' => $data['uuid']];
            
            // 验证邮箱唯一性（排除自己）
            $exists = self::where('email', $data['email'])
                ->where('uuid', '<>', $data['uuid'])
                ->where('delete_time', 0)
                ->find();
            if ($exists) {
                $this->error = '该邮箱已在平台注册';
                $this->rollback();
                return false;
            }
            
            // 验证手机号唯一性（只有非空手机号才验证，排除自己）
            if (!empty($data['phone'])) {
                $exists = self::where('phone', $data['phone'])
                    ->where('phone', '<>', '')
                    ->where('uuid', '<>', $data['uuid'])
                    ->where('delete_time', 0)
                    ->find();
                if ($exists) {
                    $this->error = '该手机号已在平台注册';
                    $this->rollback();
                    return false;
                }
            }
            
            $arr = [
                'email' => trim($data['email']),
                'phone' => trim($data['phone'] ?? ''),
                'real_name' => trim($data['real_name'] ?? ''),
                'update_time' => time(),
            ];
            
            // 如果提供了密码，使用 bcrypt 加密
            if (!empty($data['password'])) {
                $arr['password'] = hash_password_bcrypt($data['password']);
                $arr['password_change_count'] = Db::raw('password_change_count + 1');
                $arr['password_change_time'] = time();
            }
            
            self::update($arr, $where);
            
            // 如果提供了 company_list，更新门店关联和角色设置
            if (!empty($data['company_list']) && is_array($data['company_list'])) {
                $this->updateCompanyList($data['uuid'], $data['company_list'], trim($data['email']), trim($data['phone'] ?? ''));
            }
            
            $this->commit();
            return true;
        } catch (\Exception $e) {
            $this->error = $e->getMessage();
            $this->rollback();
            return false;
        }
    }
    
    /**
     * 更新账号关联的门店列表和角色设置
     */
    private function updateCompanyList($staffUuid, $companyList, $email, $phone)
    {
        $saasDb = Db::connect('saas');
        $currentTime = time();
        
        // 获取当前关联的门店列表
        $currentCompanyList = $saasDb->table('ttpos_company_staff')
            ->where('uuid', $staffUuid)
            ->where('delete_time', 0)
            ->column('company_uuid');
        
        $newCompanyUuids = array_column($companyList, 'company_uuid');
        
        // 删除不在新列表中的门店关联
        $toDelete = array_diff($currentCompanyList, $newCompanyUuids);
        if (!empty($toDelete)) {
            $saasDb->table('ttpos_company_staff')
                ->where('uuid', $staffUuid)
                ->where('company_uuid', 'in', $toDelete)
                ->update(['delete_time' => $currentTime, 'update_time' => $currentTime]);
        }
        
        // 添加或更新门店关联和角色设置
        foreach ($companyList as $company) {
            $companyUuid = $company['company_uuid'];
            $roleUuids = $company['role_uuids'] ?? [];
            
            // 从商家数据库获取员工的 is_disable 状态
            $shopPdo = $this->getShopPdoConnection($companyUuid);
            $isDisable = 0;
            if ($shopPdo) {
                try {
                    $stmt = $shopPdo->prepare("SELECT is_disable FROM ttpos_staff WHERE uuid = ? AND delete_time = 0 LIMIT 1");
                    $stmt->execute([$staffUuid]);
                    $result = $stmt->fetch(\PDO::FETCH_ASSOC);
                    if ($result) {
                        $isDisable = (int)$result['is_disable'];
                    }
                } catch (\Exception $e) {
                    // 如果查询失败，使用默认值 0
                }
            }
            
            // 检查门店关联是否存在
            $exists = $saasDb->table('ttpos_company_staff')
                ->where('uuid', $staffUuid)
                ->where('company_uuid', $companyUuid)
                ->where('delete_time', 0)
                ->find();
            
            if ($exists) {
                // 更新现有关联
                $saasDb->table('ttpos_company_staff')
                    ->where('uuid', $staffUuid)
                    ->where('company_uuid', $companyUuid)
                    ->update([
                        'username' => $email,
                        'phone' => $phone,
                        'is_disable' => $isDisable,
                        'update_time' => $currentTime,
                    ]);
            } else {
                // 创建新关联
                $saasDb->table('ttpos_company_staff')
                    ->insert([
                        'uuid' => $staffUuid,
                        'company_uuid' => $companyUuid,
                        'username' => $email,
                        'phone' => $phone,
                        'is_super' => 0,
                        'is_disable' => $isDisable,
                        'create_time' => $currentTime,
                        'update_time' => $currentTime,
                    ]);
            }
            
            // 更新门店数据库中的角色设置
            $this->updateStaffRoles($staffUuid, $companyUuid, $roleUuids);
        }
    }
    
    /**
     * 更新员工在门店下的角色设置（使用 PDO）
     */
    private function updateStaffRoles($staffUuid, $companyUuid, $roleUuids)
    {
        // 获取门店数据库的 PDO 连接
        $pdo = $this->getShopPdoConnection($companyUuid);
        if (!$pdo) {
            return;
        }
        
        $currentTime = time();
        
        try {
            // 开始事务
            $pdo->beginTransaction();
            
            // 删除旧的角色关联（软删除）
            $sql = "UPDATE ttpos_staff_role SET delete_time = :delete_time, update_time = :update_time WHERE staff_uuid = :staff_uuid AND delete_time = 0";
            $stmt = $pdo->prepare($sql);
            $stmt->execute([
                ':delete_time' => $currentTime,
                ':update_time' => $currentTime,
                ':staff_uuid' => $staffUuid,
            ]);
            
            // 添加新的角色关联
            if (!empty($roleUuids)) {
                $sql = "INSERT INTO ttpos_staff_role (uuid, staff_uuid, role_uuid, create_time, update_time) VALUES (:uuid, :staff_uuid, :role_uuid, :create_time, :update_time)";
                $stmt = $pdo->prepare($sql);
                
                foreach ($roleUuids as $roleUuid) {
                    $stmt->execute([
                        ':uuid' => createUuid(),
                        ':staff_uuid' => $staffUuid,
                        ':role_uuid' => $roleUuid,
                        ':create_time' => $currentTime,
                        ':update_time' => $currentTime,
                    ]);
                }
            }
            
            // 提交事务
            $pdo->commit();
        } catch (\Exception $e) {
            // 回滚事务
            if ($pdo->inTransaction()) {
                $pdo->rollBack();
            }
            throw $e;
        }
    }
    
    /**
     * 更新启用禁用状态
     */
    public function updateStatus()
    {
        $this->is_disable = $this->is_disable == 0 ? 1 : 0;
        $this->update_time = time();
        return $this->save();
    }
    
    /**
     * 获取账号详情
     */
    public static function detail($uuid)
    {
        $model = new static();
        return $model::where('uuid', $uuid)
            ->where('delete_time', 0)
            ->find();
    }
}
