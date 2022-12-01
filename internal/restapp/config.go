package restapp

import (
	"fmt"
)

type RestAppConfig struct {
	AppName         string
	AppVersion      string
	AppHost         string
	AppPort         int
	StorageFormSize int64
}

func (c *RestAppConfig) GetAppName() string {
	return c.AppName
}

func (c *RestAppConfig) GetAppVersion() string {
	return c.AppVersion
}

func (c *RestAppConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.AppHost, c.AppPort)
}
