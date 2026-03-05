package ai_agent

import (
	"sync"

	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// Agent orchestrates the procurement analysis workflow.
type Agent struct {
	deps *NodeDeps

	// In-memory session store (PoC: replace with Redis/DB for production)
	mu       sync.RWMutex
	sessions map[string]*ProcurementState
}

// NewAgent creates a new procurement agent.
func NewAgent(deps *NodeDeps) *Agent {
	return &Agent{
		deps:     deps,
		sessions: make(map[string]*ProcurementState),
	}
}

// RunAnalysis executes the procurement workflow up to the human review checkpoint.
// Returns the session ID and the current state (paused at awaiting_review or completed).
func (a *Agent) RunAnalysis(ctx context.Context, sessionID string, warehouseUuid uint64, forecastDays int) *ProcurementState {
	state := &ProcurementState{
		WarehouseUuid: warehouseUuid,
		ForecastDays:  forecastDays,
		Status:        "running",
		StepLog:       make([]string, 0),
	}

	// Store session
	a.mu.Lock()
	a.sessions[sessionID] = state
	a.mu.Unlock()

	// Phase 1: Collect data
	collectData(ctx, state, a.deps)
	if state.Error != "" {
		state.Status = "failed"
		return state
	}

	// Phase 2: Forecast demand (LLM)
	forecastDemand(state, a.deps)
	if state.Error != "" {
		state.Status = "failed"
		return state
	}

	// Phase 3: Compare stock (deterministic validation)
	compareStock(state)

	// Phase 4: Detect anomalies (always runs)
	detectAnomalies(state)

	// If no purchase needed, complete
	if !state.NeedsPurchase {
		state.Status = "completed"
		logger.Logger.Info("ai_agent: analysis complete, no purchase needed",
			zap.String("session_id", sessionID))
		return state
	}

	// Phase 5: Match supplier
	matchSupplier(state)

	// Phase 6: Generate proposal
	generateProposal(state)

	// Phase 7: Pause for human review
	humanReview(state)

	logger.Logger.Info("ai_agent: analysis complete, awaiting review",
		zap.String("session_id", sessionID),
		zap.Int("proposals", len(state.Proposals)))

	return state
}

// SubmitReview processes the human review decision and optionally creates purchase orders.
func (a *Agent) SubmitReview(ctx context.Context, sessionID string, decision string, comment string) *ProcurementState {
	a.mu.RLock()
	state, ok := a.sessions[sessionID]
	a.mu.RUnlock()

	if !ok {
		return &ProcurementState{
			Error:  "session not found: " + sessionID,
			Status: "failed",
		}
	}

	if state.Status != "awaiting_review" {
		return &ProcurementState{
			Error:  "session is not awaiting review (status=" + state.Status + ")",
			Status: "failed",
		}
	}

	state.ReviewDecision = decision
	state.ReviewComment = comment
	state.Status = "running"

	if decision == "approved" {
		createPurchaseOrders(ctx, state, a.deps)
	}

	state.Status = "completed"
	state.StepLog = append(state.StepLog, "workflow: completed (decision="+decision+")")

	logger.Logger.Info("ai_agent: review processed",
		zap.String("session_id", sessionID),
		zap.String("decision", decision),
		zap.Int("orders_created", len(state.CreatedOrders)))

	return state
}

// GetSession returns the state for a given session ID.
func (a *Agent) GetSession(sessionID string) (*ProcurementState, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	state, ok := a.sessions[sessionID]
	return state, ok
}
