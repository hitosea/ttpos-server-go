<?php

namespace help;

use Exception;

// 设置起始时间戳为2024-01-01的毫秒时间戳
define('EPOCH', 1704067200000);
define('NUMWORKERBITS', 10);
define('NUMSEQUENCEBITS', 12);
define('MAXWORKERID', (-1 ^ (-1 << NUMWORKERBITS)));    // 集群ID + 机器ID， 10位，最多支持1024台机器
define('MAXSEQUENCE', (-1 ^ (-1 << NUMSEQUENCEBITS)));  // 序列，12位，每台机器每毫秒内最多产生4096个序列号

/**
 * Snowflake 生成唯一ID算法，固定返回16位整数
 */
class SnowflakeHelp
{
    private $_lastTimestamp = -1;
    private $_sequence = 0;
    private $_workerId = 1;
    private $_sequenceMask = MAXSEQUENCE;
    private $_workerIdShift = NUMSEQUENCEBITS;
    private $_timestampShift = NUMWORKERBITS + NUMSEQUENCEBITS;

    /**
     * 构造函数
     * @param int $workerId 工作机器ID
     * @throws Exception 如果工作机器ID超出范围则抛出异常
     */
    public function __construct($workerId)
    {
        if ($workerId < 0 || $workerId > MAXWORKERID) {
            throw new Exception("Worker ID不能小于0或大于" . MAXWORKERID);
        }
        $this->_workerId = $workerId;
    }

    /**
     * 生成下一个ID
     * @return string 16位数字ID
     * @throws Exception 如果时钟回拨则抛出异常
     */
    public function next()
    {
        $timestamp = $this->timestamp();
        
        // 检查时钟回拨
        if ($timestamp < $this->_lastTimestamp) {
            $clockDiff = $this->_lastTimestamp - $timestamp;
            // 如果时钟回拨时间小于5毫秒，等待
            if ($clockDiff <= 5) {
                usleep($clockDiff * 1000); // 休眠相应的毫秒数
                $timestamp = $this->timestamp(); // 重新获取时间戳
                // 如果仍然小于上次时间戳，抛出异常
                if ($timestamp < $this->_lastTimestamp) {
                    throw new Exception("时钟回拨，拒绝生成ID");
                }
            } else {
                // 时钟回拨超过5毫秒，可能是系统时间被调整，抛出异常
                throw new Exception("时钟回拨过大，拒绝生成ID");
            }
        }
        
        // 如果是同一毫秒
        if ($timestamp == $this->_lastTimestamp) {
            // 序列号递增
            $this->_sequence = ($this->_sequence + 1) & $this->_sequenceMask;
            // 序列号用尽，等待下一毫秒
            if ($this->_sequence == 0) {
                $timestamp = $this->waitNextMilli();
            }
        } else {
            // 不同毫秒，序列号重置
            $this->_sequence = 0;
        }
        
        // 记录最后时间戳
        $this->_lastTimestamp = $timestamp;
        
        // 组合ID
        try {
            $id = $this->generateId($timestamp);
            // 确保ID为16位
            $idStr = (string)$id;
            if (strlen($idStr) > 16) {
                $idStr = substr($idStr, 0, 16);
            } else if (strlen($idStr) < 16) {
                $idStr = str_pad($idStr, 16, '0', STR_PAD_RIGHT);
            }
            return $idStr;
        } catch (Exception $e) {
            // 如果ID生成失败，使用备用方案
            return $this->generateBackupId();
        }
    }

    /**
     * 生成ID
     * @param int $timestamp 时间戳
     * @return int 生成的ID
     */
    private function generateId($timestamp)
    {
        // 限制时间戳位数，避免溢出
        $timestamp = $timestamp & 0x1FFFFF; // 取低21位
        
        // 组合ID各部分
        $id = ($timestamp << $this->_timestampShift) | 
              ($this->_workerId << $this->_workerIdShift) | 
              $this->_sequence;
              
        return $id;
    }
    
    /**
     * 生成备用ID（当主要算法失败时）
     * @return string 16位数字ID
     */
    private function generateBackupId()
    {
        // 使用微秒时间戳 + 随机数
        $time = microtime(true) * 10000;
        $random = mt_rand(1000, 9999);
        $id = $time . $random;
        // 确保16位
        return substr(str_replace('.', '', $id), 0, 16);
    }

    /**
     * 等待下一毫秒
     * @return int 新的时间戳
     */
    private function waitNextMilli()
    {
        $timestamp = $this->timestamp();
        while ($timestamp <= $this->_lastTimestamp) {
            usleep(100); // 休眠0.1毫秒
            $timestamp = $this->timestamp();
        }
        return $timestamp;
    }

    /**
     * 获取当前时间戳（相对于起始时间的毫秒数）
     * @return int 时间戳
     */
    private function timestamp()
    {
        return $this->millitime() - EPOCH;
    }

    /**
     * 获取当前毫秒时间戳
     * @return int 毫秒时间戳
     */
    private function millitime()
    {
        $microtime = microtime();
        $comps = explode(' ', $microtime);
        // 确保毫秒部分正确
        return (int)$comps[1] * 1000 + (int)($comps[0] * 1000);
    }
}
