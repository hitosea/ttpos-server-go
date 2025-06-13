package google_bucket

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	testBucketName      = "ttpos_print_test"
	testCredentialsFile = "/Users/weifashi/Desktop/ProjectFile/TTPOS/ttpos-server-go/credentials.json"
)

// # 谷歌云证书文件名
// GOOGLE_APPLICATION_CREDENTIALS_FILE_NAME=credentials.json
// # 谷歌云bucket - 安装包
// GOOGLE_APPLICATION_BUCKET_NAME=dc_apk
// # 谷歌云bucket - 安装包 - 环境
// GOOGLE_APPLICATION_BUCKET_ENV=Test
// # 谷歌云- 上传图片的bucket目录名称 （diacan-test）
// GOOGLE_APPLICATION_UPLOADS_BUCKET_NAME=diacan-test
// # 谷歌云 - 上传图片的目录名称 （TTPOS-Test）
// GOOGLE_APPLICATION_UPLOADS_CATALOGUE_NAME=TTPOS-Test

// 测试初始化函数
func setupTest(t *testing.T) context.Context {
	// 初始化存储桶
	ctx := context.Background()
	ctx, err := InitBucket(ctx, testBucketName, testCredentialsFile)
	if err != nil {
		t.Skipf("跳过测试：初始化存储桶失败: %v", err)
	}

	SetBucketLifecycle(ctx, 7)

	return ctx
}

// 测试清理函数
func teardownTest() {
	Close()
}

// 测试初始化存储桶
func TestInitBucket(t *testing.T) {
	ctx := context.Background()
	_, err := InitBucket(ctx, testBucketName, "invalid-credentials.json")
	if err == nil {
		t.Error("期望初始化失败，但成功了")
	}

	// 测试有效的凭证文件
	_, err = InitBucket(ctx, testBucketName, testCredentialsFile)
	if err != nil {
		t.Fatalf("初始化存储桶失败: %v", err)
	}
}

// 创建测试用的临时文件
func createTempFile(t *testing.T, content string) *os.File {
	tmpfile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("关闭临时文件失败: %v", err)
	}
	return tmpfile
}

// waitForFileVisibility 轮询等待文件在GCS中可见
func waitForFileVisibility(ctx context.Context, t *testing.T, filename string) {
	const maxAttempts = 10
	const sleepDuration = 1 * time.Second

	for i := 0; i < maxAttempts; i++ {
		_, err := GetFileMetadata(ctx, filename)
		if err == nil {
			t.Logf("文件 %s 已可见", filename)
			return // 文件已可见
		}
		t.Logf("等待文件 %s 可见，尝试 %d/%d: %v", filename, i+1, maxAttempts, err)
		time.Sleep(sleepDuration)
	}
	t.Fatalf("文件 %s 在规定时间内未可见", filename)
}

// 测试上传文件
func TestUploadFile(t *testing.T) {
	ctx := setupTest(t)
	defer teardownTest()

	// 创建测试文件
	content := "测试文件内容"
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile.Name())

	// 重新打开文件用于读取
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("打开测试文件失败: %v", err)
	}
	defer file.Close()

	// 测试上传（设置1分钟过期）
	url, err := UploadFile(ctx, "test-upload.txt", file, 1)
	if err != nil {
		t.Fatalf("上传文件失败: %v", err)
	}

	// 验证返回的URL是否有效
	if url == "" {
		t.Error("返回的签名URL为空")
	}
	if !strings.Contains(url, "test-upload.txt") {
		t.Errorf("签名URL中不包含文件名: %s", url)
	}

	t.Logf("asdsa: %s", url)
}

// 测试下载文件
func TestDownloadFile(t *testing.T) {
	ctx := setupTest(t)
	defer teardownTest()

	// 为了下载文件，先上传一个文件
	content := "测试文件内容"
	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile.Name())
	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatalf("打开测试文件失败: %v", err)
	}
	defer file.Close()
	_, err = UploadFile(ctx, "test-download.txt", file, 0)
	if err != nil {
		t.Fatalf("上传下载测试文件失败: %v", err)
	}
	defer DeleteFile(ctx, "test-download.txt")
	waitForFileVisibility(ctx, t, "test-download.txt") // 确保文件可见

	// 测试下载
	reader, err := DownloadFile(ctx, "test-download.txt")
	if err != nil {
		t.Fatalf("下载文件失败: %v", err)
	}
	defer reader.Close()

	// 读取内容
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("读取下载内容失败: %v", err)
	}

	t.Logf("读取下载内容: %s", string(data))

	expected := "测试文件内容"
	if string(data) != expected {
		t.Errorf("下载内容不匹配，期望: %s, 实际: %s", expected, string(data))
	}
}

// 测试列出文件
func TestListFiles(t *testing.T) {
	ctx := setupTest(t)
	defer teardownTest()

	// 上传一些测试文件
	// files := []string{"test1.txt", "test2.txt", "test3.txt"}
	// for _, filename := range files {
	// 	content := strings.NewReader("测试内容 " + filename)
	// 	_, err := UploadFile(ctx, filename, content, 0) // 不设置过期时间
	// 	if err != nil {
	// 		t.Fatalf("上传测试文件失败: %v", err)
	// 	}
	// }

	// 测试列出文件
	listedFiles, err := ListFiles(ctx, "txt") // 修正前缀为 "" 以列出所有文件
	if err != nil {
		t.Fatalf("列出文件失败: %v", err)
	}

	fmt.Println(listedFiles)

	// 验证文件列表
	found := make(map[string]bool)
	for _, file := range listedFiles {
		found[file] = true
	}

	// for _, filename := range files {
	// 	if !found[filename] {
	// 		t.Errorf("未找到文件: %s", filename)
	// 	}
	// }

	// // 清理测试文件
	// for _, filename := range files {
	// 	err := DeleteFile(ctx, filename)
	// 	if err != nil {
	// 		t.Fatalf("清理测试文件失败: %v", err)
	// 	}
	// }
}
