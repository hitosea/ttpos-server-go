<?php

namespace help;

use PhpAmqpLib\Message\AMQPMessage;
use PhpAmqpLib\Connection\AMQPStreamConnection;

class RabbitmqHelp
{
    /**
     * 开始队列
     * @param $name, $message
     * @return void
     */
    public static function push(string $name, array $message = [])
    {
        if (!$message) {
            return true;
        }
        // 
        try{
            $config = config('rabbitmq');
            // 获取连接
            $connection = new AMQPStreamConnection($config['host'], $config['port'], $config['user'], $config['password']);
            // 获取连接的通道
            $channel = $connection->channel();
            // 直接创建一个队列
            /**
             * 关于 queue_declare参数的说明
             * params  queue  队列的名称
             * params  passive 是否消极的声明队列，如果存在，就把队列的信息返回， 如果没有就抛出错误，（是的， 你没看错，这个参数很鸡肋，所以一般为 false）
             * params  durable 是否持久化，意思是说就算队列服务挂了， 也不会丢失队列
             * params  exclusive  是否排外，如果设置为true ,表示只有本次连接中的channel 可以访问，其它channel 是不可以访问的
             * params  auto_delete  设置是否自动删除。为true 则设置队列为自动删除。自动删除的前提是, 至少有一个消费者连接到这个队列，之后所有与这个队列连接的消费者都断开时，才会自动删除
             * params  nowait 相当于做一个异步版的声明， 如果设置成true, 就是说方法调用完就结束，也不用等待创建队列是否成功，一般也设为false
             */
            $channel->queue_declare($name,false,false,false,false,false);
            $msg = new AMQPMessage(json_encode($message), [
                "delivery_mode" => AMQPMessage::DELIVERY_MODE_PERSISTENT          //使消息持久化
            ]);
            $channel->basic_publish($msg, "", $name);  //简单模式下，routing_key 和 队列名称是一样的
            //生产者调用完成后要关闭资源
            $channel->close();
            $connection->close();
            // 
            return true;
        }catch(\Exception $e){
            throw new \Exception("队列连接失败: " . $e->getMessage());
        }
    }
}
