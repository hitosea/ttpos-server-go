package ai_agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service"
	purchaseorder "ttpos-server-go/app/service/purchase_order"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// NodeDeps holds the service dependencies injected into workflow nodes.
type NodeDeps struct {
	DBM              *database.DBManager
	MaterialSrv      service.IMaterialSrv
	WarehouseSrv     service.IWarehouseSrv
	SupplierSrv      service.ISupplierSrv
	PurchaseOrderSrv purchaseorder.IPurchaseOrderSrv
	StatisticsSrv    service.IStatisticsSrv
	LLM              *LLMClient
	Config           Config
}

// --- Node 1: Collect Data ---

func collectData(ctx context.Context, state *ProcurementState, deps *NodeDeps) {
	if state.WarehouseUuid == 0 {
		state.Error = "warehouse_uuid is required"
		state.StepLog = append(state.StepLog, "collect_data: FAILED - no warehouse_uuid")
		return
	}

	db := ctx.GetDB()

	// 1. Warehouse materials (stock data) via service
	whReq := req.WarehouseMaterialListReq{WarehouseUuid: state.WarehouseUuid}
	whReq.PageNo = 1
	whReq.PageSize = 500
	whResp, err := deps.WarehouseSrv.GetWarehouseMaterialList(ctx, whReq)
	if err != nil {
		state.Error = fmt.Sprintf("get warehouse materials: %v", err)
		state.StepLog = append(state.StepLog, "collect_data: FAILED - "+state.Error)
		return
	}
	state.StepLog = append(state.StepLog, fmt.Sprintf("collect_data: fetched %d warehouse stock items", len(whResp.List)))

	// 2. Material master data via repository (need SupplierUuid/SupplierErpCode not exposed in service response)
	materialRepo := repository.NewMaterialRepo(db)
	materials, _, err := materialRepo.GetMaterialListWithPagination(1, 500,
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		state.Error = fmt.Sprintf("get materials: %v", err)
		state.StepLog = append(state.StepLog, "collect_data: FAILED - "+state.Error)
		return
	}
	state.StepLog = append(state.StepLog, fmt.Sprintf("collect_data: fetched %d material records", len(materials)))

	// 3. Build material lookup (uuid → extra fields)
	type matExtra struct {
		SafetyStock      float64
		SupplierUuid     uint64
		SupplierErpCode  string
		PurchaseUnitUuid uint64
	}
	matMap := make(map[uint64]matExtra, len(materials))
	for _, m := range materials {
		ss := 0.0
		if m.SafetyStock != nil {
			ss = *m.SafetyStock
		}
		matMap[m.Uuid] = matExtra{
			SafetyStock:      ss,
			SupplierUuid:     m.SupplierUuid,
			SupplierErpCode:  m.SupplierErpCode,
			PurchaseUnitUuid: m.PurchaseUnitUuid,
		}
	}

	// 4. Merge warehouse stock + material master data
	enriched := make([]MaterialInfo, 0, len(whResp.List))
	for _, wh := range whResp.List {
		extra := matMap[wh.MaterialUuid]
		nameZH := wh.MaterialName.ZH
		enriched = append(enriched, MaterialInfo{
			MaterialUuid:     wh.MaterialUuid,
			MaterialCode:     wh.MaterialCode,
			MaterialNameZH:   nameZH,
			BookedQuantity:   wh.BookedQuantity,
			SafetyStock:      extra.SafetyStock,
			SupplierUuid:     extra.SupplierUuid,
			SupplierErpCode:  extra.SupplierErpCode,
			PurchaseUnitUuid: extra.PurchaseUnitUuid,
		})
	}
	state.Materials = enriched
	state.StepLog = append(state.StepLog, fmt.Sprintf("collect_data: merged into %d enriched records", len(state.Materials)))

	// 5. Product sales (last 7 days)
	salesReq := service.CountReq{
		TimeType: 5, // 近7天
		PageNo:   1,
		PageSize: 200,
	}
	salesResp := deps.StatisticsSrv.CountProductSale(ctx, salesReq)
	state.SalesData = make([]ProductSaleInfo, 0, len(salesResp.Data))
	for _, s := range salesResp.Data {
		state.SalesData = append(state.SalesData, ProductSaleInfo{
			ProductName:  s.ProductName,
			CategoryName: s.CategoryName,
			SalesNum:     s.TotalSaleNum,
			SalesAmount:  s.TotalActualSaleAmount,
		})
	}
	state.StepLog = append(state.StepLog, fmt.Sprintf("collect_data: fetched %d product sales (last 7 days)", len(state.SalesData)))

	// 6. Suppliers via repository (resp.SupplierInfo lacks ErpCode)
	supplierRepo := repository.NewSupplierRepo(db)
	suppliers, err := supplierRepo.GetList(repository.CommonRepo.WhereBySoftDelete())
	if err != nil {
		logger.Logger.Warn("collect_data: failed to get suppliers", zap.Error(err))
	} else {
		state.Suppliers = make([]SupplierInfo, 0, len(suppliers))
		for _, s := range suppliers {
			state.Suppliers = append(state.Suppliers, SupplierInfo{
				Uuid:    s.Uuid,
				Name:    s.Name,
				Code:    s.Code,
				ErpCode: s.ErpCode,
			})
		}
	}
	state.StepLog = append(state.StepLog, fmt.Sprintf("collect_data: fetched %d suppliers", len(state.Suppliers)))
}

// --- Node 2: Forecast Demand (LLM) ---

const forecastSystemPrompt = `You are an inventory demand forecasting assistant for a restaurant POS system.

You are given:
1. A list of materials with current stock and safety stock levels
2. Recent product sales data (last 7 days) — these are finished dishes, not raw materials

Your task: predict material demand for the next %d days.

## How to estimate demand

- Use the product sales data to understand consumption velocity
- Materials used in high-selling products will have higher demand
- If a material's name matches or relates to a product name, use that product's daily sales rate
  to estimate material consumption
- For materials without clear product matches, estimate based on the material type

## Calculations

For each material:
1. predicted_demand: estimated total consumption over %d days
2. shortage: max(0, predicted_demand - current_stock + safety_buffer)
   where safety_buffer = safety_stock * %.1f
   If safety_stock = 0, use 20%% of predicted_demand as safety_buffer

## Output format

Return a JSON array with these fields:
- material_uuid (int)
- material_code (string)
- material_name_zh (string)
- current_stock (float)
- safety_stock (float)
- predicted_demand (float)
- shortage (float)

ONLY include materials where shortage > 0.
Return an empty array [] if nothing needs restocking.
Return ONLY the JSON array, no markdown fences or explanation.`

func forecastDemand(state *ProcurementState, deps *NodeDeps) {
	if len(state.Materials) == 0 {
		state.StepLog = append(state.StepLog, "forecast_demand: no materials to forecast")
		state.NeedsPurchase = false
		return
	}

	days := state.ForecastDays
	if days == 0 {
		days = deps.Config.ForecastDays
	}

	// Prepare material summary for LLM
	type matSummary struct {
		MaterialUuid   uint64  `json:"material_uuid"`
		MaterialCode   string  `json:"material_code"`
		MaterialNameZH string  `json:"material_name_zh"`
		CurrentStock   float64 `json:"current_stock"`
		SafetyStock    float64 `json:"safety_stock"`
	}
	mats := make([]matSummary, 0, len(state.Materials))
	for _, m := range state.Materials {
		mats = append(mats, matSummary{
			MaterialUuid:   m.MaterialUuid,
			MaterialCode:   m.MaterialCode,
			MaterialNameZH: m.MaterialNameZH,
			CurrentStock:   m.BookedQuantity,
			SafetyStock:    m.SafetyStock,
		})
	}

	contextData, _ := json.Marshal(map[string]any{
		"materials":                  mats,
		"recent_product_sales_7days": state.SalesData,
	})

	sysPrompt := fmt.Sprintf(forecastSystemPrompt, days, days, deps.Config.SafetyStockThreshold)

	response, err := deps.LLM.Chat(sysPrompt, string(contextData))
	if err != nil {
		state.Error = fmt.Sprintf("LLM forecast error: %v", err)
		state.StepLog = append(state.StepLog, "forecast_demand: FAILED - "+state.Error)
		return
	}

	// Parse LLM response (strip markdown fences if present)
	content := strings.TrimSpace(response)
	if strings.HasPrefix(content, "```") {
		if idx := strings.Index(content[3:], "\n"); idx >= 0 {
			content = content[3+idx+1:]
		}
		if idx := strings.LastIndex(content, "```"); idx >= 0 {
			content = content[:idx]
		}
		content = strings.TrimSpace(content)
	}

	var forecasts []ForecastItem
	if err := json.Unmarshal([]byte(content), &forecasts); err != nil {
		truncated := content
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		logger.Logger.Warn("forecast_demand: failed to parse LLM response",
			zap.String("response", truncated),
			zap.Error(err))
		state.Forecasts = nil
	} else {
		state.Forecasts = forecasts
	}

	state.NeedsPurchase = len(state.Forecasts) > 0
	state.StepLog = append(state.StepLog, fmt.Sprintf(
		"forecast_demand: analyzed %d materials + %d product sales, %d need restocking (next %d days)",
		len(state.Materials), len(state.SalesData), len(state.Forecasts), days))
}

// --- Node 3: Compare Stock (deterministic validation) ---

func compareStock(state *ProcurementState) {
	if len(state.Forecasts) == 0 {
		state.NeedsPurchase = false
		state.StepLog = append(state.StepLog, "compare_stock: no forecasts to validate")
		return
	}

	matMap := make(map[uint64]MaterialInfo, len(state.Materials))
	for _, m := range state.Materials {
		matMap[m.MaterialUuid] = m
	}

	validated := make([]ForecastItem, 0)
	for _, f := range state.Forecasts {
		mat, ok := matMap[f.MaterialUuid]
		if !ok {
			continue
		}

		actualStock := mat.BookedQuantity
		safetyStock := mat.SafetyStock
		demand := f.PredictedDemand

		// Recalculate shortage using actual stock
		safetyBuffer := safetyStock
		if safetyStock == 0 {
			safetyBuffer = demand * 0.2
		}
		shortage := demand - actualStock + safetyBuffer
		if shortage <= 0 {
			continue
		}

		f.CurrentStock = actualStock
		f.SafetyStock = safetyStock
		f.Shortage = shortage
		f.OrderQuantity = shortage
		validated = append(validated, f)
	}

	state.Forecasts = validated
	state.NeedsPurchase = len(validated) > 0
	state.StepLog = append(state.StepLog, fmt.Sprintf("compare_stock: validated %d items need purchase", len(validated)))
}

// --- Node 4: Match Supplier ---

func matchSupplier(state *ProcurementState) {
	matMap := make(map[uint64]MaterialInfo, len(state.Materials))
	for _, m := range state.Materials {
		matMap[m.MaterialUuid] = m
	}

	supMap := make(map[uint64]SupplierInfo, len(state.Suppliers))
	for _, s := range state.Suppliers {
		supMap[s.Uuid] = s
	}

	matched := 0
	for i, f := range state.Forecasts {
		mat, ok := matMap[f.MaterialUuid]
		if !ok || mat.SupplierUuid == 0 {
			continue
		}
		sup, ok := supMap[mat.SupplierUuid]
		if !ok {
			continue
		}
		state.Forecasts[i].SupplierName = sup.Name
		state.Forecasts[i].SupplierErpCode = sup.ErpCode
		matched++
	}

	// Assign default supplier to unmatched items
	if len(state.Suppliers) > 0 {
		defaultSup := state.Suppliers[0]
		for i, f := range state.Forecasts {
			if f.SupplierName == "" {
				state.Forecasts[i].SupplierName = defaultSup.Name
				state.Forecasts[i].SupplierErpCode = defaultSup.ErpCode
			}
		}
	}

	state.StepLog = append(state.StepLog, fmt.Sprintf(
		"match_supplier: %d/%d matched to known suppliers", matched, len(state.Forecasts)))
}

// --- Node 5: Generate Proposal ---

func generateProposal(state *ProcurementState) {
	if len(state.Forecasts) == 0 {
		state.Proposals = make([]PurchaseProposal, 0)
		state.StepLog = append(state.StepLog, "generate_proposal: no items to propose")
		return
	}

	groups := make(map[string]*PurchaseProposal)
	for _, f := range state.Forecasts {
		key := f.SupplierName
		if key == "" {
			key = "未知供应商"
		}
		if _, ok := groups[key]; !ok {
			groups[key] = &PurchaseProposal{
				SupplierName:    f.SupplierName,
				SupplierErpCode: f.SupplierErpCode,
				Items:           make([]ProposalItem, 0),
			}
		}
		p := groups[key]
		p.Items = append(p.Items, ProposalItem{
			MaterialUuid:   f.MaterialUuid,
			MaterialCode:   f.MaterialCode,
			MaterialNameZH: f.MaterialNameZH,
			OrderQuantity:  f.OrderQuantity,
		})
		p.TotalQuantity += f.OrderQuantity
	}

	proposals := make([]PurchaseProposal, 0, len(groups))
	for _, p := range groups {
		proposals = append(proposals, *p)
	}
	state.Proposals = proposals
	state.StepLog = append(state.StepLog, fmt.Sprintf(
		"generate_proposal: created %d proposals for %d items", len(proposals), len(state.Forecasts)))
}

// --- Node 6: Human Review ---

func humanReview(state *ProcurementState) {
	state.Status = "awaiting_review"
	state.StepLog = append(state.StepLog, "human_review: awaiting review decision")
}

// --- Node 7: Create Purchase Orders ---

func createPurchaseOrders(ctx context.Context, state *ProcurementState, deps *NodeDeps) {
	if state.ReviewDecision != "approved" {
		state.StepLog = append(state.StepLog, fmt.Sprintf(
			"create_po: skipped (decision=%s)", state.ReviewDecision))
		return
	}

	// Build material lookup for purchase unit uuid
	matMap := make(map[uint64]MaterialInfo, len(state.Materials))
	for _, m := range state.Materials {
		matMap[m.MaterialUuid] = m
	}

	created := make([]CreatedOrder, 0)
	for _, proposal := range state.Proposals {
		if len(proposal.Items) == 0 {
			continue
		}

		items := make([]req.PurchaseOrderItemCreateReq, 0, len(proposal.Items))
		for _, item := range proposal.Items {
			unitUuid := item.UnitUuid
			if unitUuid == 0 {
				if mat, ok := matMap[item.MaterialUuid]; ok && mat.PurchaseUnitUuid > 0 {
					unitUuid = mat.PurchaseUnitUuid
				}
			}
			unitList := []req.PurchaseOrderItemMaterialUnitReq{
				{Uuid: unitUuid, Num: item.OrderQuantity},
			}
			items = append(items, req.PurchaseOrderItemCreateReq{
				MaterialUuid: item.MaterialUuid,
				UnitList:     unitList,
			})
		}

		createReq := req.PurchaseOrderCreateReq{
			OrderTime:       time.Now().Unix(),
			SupplierName:    proposal.SupplierName,
			SupplierErpCode: proposal.SupplierErpCode,
			Items:           items,
			PurchaseType:    1, // 外部采购
		}

		poResp, err := deps.PurchaseOrderSrv.CreatePurchaseOrder(ctx, createReq)
		if err != nil {
			logger.Logger.Warn("create_po: failed to create order",
				zap.String("supplier", proposal.SupplierName),
				zap.Error(err))
			continue
		}
		created = append(created, CreatedOrder{
			Uuid:    poResp.Uuid,
			OrderNo: poResp.OrderNo,
		})
	}

	state.CreatedOrders = created
	state.StepLog = append(state.StepLog, fmt.Sprintf(
		"create_po: created %d/%d purchase orders", len(created), len(state.Proposals)))
}

// --- Node 8: Detect Anomalies ---

func detectAnomalies(state *ProcurementState) {
	anomalies := make([]AnomalyInfo, 0)

	for _, m := range state.Materials {
		stock := m.BookedQuantity
		safety := m.SafetyStock

		if stock == 0 && safety > 0 {
			anomalies = append(anomalies, AnomalyInfo{
				MaterialUuid:   m.MaterialUuid,
				MaterialCode:   m.MaterialCode,
				MaterialNameZH: m.MaterialNameZH,
				AnomalyType:    "zero_stock",
				Severity:       "high",
				CurrentStock:   stock,
				SafetyStock:    safety,
				Message:        fmt.Sprintf("%s 库存为零 (安全库存: %.1f)", m.MaterialNameZH, safety),
			})
		} else if safety > 0 && stock < safety {
			anomalies = append(anomalies, AnomalyInfo{
				MaterialUuid:   m.MaterialUuid,
				MaterialCode:   m.MaterialCode,
				MaterialNameZH: m.MaterialNameZH,
				AnomalyType:    "below_safety_stock",
				Severity:       "medium",
				CurrentStock:   stock,
				SafetyStock:    safety,
				Message:        fmt.Sprintf("%s 库存 %.1f 低于安全库存 %.1f", m.MaterialNameZH, stock, safety),
			})
		}
	}

	state.Anomalies = anomalies
	state.StepLog = append(state.StepLog, fmt.Sprintf("detect_anomaly: found %d anomalies", len(anomalies)))
}
