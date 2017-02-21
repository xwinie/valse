package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kildevaeld/valse"
)

func NewWithNameAndLogrus(name string, l logrus.FieldLogger) valse.MiddlewareHandler {
	return func(next valse.RequestHandler) valse.RequestHandler {
		return func(c *valse.Context) error {
			start := time.Now()

			entry := l.WithFields(logrus.Fields{
				"request": string(c.URI().FullURI()),
				"method":  string(c.Method()),
				"remote":  c.RemoteAddr(),
			})

			if reqID := c.Request.Header.Peek("X-Request-Id"); string(reqID) != "" {
				entry = entry.WithField("request_id", string(reqID))
			}

			entry.Info("started handling request")

			if err := next(c); err != nil {
				return err
			}

			latency := time.Since(start)

			entry.WithFields(logrus.Fields{
				"status":      c.Response.StatusCode(),
				"text_status": http.StatusText(c.Response.StatusCode()),
				"took":        latency,
				fmt.Sprintf("measure#%s.latency", name): latency.Nanoseconds(),
			}).Info("completed handling request")

			return nil
		}
	}
}
