package fault

import (
	"context"
	"net/http"
	"time"

	"github.com/ltick/tick-routing"
)

func TimeoutHandler(timeoutDuration time.Duration) routing.Handler {
	return func(c *routing.Context) (err error) {
		var (
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
			writeError(c, routing.NewHTTPError(http.StatusRequestTimeout))
			return nil
		case err = <-errChan:
			return
		}
	}
}
