package healthcheck

import (
	"context"
	"time"

	"github.com/go-seidon/provider/health"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
)

type HealthCheck interface {
	Check(ctx context.Context) (*CheckResult, *system.Error)
}

type CheckResult struct {
	Success system.Success
	Status  string
	Items   map[string]CheckResultItem
}

type CheckResultItem struct {
	Name      string
	Status    string
	Error     string
	Fatal     bool
	CheckedAt time.Time
}

type healthCheck struct {
	healthClient health.HealthCheck
}

func (h *healthCheck) Check(ctx context.Context) (*CheckResult, *system.Error) {
	checkRes, err := h.healthClient.Check(ctx)
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	items := map[string]CheckResultItem{}
	for key, item := range checkRes.Items {
		items[key] = CheckResultItem{
			Name:      item.Name,
			Status:    item.Status,
			Error:     item.Error,
			Fatal:     item.Fatal,
			CheckedAt: item.CheckedAt,
		}
	}

	res := &CheckResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success check health",
		},
		Status: checkRes.Status,
		Items:  items,
	}
	return res, nil
}

type HealthCheckParam struct {
	HealthClient health.HealthCheck
}

func NewHealthCheck(p HealthCheckParam) *healthCheck {
	return &healthCheck{
		healthClient: p.HealthClient,
	}
}
