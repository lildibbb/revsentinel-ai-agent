package ops

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
)

// QueueInspector is a minimal interface for queue inspection
type QueueInspector interface {
	Queues() ([]string, error)
	GetQueueInfo(queue string) (*asynq.QueueInfo, error)
	ListArchivedTasks(queue string) ([]*asynq.TaskInfo, error)
	RunTask(queue, id string) error
}

// AsynqInspectorAdapter wraps Asynq's Inspector to match QueueInspector interface
type AsynqInspectorAdapter struct {
	inspector *asynq.Inspector
}

// NewAsynqInspectorAdapter creates an adapter for Asynq's Inspector
func NewAsynqInspectorAdapter(inspector *asynq.Inspector) *AsynqInspectorAdapter {
	return &AsynqInspectorAdapter{inspector: inspector}
}

func (a *AsynqInspectorAdapter) Queues() ([]string, error) {
	return a.inspector.Queues()
}

func (a *AsynqInspectorAdapter) GetQueueInfo(queue string) (*asynq.QueueInfo, error) {
	return a.inspector.GetQueueInfo(queue)
}

func (a *AsynqInspectorAdapter) ListArchivedTasks(queue string) ([]*asynq.TaskInfo, error) {
	// ListArchivedTasks accepts optional ListOptions - we call it without options
	return a.inspector.ListArchivedTasks(queue)
}

func (a *AsynqInspectorAdapter) RunTask(queue, id string) error {
	return a.inspector.RunTask(queue, id)
}

// Handlers manages ops endpoints
type Handlers struct {
	inspector QueueInspector
}

// NewHandlers creates a new ops handlers instance
func NewHandlers(inspector QueueInspector) *Handlers {
	return &Handlers{
		inspector: inspector,
	}
}

// RegisterOpsRoutes mounts ops endpoints under /ops prefix
func RegisterOpsRoutes(router chi.Router, handlers *Handlers) {
	router.Route("/ops", func(r chi.Router) {
		r.Route("/queue", func(r chi.Router) {
			r.Get("/stats", handlers.QueueStats)
			r.Get("/dlq", handlers.DLQList)
			r.Post("/dlq/retry/{task_id}", handlers.RetryDLQTask)
		})
	})
}

// QueueStats returns statistics about all queues
// GET /ops/queue/stats
func (h *Handlers) QueueStats(w http.ResponseWriter, r *http.Request) {
	// Get list of all queues
	queues, err := h.inspector.Queues()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "queue_list_failed", err.Error())
		return
	}

	// Get info for each queue
	queueInfoList := make([]map[string]any, len(queues))
	for i, qname := range queues {
		info, err := h.inspector.GetQueueInfo(qname)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "queue_stats_failed", err.Error())
			return
		}
		queueInfoList[i] = map[string]any{
			"queue":     info.Queue,
			"size":      info.Size,
			"processed": info.Processed,
			"failed":    info.Failed,
			"paused":    info.Paused,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"queues": queueInfoList,
	})
}

// DLQList returns tasks in the dead letter queue (archived tasks)
// GET /ops/queue/dlq
func (h *Handlers) DLQList(w http.ResponseWriter, r *http.Request) {
	// Get list of archived tasks (DLQ)
	tasks, err := h.inspector.ListArchivedTasks("default")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "dlq_list_failed", err.Error())
		return
	}

	// Format tasks for response
	tasksList := make([]map[string]any, len(tasks))
	for i, task := range tasks {
		tasksList[i] = map[string]any{
			"id":    task.ID,
			"type":  task.Type,
			"queue": task.Queue,
			"state": task.State,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tasks": tasksList,
	})
}

// RetryDLQTask moves a task from DLQ back to the queue
// POST /ops/queue/dlq/retry/{task_id}
func (h *Handlers) RetryDLQTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")

	if taskID == "" {
		writeErr(w, http.StatusBadRequest, "task_id_required", "task_id path parameter is required")
		return
	}

	// Run the task (move from DLQ back to queue)
	if err := h.inspector.RunTask("default", taskID); err != nil {
		writeErr(w, http.StatusBadGateway, "dlq_retry_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"task_id": taskID,
	})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr writes an error response with the pattern: {error: "code", message: "description"}
func writeErr(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"code":    code,
		"message": message,
	})
}
