package fault

import (
	"context"
	"net/http"
	"time"

	"github.com/ltick/tick-routing"
)

func TimeoutHandler(timeoutDuration time.Duration) routing.Handler {
	return func(c *routing.Context) error {
		var (
			err     error
			errChan chan error      = make(chan error, 1)
			ctx     context.Context = context.Background()
		)
		if timeoutDuration > 0 {
			ctx, _ = context.WithTimeout(ctx, timeoutDuration)
		}
		go func(c *routing.Context) {
			errChan <- c.Next()
		}(c)
		select {
		case <-ctx.Done():
			c.Abort()
			err = routing.NewHTTPError(http.StatusRequestTimeout)
		case err = <-errChan:
		}
		// TODO 根据src/net/http/client.go细化超时取消流程
		if err != nil {
			writeError(c, err)
		}
		return nil
	}
}
