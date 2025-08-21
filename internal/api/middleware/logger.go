package middleware

import (
	"time"

	"cobra-template/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		param := gin.LogFormatterParams{
			StatusCode: c.Writer.Status(),
			Latency:    time.Since(start),
			ClientIP:   c.ClientIP(),
			Method:     c.Request.Method,
			Path:       path,
		}

		if raw != "" {
			param.Path = path + "?" + raw
		}

		logFields := logrus.Fields{
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
		}

		if len(c.Errors) > 0 {

			logFields["error"] = c.Errors.String()
			logger.WithFields(logFields).Error("Request completed with errors")
		} else {

			if param.StatusCode >= 500 {
				logger.WithFields(logFields).Error("Request completed with server error")
			} else if param.StatusCode >= 400 {
				logger.WithFields(logFields).Warn("Request completed with client error")
			} else {
				logger.WithFields(logFields).Info("Request completed")
			}
		}
	}
}
