package ops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
)

// MockInspector implements QueueInspector interface for testing
type MockInspector struct {
	queuesFunc         func() ([]string, error)
	getQueueInfoFunc   func(string) (*asynq.QueueInfo, error)
	listArchivedFunc   func(string) ([]*asynq.TaskInfo, error)
	runTaskFunc        func(string, string) error
}

func (m *MockInspector) Queues() ([]string, error) {
	if m.queuesFunc != nil {
		return m.queuesFunc()
	}
	return []string{"default"}, nil
}

func (m *MockInspector) GetQueueInfo(queue string) (*asynq.QueueInfo, error) {
	if m.getQueueInfoFunc != nil {
		return m.getQueueInfoFunc(queue)
	}
	return &asynq.QueueInfo{Queue: queue}, nil
}

func (m *MockInspector) ListArchivedTasks(queue string) ([]*asynq.TaskInfo, error) {
	if m.listArchivedFunc != nil {
		return m.listArchivedFunc(queue)
	}
	return []*asynq.TaskInfo{}, nil
}

func (m *MockInspector) RunTask(queue, taskID string) error {
	if m.runTaskFunc != nil {
		return m.runTaskFunc(queue, taskID)
	}
	return nil
}

func setupOpsRouterWithMockInspector(inspector QueueInspector) *chi.Mux {
	handlers := NewHandlers(inspector)
	router := chi.NewRouter()
	RegisterOpsRoutes(router, handlers)
	return router
}

func TestQueueStatsEndpoint_ReturnsJSON(t *testing.T) {
	mockInspector := &MockInspector{
		getQueueInfoFunc: func(queue string) (*asynq.QueueInfo, error) {
			return &asynq.QueueInfo{
				Queue:     queue,
				Size:      5,
				Processed: 100,
				Failed:    2,
				Paused:    false,
			}, nil
		},
	}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodGet, "/ops/queue/stats", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["queues"]; !ok {
		t.Fatal("response missing 'queues' key")
	}
}

func TestQueueStatsEndpoint_InspectorError(t *testing.T) {
	mockInspector := &MockInspector{
		getQueueInfoFunc: func(queue string) (*asynq.QueueInfo, error) {
			return nil, fmt.Errorf("connection error")
		},
	}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodGet, "/ops/queue/stats", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if code, ok := resp["code"].(string); !ok || code == "" {
		t.Fatal("response missing error code")
	}
}

func TestDLQListEndpoint_ReturnsList(t *testing.T) {
	mockInspector := &MockInspector{
		listArchivedFunc: func(queue string) ([]*asynq.TaskInfo, error) {
			return []*asynq.TaskInfo{
				{
					ID:    "task1",
					Type:  "process_event.v1",
					Queue: queue,
					State: asynq.TaskStateArchived,
				},
				{
					ID:    "task2",
					Type:  "process_event.v1",
					Queue: queue,
					State: asynq.TaskStateArchived,
				},
			}, nil
		},
	}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodGet, "/ops/queue/dlq", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["tasks"]; !ok {
		t.Fatal("response missing 'tasks' key")
	}
}

func TestDLQListEndpoint_InspectorError(t *testing.T) {
	mockInspector := &MockInspector{
		listArchivedFunc: func(queue string) ([]*asynq.TaskInfo, error) {
			return nil, fmt.Errorf("connection error")
		},
	}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodGet, "/ops/queue/dlq", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestRetryDLQTaskEndpoint_Success(t *testing.T) {
	mockInspector := &MockInspector{
		runTaskFunc: func(queue, taskID string) error {
			if taskID != "task1" {
				return fmt.Errorf("task not found")
			}
			return nil
		},
	}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodPost, "/ops/queue/dlq/retry/task1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if ok, exists := resp["ok"].(bool); !exists || !ok {
		t.Fatal("response missing or false 'ok' key")
	}
	if taskID, exists := resp["task_id"].(string); !exists || taskID != "task1" {
		t.Fatal("response missing or incorrect 'task_id' key")
	}
}

func TestRetryDLQTaskEndpoint_MissingTaskID(t *testing.T) {
	mockInspector := &MockInspector{}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodPost, "/ops/queue/dlq/retry/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// chi returns 404 for unmatched route (no task_id param)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRetryDLQTaskEndpoint_InspectorError(t *testing.T) {
	mockInspector := &MockInspector{
		runTaskFunc: func(queue, taskID string) error {
			return fmt.Errorf("task not found")
		},
	}

	r := setupOpsRouterWithMockInspector(mockInspector)
	req := httptest.NewRequest(http.MethodPost, "/ops/queue/dlq/retry/task1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rec.Code)
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if code, ok := resp["code"].(string); !ok || code == "" {
		t.Fatal("response missing error code")
	}
}
