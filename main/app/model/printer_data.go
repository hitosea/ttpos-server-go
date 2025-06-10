package model

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"strings"
)

// PrinterLogData 打印日志数据表 ttpos_printer_log_data
type PrinterLogData struct {
	BaseModel
	LogUuid uint64 `gorm:"column:log_uuid;type:bigint(20) unsigned;default:0;comment:打印日志UUID;NOT NULL" json:"log_uuid"`
	Data    string `gorm:"column:data;type:text;comment:打印数据" json:"data"`
}

// 压缩数据
func (model *PrinterLogData) CompressData() string {
	data := model.Data
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
func (model *PrinterLogData) DecompressData() string {
	if model.Data == "" {
		return ""
	}
	// 检查数据是否是压缩过的
	if !strings.HasPrefix(model.Data, "GZIP:") {
		return model.Data // 如果不是压缩过的数据，直接返回
	}
	// 提取Base64编码的部分
	encoded := strings.TrimPrefix(model.Data, "GZIP:")
	// 解码Base64
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return model.Data // 如果解码失败，返回原始数据
	}
	// 解压数据
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return model.Data
	}
	defer zr.Close()
	//
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return model.Data
	}
	return buf.String()
}
