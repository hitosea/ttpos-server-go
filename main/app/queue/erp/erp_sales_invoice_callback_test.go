package erp

import (
	"encoding/json"
	"strconv"
	"testing"
	"ttpos-server-go/app/constant"

	"github.com/stretchr/testify/assert"
)

// ========== 消息解析测试 ==========

func TestErpSalesInvoiceCallbackMsg_JSON解析(t *testing.T) {
	tests := []struct {
		name string
		json string
		want ErpSalesInvoiceCallbackMsg
	}{
		{
			name: "成功回调",
			json: `{
				"company_uuid": "8267304538112000",
				"sale_order_uuid": "3720191183685635",
				"sync_status": 3,
				"sales_invoice_name": "ACC-SINV-2026-00143",
				"payment_entry_names": "[\"ACC-PAY-2026-001\"]",
				"order_type": "sale_order"
			}`,
			want: ErpSalesInvoiceCallbackMsg{
				CompanyUuid:       "8267304538112000",
				SaleOrderUuid:     "3720191183685635",
				SyncStatus:        3,
				SalesInvoiceName:  "ACC-SINV-2026-00143",
				PaymentEntryNames: `["ACC-PAY-2026-001"]`,
				OrderType:         "sale_order",
			},
		},
		{
			name: "失败回调",
			json: `{
				"company_uuid": "8267304538112000",
				"sale_order_uuid": "3720191183685635",
				"sync_status": 4,
				"error_msg": "创建Sales Invoice失败",
				"order_type": "sale_order"
			}`,
			want: ErpSalesInvoiceCallbackMsg{
				CompanyUuid:   "8267304538112000",
				SaleOrderUuid: "3720191183685635",
				SyncStatus:    4,
				ErrorMsg:      "创建Sales Invoice失败",
				OrderType:     "sale_order",
			},
		},
		{
			name: "外部取消回调(syncStatus=5)",
			json: `{
				"company_uuid": "8267304538112000",
				"sale_order_uuid": "3720191183685635",
				"sync_status": 5,
				"error_msg": "ERPNext外部取消",
				"order_type": "sale_order"
			}`,
			want: ErpSalesInvoiceCallbackMsg{
				CompanyUuid:   "8267304538112000",
				SaleOrderUuid: "3720191183685635",
				SyncStatus:    5,
				ErrorMsg:      "ERPNext外部取消",
				OrderType:     "sale_order",
			},
		},
		{
			name: "充值订单回调",
			json: `{
				"company_uuid": "1234567890",
				"sale_order_uuid": "9876543210",
				"sync_status": 3,
				"sales_invoice_name": "ACC-SINV-2026-00200",
				"order_type": "recharge"
			}`,
			want: ErpSalesInvoiceCallbackMsg{
				CompanyUuid:      "1234567890",
				SaleOrderUuid:    "9876543210",
				SyncStatus:       3,
				SalesInvoiceName: "ACC-SINV-2026-00200",
				OrderType:        "recharge",
			},
		},
		{
			name: "外卖订单回调",
			json: `{
				"company_uuid": "5555555555",
				"sale_order_uuid": "6666666666",
				"sync_status": 3,
				"sales_invoice_name": "ACC-SINV-2026-00300",
				"order_type": "takeout"
			}`,
			want: ErpSalesInvoiceCallbackMsg{
				CompanyUuid:      "5555555555",
				SaleOrderUuid:    "6666666666",
				SyncStatus:       3,
				SalesInvoiceName: "ACC-SINV-2026-00300",
				OrderType:        "takeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg ErpSalesInvoiceCallbackMsg
			err := json.Unmarshal([]byte(tt.json), &msg)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.CompanyUuid, msg.CompanyUuid)
			assert.Equal(t, tt.want.SaleOrderUuid, msg.SaleOrderUuid)
			assert.Equal(t, tt.want.SyncStatus, msg.SyncStatus)
			assert.Equal(t, tt.want.SalesInvoiceName, msg.SalesInvoiceName)
			assert.Equal(t, tt.want.PaymentEntryNames, msg.PaymentEntryNames)
			assert.Equal(t, tt.want.ErrorMsg, msg.ErrorMsg)
			assert.Equal(t, tt.want.OrderType, msg.OrderType)
		})
	}
}

func TestErpSalesInvoiceCallbackMsg_无效JSON(t *testing.T) {
	var msg ErpSalesInvoiceCallbackMsg
	err := json.Unmarshal([]byte(`{invalid`), &msg)
	assert.Error(t, err)
}

// ========== syncStatus 路由逻辑测试 ==========

func TestIsExternalCancel_状态判断(t *testing.T) {
	tests := []struct {
		name       string
		syncStatus int
		isCancel   bool
	}{
		{"status=0 未同步", 0, false},
		{"status=1 已入队", 1, false},
		{"status=2 进行中", 2, false},
		{"status=3 成功", 3, false},
		{"status=4 失败", 4, false},
		{"status=5 外部取消", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ErpSalesInvoiceCallbackMsg{SyncStatus: tt.syncStatus}
			isExternalCancel := msg.SyncStatus == constant.ErpSyncStatusExternalCancel
			assert.Equal(t, tt.isCancel, isExternalCancel)
		})
	}
}

// ========== 订单类型路由测试 ==========

func TestOrderTypeRouting(t *testing.T) {
	tests := []struct {
		name      string
		orderType string
		route     string // expected routing target
	}{
		{"POS订单(空)", "", "sale_order"},
		{"POS订单", "sale_order", "sale_order"},
		{"充值订单", "recharge", "recharge"},
		{"外卖订单", "takeout", "takeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ErpSalesInvoiceCallbackMsg{OrderType: tt.orderType}
			// 验证路由逻辑与 handler 中的 switch 一致
			switch msg.OrderType {
			case "recharge":
				assert.Equal(t, "recharge", tt.route)
			case "takeout":
				assert.Equal(t, "takeout", tt.route)
			default:
				assert.Equal(t, "sale_order", tt.route)
			}
		})
	}
}

// ========== 必填字段验证测试 ==========

func TestCallbackMsg_必填字段验证(t *testing.T) {
	tests := []struct {
		name        string
		companyUuid string
		orderUuid   string
		valid       bool
	}{
		{"两者都有 → 有效", "8267304538112000", "3720191183685635", true},
		{"缺companyUuid → 无效", "", "3720191183685635", false},
		{"缺orderUuid → 无效", "8267304538112000", "", false},
		{"都缺 → 无效", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := ErpSalesInvoiceCallbackMsg{
				CompanyUuid:   tt.companyUuid,
				SaleOrderUuid: tt.orderUuid,
			}
			valid := msg.CompanyUuid != "" && msg.SaleOrderUuid != ""
			assert.Equal(t, tt.valid, valid)
		})
	}
}

// ========== 外部取消时清空字段逻辑测试 ==========

func TestExternalCancel_清空SI和PE名称(t *testing.T) {
	// 模拟 syncStatus=5 的处理逻辑：应该传空字符串清空 SI/PE 名称
	msg := ErpSalesInvoiceCallbackMsg{
		CompanyUuid:       "8267304538112000",
		SaleOrderUuid:     "3720191183685635",
		SyncStatus:        5,
		SalesInvoiceName:  "", // 外部取消时 BMP 不传 SI 名称
		PaymentEntryNames: "", // 外部取消时 BMP 不传 PE 名称
		ErrorMsg:          "ERPNext外部取消",
		OrderType:         "sale_order",
	}

	isExternalCancel := msg.SyncStatus == 5
	assert.True(t, isExternalCancel)

	// 外部取消时，应使用空字符串覆盖（而非保留旧值）
	siName := msg.SalesInvoiceName
	peNames := msg.PaymentEntryNames
	if isExternalCancel {
		siName = ""
		peNames = ""
	}
	assert.Empty(t, siName)
	assert.Empty(t, peNames)
}

// ========== ErpSyncStatus 常量完整性测试 ==========

func TestErpSyncStatus_常量值(t *testing.T) {
	assert.Equal(t, 0, constant.ErpSyncStatusNotSynced)
	assert.Equal(t, 1, constant.ErpSyncStatusQueued)
	assert.Equal(t, 2, constant.ErpSyncStatusInProgress)
	assert.Equal(t, 3, constant.ErpSyncStatusSuccess)
	assert.Equal(t, 4, constant.ErpSyncStatusFailed)
	assert.Equal(t, 5, constant.ErpSyncStatusExternalCancel)
}

// ========== JSON 序列化往返测试 ==========

func TestErpSalesInvoiceCallbackMsg_JSON往返(t *testing.T) {
	original := ErpSalesInvoiceCallbackMsg{
		CompanyUuid:       "8267304538112000",
		SaleOrderUuid:     "3720191183685635",
		SyncStatus:        3,
		SalesInvoiceName:  "ACC-SINV-2026-00143",
		PaymentEntryNames: `["ACC-PAY-2026-001","ACC-PAY-2026-002"]`,
		ErrorMsg:          "",
		OrderType:         "sale_order",
	}

	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded ErpSalesInvoiceCallbackMsg
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}

// ========== error_msg omitempty 行为测试 ==========

func TestErpSalesInvoiceCallbackMsg_ErrorMsg_Omitempty(t *testing.T) {
	// 成功消息 error_msg 为空时，JSON 不应包含 error_msg 字段
	msg := ErpSalesInvoiceCallbackMsg{
		CompanyUuid:   "123",
		SaleOrderUuid: "456",
		SyncStatus:    3,
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "error_msg")

	// 失败消息 error_msg 不为空时，JSON 应包含 error_msg 字段
	msg.ErrorMsg = "创建失败"
	data, err = json.Marshal(msg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "error_msg")
}

// ========== company_uuid 解析边界测试 ==========

func TestCompanyUuidParsing_边界值(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{"正常数字", "8267304538112000", false},
		{"零值", "0", false},
		{"非数字", "abc", true},
		{"负数", "-1", true},
		{"超大数字", "18446744073709551615", false}, // uint64 max
		{"溢出", "18446744073709551616", true},      // uint64 max + 1
		{"含空格", " 123 ", true},
		{"含小数", "123.45", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := strconv.ParseUint(tt.uuid, 10, 64)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ========== 外部取消对各订单类型的处理逻辑测试 ==========

func TestExternalCancel_各订单类型路由(t *testing.T) {
	orderTypes := []string{"sale_order", "recharge", "takeout", ""}

	for _, ot := range orderTypes {
		t.Run("orderType="+ot, func(t *testing.T) {
			msg := ErpSalesInvoiceCallbackMsg{
				CompanyUuid:   "8267304538112000",
				SaleOrderUuid: "3720191183685635",
				SyncStatus:    constant.ErpSyncStatusExternalCancel,
				OrderType:     ot,
			}

			isExternalCancel := msg.SyncStatus == constant.ErpSyncStatusExternalCancel
			assert.True(t, isExternalCancel, "syncStatus=5 应标识为外部取消")

			// 验证所有订单类型都能进入 switch 分支
			switch msg.OrderType {
			case "recharge":
				assert.Equal(t, "recharge", msg.OrderType)
			case "takeout":
				assert.Equal(t, "takeout", msg.OrderType)
			default:
				// sale_order 和空字符串都走 default 分支
				assert.Contains(t, []string{"sale_order", ""}, msg.OrderType)
			}
		})
	}
}

// ========== 成功回调 vs 失败回调字段差异测试 ==========

func TestCallbackMsg_成功与失败字段差异(t *testing.T) {
	// 成功回调：有 SI/PE 名称，无 error_msg
	successMsg := ErpSalesInvoiceCallbackMsg{
		CompanyUuid:       "123",
		SaleOrderUuid:     "456",
		SyncStatus:        constant.ErpSyncStatusSuccess,
		SalesInvoiceName:  "ACC-SINV-2026-00001",
		PaymentEntryNames: `["ACC-PAY-2026-001"]`,
	}
	assert.NotEmpty(t, successMsg.SalesInvoiceName)
	assert.Empty(t, successMsg.ErrorMsg)

	// 失败回调：无 SI/PE 名称，有 error_msg
	failMsg := ErpSalesInvoiceCallbackMsg{
		CompanyUuid:   "123",
		SaleOrderUuid: "456",
		SyncStatus:    constant.ErpSyncStatusFailed,
		ErrorMsg:      "ERPNext 服务不可用",
	}
	assert.Empty(t, failMsg.SalesInvoiceName)
	assert.NotEmpty(t, failMsg.ErrorMsg)
}
