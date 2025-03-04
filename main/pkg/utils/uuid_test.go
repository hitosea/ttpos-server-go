package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	InitSonyFlakeId()
	id, err := GetID18()
	if err != nil {
		log.Fatalf("GetID Failed Err: %#v\n", err)
	}
	fmt.Println("sonyFlake 生成 Uuid: ", id)
}

func TestGetUUID16(t *testing.T) {
	idMap := sync.Map{}
	wg := sync.WaitGroup{}
	var index uint64 = 0
	for i := 0; i < 1048575; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000000; j++ {
				id, err := GetID()
				if err != nil {
					log.Printf("生成ID失败: %v", err)
					continue
				}
				if _, exists := idMap.Load(id); exists {
					t.Errorf("重复的ID: %d", id)
				} else {
					idMap.Store(id, true)
				}
				index++
				fmt.Println(" 当前索引: ", index, "uuid16 生成 Uuid: ", id)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestGetUUID16For(t *testing.T) {
	idMap := sync.Map{}
	var index uint64 = 0
	for j := 0; j < 10000000000; j++ {
		id, err := GetID()
		if err != nil {
			log.Printf("生成ID失败: %v", err)
			continue
		}
		if _, exists := idMap.Load(id); exists {
			t.Errorf("重复的ID: %d", id)
		} else {
			idMap.Store(id, true)
		}
		index++
		fmt.Println(" 当前索引: ", index, "uuid16 生成 Uuid: ", id)
	}

}

func TestGetUUIDFile(t *testing.T) {
	gen()
}

func TestGetUUIDFile2(t *testing.T) {
	gen()
}
func gen() {
	idMap := sync.Map{}
	wg := sync.WaitGroup{}
	var index uint64 = 0
	file, err := os.Create("uuid16_ids.txt")
	if err != nil {
		log.Fatalf("无法创建文件: %v", err)
	}
	defer file.Close()
	for i := 0; i < 1048575; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000000; j++ {
				id, err := GetID()
				if err != nil {
					log.Printf("生成ID失败: %v", err)
					continue
				}
				if _, exists := idMap.Load(id); exists {
					log.Printf("重复的ID: %d", id)
				} else {
					idMap.Store(id, true)
					// 保存到文件中
					if _, err := file.WriteString(fmt.Sprintf("%d\n", id)); err != nil {
						log.Printf("写入文件失败: %v", err)
					}
				}
				index++
				fmt.Println(" 当前索引: ", index, "uuid16 生成 Uuid: ", id)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
func TestGetUUIDFile3(t *testing.T) {
	idMap := sync.Map{}
	file, err := os.Open("uuid16_ids.txt")
	if err != nil {
		log.Fatalf("无法打开文件: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := scanner.Text()
		if _, exists := idMap.Load(id); exists {
			log.Printf("重复的ID: %s", id)
		} else {
			idMap.Store(id, true)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("扫描文件时出错: %v", err)
	}
}
