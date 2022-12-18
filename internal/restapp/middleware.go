package restapp

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-seidon/provider/logging"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type RequestLogParam struct {
	Logger logging.Logger
}

func NewRequestLog(p RequestLogParam) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:       true,
		LogProtocol:      true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURIPath:       true,
		LogRequestID:     true,
		LogReferer:       true,
		LogContentLength: true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogResponseSize:  true,
		LogHeaders: []string{
			"User-Agent", "Referer",
			"X-Forwarded-For", "X-Correlation-Id",
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestSize, _ := strconv.ParseInt(v.ContentLength, 10, 64)
			httpRequest := map[string]interface{}{
				"requestMethod": v.Method,
				"requestUrl":    v.URIPath,
				"requestSize":   requestSize,
				"responseSize":  v.ResponseSize,
				"status":        v.Status,
				"serverIp":      v.Host,
				"remoteAddr":    v.RemoteIP,
				"protocol":      v.Protocol,
				"receivedAt":    v.StartTime.UTC().Format(time.RFC3339),
				"duration":      v.Latency.Milliseconds(),
			}

			httpHeader := map[string]interface{}{
				"X-Request-Id": []string{v.RequestID},
			}
			for key, header := range v.Headers {
				httpHeader[key] = header
			}

			logger := p.Logger.WithFields(map[string]interface{}{
				"httpRequest": httpRequest,
				"httpHeader":  httpHeader,
			})

			message := fmt.Sprintf("request: %s %s", v.Method, v.URIPath)
			if v.Status >= 100 && v.Status <= 399 {
				logger.Info(message)
			} else if v.Status >= 400 && v.Status <= 499 {
				logger.Warn(message)
			} else {
				logger.Error(message)
			}
			return nil
		},
	})
}
