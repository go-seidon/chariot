package app

import (
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/health"
	"github.com/go-seidon/provider/health/job"
	"github.com/go-seidon/provider/logging"
)

func NewDefaultHealthCheck(logger logging.Logger, repo repository.Repository) (health.HealthCheck, error) {

	inetPingJob, err := job.NewHttpPing(job.HttpPingParam{
		Name:     "internet-connection",
		Interval: 30 * time.Second,
		Url:      "https://google.com",
	})
	if err != nil {
		return nil, err
	}

	appDiskJob, err := job.NewDiskUsage(job.DiskUsageParam{
		Name:      "app-disk",
		Interval:  10 * time.Minute,
		Directory: "/",
	})
	if err != nil {
		return nil, err
	}

	repoPingJob, err := job.NewRepoPing(job.RepoPingParam{
		Name:       "repository-connection",
		Interval:   15 * time.Minute,
		DataSource: repo,
	})
	if err != nil {
		return nil, err
	}

	healthClient := health.NewHealthCheck(
		health.WithLogger(logger),
		health.AddJob(inetPingJob),
		health.AddJob(appDiskJob),
		health.AddJob(repoPingJob),
	)

	return healthClient, nil
}
