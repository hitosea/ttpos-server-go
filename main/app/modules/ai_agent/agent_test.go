package ai_agent

import (
	"testing"
)

func TestNewAgent(t *testing.T) {
	deps := &NodeDeps{
		Config: DefaultConfig(),
	}
	agent := NewAgent(deps)

	if agent == nil {
		t.Fatal("expected non-nil agent")
	}
	if agent.sessions == nil {
		t.Fatal("expected non-nil sessions map")
	}
}

func TestGetSession_NotFound(t *testing.T) {
	agent := NewAgent(&NodeDeps{Config: DefaultConfig()})

	_, ok := agent.GetSession("nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent session")
	}
}

func TestGetSession_Found(t *testing.T) {
	agent := NewAgent(&NodeDeps{Config: DefaultConfig()})

	// Manually insert a session
	agent.mu.Lock()
	agent.sessions["test-123"] = &ProcurementState{
		Status:        "awaiting_review",
		WarehouseUuid: 100,
	}
	agent.mu.Unlock()

	state, ok := agent.GetSession("test-123")
	if !ok {
		t.Fatal("expected ok=true for existing session")
	}
	if state.Status != "awaiting_review" {
		t.Errorf("expected status=awaiting_review, got %s", state.Status)
	}
	if state.WarehouseUuid != 100 {
		t.Errorf("expected warehouse_uuid=100, got %d", state.WarehouseUuid)
	}
}

func TestSubmitReview_SessionNotFound(t *testing.T) {
	agent := NewAgent(&NodeDeps{Config: DefaultConfig()})

	state := agent.SubmitReview(nil, "nonexistent", "approved", "")
	if state.Status != "failed" {
		t.Errorf("expected status=failed, got %s", state.Status)
	}
	if state.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestSubmitReview_WrongStatus(t *testing.T) {
	agent := NewAgent(&NodeDeps{Config: DefaultConfig()})

	agent.mu.Lock()
	agent.sessions["test-123"] = &ProcurementState{
		Status: "running",
	}
	agent.mu.Unlock()

	state := agent.SubmitReview(nil, "test-123", "approved", "")
	if state.Status != "failed" {
		t.Errorf("expected status=failed, got %s", state.Status)
	}
}

func TestSubmitReview_Rejected(t *testing.T) {
	agent := NewAgent(&NodeDeps{Config: DefaultConfig()})

	agent.mu.Lock()
	agent.sessions["test-123"] = &ProcurementState{
		Status:  "awaiting_review",
		StepLog: make([]string, 0),
	}
	agent.mu.Unlock()

	state := agent.SubmitReview(nil, "test-123", "rejected", "不需要采购")
	if state.Status != "completed" {
		t.Errorf("expected status=completed, got %s", state.Status)
	}
	if state.ReviewDecision != "rejected" {
		t.Errorf("expected decision=rejected, got %s", state.ReviewDecision)
	}
	if state.ReviewComment != "不需要采购" {
		t.Errorf("expected comment=不需要采购, got %s", state.ReviewComment)
	}
	if len(state.CreatedOrders) != 0 {
		t.Errorf("expected 0 created orders on rejection, got %d", len(state.CreatedOrders))
	}
}
