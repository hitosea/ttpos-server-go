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
    private $_lastTimestamp;
    private $_sequence = 0;
    private $_workerId = 1;


    public function __construct($workerId)
    {
        if (($workerId < 0) || ($workerId > MAXWORKERID)) {
            return null;
        }
        $this->_workerId = $workerId;
    }

    public function next()
    {
        $ts = $this->timestamp();
        if ($ts == $this->_lastTimestamp) {
            $this->_sequence = ($this->_sequence + 1) & MAXSEQUENCE;
            if ($this->_sequence == 0) {
                $ts = $this->waitNextMilli($ts);
            }
        } else {
            $this->_sequence = 0;
        }

        if ($ts < $this->_lastTimestamp) {
            return 0;
        }

        $this->_lastTimestamp = $ts;

        // 生成ID并确保16位
        $id = $this->pack();
        return str_pad($id, 16, '0', STR_PAD_RIGHT);
    }

    private function pack()
    {
        // 调整移位以确保结果在16位范围内
        $timestamp = $this->_lastTimestamp & 0x1FFFFF; // 取低21位时间戳
        return ($timestamp << (NUMWORKERBITS + NUMSEQUENCEBITS))
             | ($this->_workerId << NUMSEQUENCEBITS)
             | $this->_sequence;
    }

    private function waitNextMilli($ts)
    {
        if ($ts = $this->_lastTimestamp) {
            sleep(0.1);
            $ts = $this->timestamp();
        }

        return $ts;
    }

    private function timestamp()
    {
        return $this->millitime() - EPOCH;
    }

    private function millitime()
    {
        $microtime = microtime();
        $comps = explode(' ', $microtime);
        return sprintf('%d%03d', $comps[1], $comps[0] * 1000);
    }
}
