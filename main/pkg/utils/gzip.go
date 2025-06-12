package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"strings"
)

// 压缩数据
func GzipCompressData(data string) string {
	if data == "" {
		return data
	}
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression) // 或使用gzip.DefaultCompression
	if err != nil {
		return ""
	}
	_, err = zw.Write([]byte(data))
	if err != nil {
		return ""
	}
	if err := zw.Close(); err != nil {
		return ""
	}
	// 将压缩后的二进制数据转换为Base64编码的字符串
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "GZIP:" + encoded

}

// 还原压缩的数据
func GzipDecompressData(data string) string {
	if data == "" {
		return ""
	}
	// 检查数据是否是压缩过的
	if !strings.HasPrefix(data, "GZIP:") {
		return data // 如果不是压缩过的数据，直接返回
	}
	// 提取Base64编码的部分
	encoded := strings.TrimPrefix(data, "GZIP:")
	// 解码Base64
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return data // 如果解码失败，返回原始数据
	}
	// 解压数据
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return data
	}
	defer zr.Close()
	//
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return data
	}
	return buf.String()
}
