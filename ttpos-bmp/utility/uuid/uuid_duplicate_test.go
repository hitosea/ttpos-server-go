package uuid

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	goid "github.com/ace-zhaoy/go-id"
	"github.com/gogf/gf/v2/frame/g"
)

// TestIDDuplicateIssue 测试多实例在同一秒生成重复 ID 的问题
func TestIDDuplicateIssue(t *testing.T) {
	ctx := context.Background()

	// 模拟两个不同的实例：appType=1, instanceID=1 和 appType=2, instanceID=2
	var instance1, instance2 *goid.ID

	// 初始化实例 1: appType=1 (ERP), instanceID=1
	instance1 = goid.NewID()
	instance1.SetNode((1<<6)|1, 10) // nodeID = (1 << 6) | 1 = 65

	// 初始化实例 2: appType=2 (Message), instanceID=2
	instance2 = goid.NewID()
	instance2.SetNode((2<<6)|2, 10) // nodeID = (2 << 6) | 2 = 130

	g.Log().Infof(ctx, "[Test] Instance1 nodeID: %d", 65)
	g.Log().Infof(ctx, "[Test] Instance2 nodeID: %d", 130)

	// 同步启动，确保在同一秒内生成
	var wg sync.WaitGroup
	var ids1, ids2 []int64
	var mu sync.Mutex

	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			id := instance1.Generate()
			mu.Lock()
			ids1 = append(ids1, id)
			mu.Unlock()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			id := instance2.Generate()
			mu.Lock()
			ids2 = append(ids2, id)
			mu.Unlock()
		}
	}()

	wg.Wait()

	// 检查是否有重复
	idMap := make(map[int64]bool)
	for _, id := range ids1 {
		if idMap[id] {
			t.Errorf("[Instance1] 发现重复 ID: %d", id)
		}
		idMap[id] = true
	}

	for _, id := range ids2 {
		if idMap[id] {
			t.Errorf("[跨实例] 发现重复 ID: %d - 实例间冲突！", id)
		}
		idMap[id] = true
	}

	// 打印生成的 ID
	g.Log().Infof(ctx, "[Test] Instance1 IDs: %v", ids1)
	g.Log().Infof(ctx, "[Test] Instance2 IDs: %v", ids2)

	// 解析 ID 结构
	for i, id := range ids1 {
		timestamp, counter := goid.ResolveID(id, instance1)
		nodeID, _ := instance1.GetNode()
		g.Log().Infof(ctx, "[Instance1 ID %d] timestamp=%d, counter=%d, nodeID=%d", i, timestamp, counter, nodeID)
	}

	for i, id := range ids2 {
		timestamp, counter := goid.ResolveID(id, instance2)
		nodeID, _ := instance2.GetNode()
		g.Log().Infof(ctx, "[Instance2 ID %d] timestamp=%d, counter=%d, nodeID=%d", i, timestamp, counter, nodeID)
	}
}

// TestIDDetailedAnalysis 详细分析 ID 结构
func TestIDDetailedAnalysis(t *testing.T) {
	ctx := context.Background()

	// 创建两个实例
	instance1 := goid.NewID()
	instance1.SetNode((1<<6)|1, 10) // nodeID = 65

	instance2 := goid.NewID()
	instance2.SetNode((2<<6)|2, 10) // nodeID = 130

	g.Log().Infof(ctx, "[Detail] 实例1 nodeID=%d (0b%s)", 65, fmt.Sprintf("%010b", 65))
	g.Log().Infof(ctx, "[Detail] 实例2 nodeID=%d (0b%s)", 130, fmt.Sprintf("%010b", 130))

	// 等待到整秒（这是为了确保两个实例在同一秒内生成 ID，便于测试复现问题）
	nextSecond := time.Now().Add(time.Duration(1000000000 - time.Now().UnixNano()%1000000000))
	time.Sleep(time.Until(nextSecond))

	// 生成 ID
	id1 := instance1.Generate()
	id2 := instance2.Generate()

	g.Log().Infof(ctx, "[Detail] ID1 = %d (0b%064b)", id1, id1)
	g.Log().Infof(ctx, "[Detail] ID2 = %d (0b%064b)", id2, id2)

	// 解析 ID
	timestamp1, counter1 := goid.ResolveID(id1, instance1)
	timestamp2, counter2 := goid.ResolveID(id2, instance2)

	g.Log().Infof(ctx, "[Detail] ID1 解析: timestamp=%d, counter=%d", timestamp1, counter1)
	g.Log().Infof(ctx, "[Detail] ID2 解析: timestamp=%d, counter=%d", timestamp2, counter2)

	// 检查 nodeBits
	nodeBits1, _ := instance1.GetNode()
	nodeBits2, _ := instance2.GetNode()
	g.Log().Infof(ctx, "[Detail] nodeBits=%d", nodeBits1)

	// 计算 ID 组成
	cBits := 21 - nodeBits1
	g.Log().Infof(ctx, "[Detail] 计数器位数 cBits=%d, 掩码=0x%08x", cBits, (1<<cBits)-1)

	// 分析低位
	lowBits1 := id1 & ((1 << cBits) - 1)
	lowBits2 := id2 & ((1 << cBits) - 1)
	g.Log().Infof(ctx, "[Detail] ID1 低%d位 = %d (0b%0*b)", cBits, lowBits1, cBits, lowBits1)
	g.Log().Infof(ctx, "[Detail] ID2 低%d位 = %d (0b%0*b)", cBits, lowBits2, cBits, lowBits2)

	if lowBits1 == lowBits2 {
		t.Errorf("[Detail] 低%d位相同！可能发生冲突", cBits)
	}

	// 分析节点 ID 位置
	nodePos1 := (id1 >> cBits) & ((1 << nodeBits1) - 1)
	nodePos2 := (id2 >> cBits) & ((1 << nodeBits2) - 1)
	g.Log().Infof(ctx, "[Detail] ID1 节点ID位置 = %d", nodePos1)
	g.Log().Infof(ctx, "[Detail] ID2 节点ID位置 = %d", nodePos2)
}
