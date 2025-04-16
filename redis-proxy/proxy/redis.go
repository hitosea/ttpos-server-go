package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"redis-proxy/cache"
	"redis-proxy/config"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Server struct {
	listener      net.Listener
	client        *redis.Client
	clusterClient *redis.ClusterClient
}

func NewServer(listenAddr string) (*Server, error) {
	// 创建Redis集群客户端
	var client *redis.Client
	var clusterClient *redis.ClusterClient
	hostList := strings.Split(config.Redis.Host, ",")
	if len(hostList) > 1 {
		clusterClient = cache.Global.GetClusterClient()
	} else {
		client = cache.Global.GetClient()
	}
	if client == nil && clusterClient == nil {
		return nil, fmt.Errorf("无法获取Redis客户端")
	}

	// 测试集群连接
	ctx := context.Background()
	if clusterClient != nil {
		if err := clusterClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("无法连接到Redis集群: %v", err)
		}
	} else {
		if err := client.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("无法连接到Redis: %v", err)
		}
	}

	// 创建TCP监听器
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("无法监听端口 %s: %v", listenAddr, err)
	}

	return &Server{
		listener:      listener,
		client:        client,
		clusterClient: clusterClient,
	}, nil
}

func (s *Server) Start() error {
	if s == nil || s.listener == nil {
		return fmt.Errorf("服务器未正确初始化")
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("接受连接错误: %v\n", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// 读取RESP协议命令
		command, err := readCommand(reader)
		if err != nil {
			return
		}

		// 处理命令
		response, err := s.handleCommand(command)
		if err != nil {
			fmt.Printf("处理命令错误: %v\n", err)
			response = formatError(err)
		}

		// 发送响应
		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}

func (s *Server) handleCommand(cmd []string) ([]byte, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("ERR 未指定命令")
	}

	ctx := context.Background()
	var result interface{}
	var err error

	// 处理特殊错误的函数
	handleRedisError := func(err error) ([]byte, error) {
		if err == redis.Nil {
			// 当键不存在时，返回 nil
			return formatResponse(nil), nil
		}
		return nil, err
	}

	// 统一使用大写命令进行处理
	cmdName := strings.ToUpper(cmd[0])
	switch cmdName {
	case "PING":
		return []byte("+PONG\r\n"), nil

	case "AUTH":
		// 总是返回 OK，因为我们不需要实际的认证
		return []byte("+OK\r\n"), nil

	case "QUIT":
		return []byte("+OK\r\n"), nil

	case "GET":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.Get(ctx, cmd[1]).Result()
		} else {
			result, err = s.clusterClient.Get(ctx, cmd[1]).Result()
		}

	case "SET":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.Set(ctx, cmd[1], cmd[2], 0).Result()
		} else {
			result, err = s.clusterClient.Set(ctx, cmd[1], cmd[2], 0).Result()
		}

	case "SETEX":
		if len(cmd) < 4 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		// 解析过期时间（秒）
		seconds, parseErr := strconv.ParseInt(cmd[2], 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("ERR value is not an integer or out of range")
		}
		if s.client != nil {
			result, err = s.client.SetEx(ctx, cmd[1], cmd[3], time.Duration(seconds)*time.Second).Result()
		} else {
			result, err = s.clusterClient.SetEx(ctx, cmd[1], cmd[3], time.Duration(seconds)*time.Second).Result()
		}

	case "DEL":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.Del(ctx, cmd[1:len(cmd)]...).Result()
		} else {
			result, err = s.clusterClient.Del(ctx, cmd[1:len(cmd)]...).Result()
		}

	case "EXISTS":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.Exists(ctx, cmd[1:len(cmd)]...).Result()
		} else {
			result, err = s.clusterClient.Exists(ctx, cmd[1:len(cmd)]...).Result()
		}

	case "EXPIRE":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		seconds, parseErr := strconv.ParseInt(cmd[2], 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("ERR value is not an integer or out of range")
		}
		if s.client != nil {
			result, err = s.client.Expire(ctx, cmd[1], time.Duration(seconds)*time.Second).Result()
		} else {
			result, err = s.clusterClient.Expire(ctx, cmd[1], time.Duration(seconds)*time.Second).Result()
		}

	case "INCR":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.Incr(ctx, cmd[1]).Result()
		} else {
			result, err = s.clusterClient.Incr(ctx, cmd[1]).Result()
		}

	case "DECR":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.Decr(ctx, cmd[1]).Result()
		} else {
			result, err = s.clusterClient.Decr(ctx, cmd[1]).Result()
		}

	case "HGET":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.HGet(ctx, cmd[1], cmd[2]).Result()
		} else {
			result, err = s.clusterClient.HGet(ctx, cmd[1], cmd[2]).Result()
		}

	case "HSET":
		if len(cmd) < 4 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.HSet(ctx, cmd[1], cmd[2], cmd[3]).Result()
		} else {
			result, err = s.clusterClient.HSet(ctx, cmd[1], cmd[2], cmd[3]).Result()
		}

	case "HDEL":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.HDel(ctx, cmd[1], cmd[2:len(cmd)]...).Result()
		} else {
			result, err = s.clusterClient.HDel(ctx, cmd[1], cmd[2:len(cmd)]...).Result()
		}

	case "LPUSH":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		// 将[]string转换为[]interface{}
		members := make([]interface{}, len(cmd)-2)
		for i, v := range cmd[2:] {
			members[i] = v
		}
		if s.client != nil {
			result, err = s.client.LPush(ctx, cmd[1], members...).Result()
		} else {
			result, err = s.clusterClient.LPush(ctx, cmd[1], members...).Result()
		}

	case "RPUSH":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		// 将[]string转换为[]interface{}
		members := make([]interface{}, len(cmd)-2)
		for i, v := range cmd[2:] {
			members[i] = v
		}
		if s.client != nil {
			result, err = s.client.RPush(ctx, cmd[1], members...).Result()
		} else {
			result, err = s.clusterClient.RPush(ctx, cmd[1], members...).Result()
		}

	case "LPOP":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.LPop(ctx, cmd[1]).Result()
		} else {
			result, err = s.clusterClient.LPop(ctx, cmd[1]).Result()
		}

	case "RPOP":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.RPop(ctx, cmd[1]).Result()
		} else {
			result, err = s.clusterClient.RPop(ctx, cmd[1]).Result()
		}

	case "SADD":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		// 将[]string转换为[]interface{}
		members := make([]interface{}, len(cmd)-2)
		for i, v := range cmd[2:] {
			members[i] = v
		}
		if s.client != nil {
			result, err = s.client.SAdd(ctx, cmd[1], members...).Result()
		} else {
			result, err = s.clusterClient.SAdd(ctx, cmd[1], members...).Result()
		}

	case "SREM":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		// 将[]string转换为[]interface{}
		members := make([]interface{}, len(cmd)-2)
		for i, v := range cmd[2:] {
			members[i] = v
		}
		if s.client != nil {
			result, err = s.client.SRem(ctx, cmd[1], members...).Result()
		} else {
			result, err = s.clusterClient.SRem(ctx, cmd[1], members...).Result()
		}

	case "SMEMBERS":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		if s.client != nil {
			result, err = s.client.SMembers(ctx, cmd[1]).Result()
		} else {
			result, err = s.clusterClient.SMembers(ctx, cmd[1]).Result()
		}

	case "ZADD":
		if len(cmd) < 4 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		score, parseErr := strconv.ParseFloat(cmd[2], 64)
		if parseErr != nil {
			return nil, fmt.Errorf("ERR value is not a valid float")
		}
		if s.client != nil {
			result, err = s.client.ZAdd(ctx, cmd[1], redis.Z{Score: score, Member: cmd[3]}).Result()
		} else {
			result, err = s.clusterClient.ZAdd(ctx, cmd[1], redis.Z{Score: score, Member: cmd[3]}).Result()
		}

	case "ZREM", "zrem":
		if len(cmd) < 3 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		// 将[]string转换为[]interface{}
		members := make([]interface{}, len(cmd)-2)
		for i, v := range cmd[2:] {
			members[i] = v
		}
		if s.client != nil {
			result, err = s.client.ZRem(ctx, cmd[1], members...).Result()
		} else {
			result, err = s.clusterClient.ZRem(ctx, cmd[1], members...).Result()
		}

	case "ZRANGE":
		if len(cmd) < 4 {
			return nil, fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd[0])
		}
		start, parseErr := strconv.ParseInt(cmd[2], 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("ERR value is not an integer or out of range")
		}
		stop, parseErr := strconv.ParseInt(cmd[3], 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("ERR value is not an integer or out of range")
		}
		if s.client != nil {
			result, err = s.client.ZRange(ctx, cmd[1], start, stop).Result()
		} else {
			result, err = s.clusterClient.ZRange(ctx, cmd[1], start, stop).Result()
		}

	case "CLUSTER":
		if s.client != nil {
			result, err = s.client.ClusterInfo(ctx).Result()
		} else {
			result, err = s.clusterClient.ClusterInfo(ctx).Result()
		}

	case "INFO", "info":
		info := "# Server\r\n" +
			"redis_version:6.0.0\r\n" +
			"redis_mode:standalone\r\n" +
			"tcp_port:6739\r\n" +
			"# Clients\r\n" +
			"connected_clients:1\r\n" +
			"# Memory\r\n" +
			"used_memory_human:1.00M\r\n"
		result = info

	case "CONFIG", "config":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for 'config' command")
		}
		switch strings.ToLower(cmd[1]) {
		case "get":
			if len(cmd) < 3 {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'config get' command")
			}
			// 返回一个空的配置列表
			result = []interface{}{}
		case "set":
			if len(cmd) < 4 {
				return nil, fmt.Errorf("ERR wrong number of arguments for 'config set' command")
			}
			// 假装设置成功
			result = "OK"
		default:
			return nil, fmt.Errorf("ERR unknown subcommand or wrong number of arguments for '%s'. Try CONFIG HELP", cmd[1])
		}

	case "DBSIZE":
		if s.client != nil {
			result, err = s.client.DBSize(ctx).Result()
		} else {
			result, err = s.clusterClient.DBSize(ctx).Result()
		}

	case "SCAN":
		var cursor uint64
		var scanErr error
		if len(cmd) > 1 {
			cursor, scanErr = strconv.ParseUint(cmd[1], 10, 64)
			if scanErr != nil {
				return nil, fmt.Errorf("ERR value is not an integer or out of range")
			}
		}

		var match string
		var count int64 = 10
		for i := 2; i < len(cmd); i += 2 {
			if i+1 >= len(cmd) {
				break
			}
			switch strings.ToLower(cmd[i]) {
			case "match":
				match = cmd[i+1]
			case "count":
				if c, err := strconv.ParseInt(cmd[i+1], 10, 64); err == nil {
					count = c
				}
			}
		}

		var keys []string
		var nextCursor uint64
		if s.client != nil {
			keys, nextCursor, err = s.client.Scan(ctx, cursor, match, count).Result()
			if err != nil {
				return nil, err
			}
		} else {
			keys, nextCursor, err = s.clusterClient.Scan(ctx, cursor, match, count).Result()
			if err != nil {
				return nil, err
			}
		}
		result = []interface{}{nextCursor, keys}

	case "MODULE":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for 'module' command")
		}
		switch strings.ToLower(cmd[1]) {
		case "list":
			// 返回空模块列表
			result = []interface{}{}
		case "load":
			result = "OK"
		default:
			return nil, fmt.Errorf("ERR unknown subcommand or wrong number of arguments for 'module' command")
		}

	case "SELECT":
		if len(cmd) < 2 {
			return nil, fmt.Errorf("ERR wrong number of arguments for 'select' command")
		}
		// 检查数据库索引是否为数字
		db, err := strconv.Atoi(cmd[1])
		if err != nil {
			return nil, fmt.Errorf("ERR invalid DB index")
		}
		if db < 0 {
			return nil, fmt.Errorf("ERR DB index is out of range")
		}
		// 切换数据库
		if s.client != nil {
			err = s.client.Do(ctx, "SELECT", db).Err()
		} else {
			err = s.clusterClient.Do(ctx, "SELECT", db).Err()
		}
		if err != nil {
			return nil, err
		}
		result = "OK"

	default:
		return nil, fmt.Errorf("ERR unknown command '%s'", cmd[0])
	}

	if err != nil {
		return handleRedisError(err)
	}

	return formatResponse(result), nil
}

func readCommand(reader *bufio.Reader) ([]string, error) {
	// 读取第一个字节
	firstByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch firstByte {
	case '*':
		// 读取数组长度
		length, err := readInteger(reader)
		if err != nil {
			return nil, err
		}

		// 读取数组元素
		command := make([]string, length)
		for i := 0; i < length; i++ {
			// 读取 $
			if b, err := reader.ReadByte(); err != nil || b != '$' {
				return nil, fmt.Errorf("protocol error: expected '$', got '%c'", b)
			}

			// 读取字符串长度
			strLen, err := readInteger(reader)
			if err != nil {
				return nil, err
			}

			// 读取字符串内容
			str := make([]byte, strLen)

			// 使用ReadFull确保完整读取所有字节
			totalRead := 0
			for totalRead < strLen {
				n, err := reader.Read(str[totalRead:])
				if err != nil {
					return nil, err
				}
				totalRead += n
				if totalRead >= strLen {
					break
				}
			}

			// 读取 \r\n
			if _, err := reader.ReadBytes('\n'); err != nil {
				return nil, err
			}

			command[i] = string(str)
		}
		return command, nil
	default:
		return nil, fmt.Errorf("protocol error: expected '*', got '%c'", firstByte)
	}
}

func readInteger(reader *bufio.Reader) (int, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimRight(line, "\r\n"))
}

func formatResponse(result interface{}) []byte {
	switch v := result.(type) {
	case nil:
		return []byte("$-1\r\n")
	case string:
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case error:
		return []byte(fmt.Sprintf("-ERR %s\r\n", v.Error()))
	case bool:
		// 布尔值转换为整数
		if v {
			return []byte(":1\r\n")
		}
		return []byte(":0\r\n")
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return []byte(fmt.Sprintf(":%v\r\n", v))
	case float32, float64:
		return []byte(fmt.Sprintf(":%v\r\n", v))
	case []string:
		// 处理字符串数组
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("*%d\r\n", len(v)))
		for _, item := range v {
			buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(item), item))
		}
		return buffer.Bytes()
	case []interface{}:
		// 处理接口数组
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("*%d\r\n", len(v)))
		for _, item := range v {
			itemBytes := formatResponse(item)
			buffer.Write(itemBytes)
		}
		return buffer.Bytes()
	default:
		// 默认情况，尝试转换为字符串
		return []byte(fmt.Sprintf("+%v\r\n", v))
	}
}

func formatError(err error) []byte {
	return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
}
