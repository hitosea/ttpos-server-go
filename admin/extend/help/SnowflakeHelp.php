<?php

namespace help;

use Exception;

// 设置起始时间戳为2024-01-01的毫秒时间戳
define('EPOCH', 1704067200000);
define('NUMWORKERBITS', 8);      // 减少工作机器位数
define('NUMSEQUENCEBITS', 12);
define('MAXWORKERID', (-1 ^ (-1 << NUMWORKERBITS)));    // 最多支持256台机器
define('MAXSEQUENCE', (-1 ^ (-1 << NUMSEQUENCEBITS)));  // 每毫秒最多4096个序列号

/**
 * Snowflake 生成唯一ID算法，固定返回18位整数
 */
class SnowflakeHelp
{
    private $_lastTimestamp;
    private $_sequence = 0;
    private $_workerId = 1;

    public function __construct($workerId)
    {
        if (($workerId < 0) || ($workerId > MAXWORKERID)) {
            throw new \Exception("Worker ID must be between 0 and " . MAXWORKERID);
        }
        $this->_workerId = $workerId;
    }

    /**
     * 生成下一个ID
     * @return int
     * @throws \Exception
     */
    public function next()
    {
        $ts = $this->timestamp();

        if ($ts < 0) {
            throw new \Exception("Clock moved backwards!");
        }

        if ($ts == $this->_lastTimestamp) {
            $this->_sequence = ($this->_sequence + 1) & MAXSEQUENCE;
            if ($this->_sequence == 0) {
                $ts = $this->waitNextMilli($this->_lastTimestamp);
            }
        } else {
            $this->_sequence = 0;
        }

        if ($ts < $this->_lastTimestamp) {
            throw new \Exception("Clock moved backwards!");
        }

        $this->_lastTimestamp = $ts;

        // 调整位运算
        $id = (($ts & 0x7FFFFFF) << 20) | ($this->_workerId << 12) | $this->_sequence;

        // 确保ID为18位且以55开头
        $id = intval('55' . substr(strval($id), 0, 16));

        return $id;
    }

    /**
     * 等待下一个毫秒
     */
    private function waitNextMilli($lastTimestamp)
    {
        $ts = $this->timestamp();
        while ($ts <= $lastTimestamp) {
            usleep(1);
            $ts = $this->timestamp();
        }
        return $ts;
    }

    /**
     * 获取当前时间戳（毫秒）
     */
    private function timestamp()
    {
        return $this->millitime() - EPOCH;
    }

    /**
     * 获取毫秒级时间戳
     */
    private function millitime()
    {
        $microtime = explode(' ', microtime());
        return intval(sprintf('%d%03d', $microtime[1], $microtime[0] * 1000));
    }
}
