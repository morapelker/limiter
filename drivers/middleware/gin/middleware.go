package gin

import (
	"github.com/gin-gonic/gin"

	"github.com/ulule/limiter/v3"
)

// Middleware is the middleware for gin.
type Middleware struct {
	Limiter            *limiter.Limiter
	OnError            ErrorHandler
	OnLimitReached     LimitReachedHandler
	KeyGetter          KeyGetter
	ExcludedKey        func(string) bool
	SkipFailedRequests bool
}

// NewMiddleware return a new instance of a gin middleware.
func NewMiddleware(limiter *limiter.Limiter, options ...Option) gin.HandlerFunc {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        DefaultErrorHandler,
		OnLimitReached: DefaultLimitReachedHandler,
		KeyGetter:      DefaultKeyGetter,
		ExcludedKey:    nil,
	}

	for _, option := range options {
		option.apply(middleware)
	}

	return func(ctx *gin.Context) {
		middleware.Handle(ctx)
	}
}

// Handle gin request.
func (middleware *Middleware) Handle(c *gin.Context) {
	key := middleware.KeyGetter(c)
	if middleware.ExcludedKey != nil && middleware.ExcludedKey(key) {
		c.Next()
		return
	}

	context, err := middleware.Limiter.Get(c, key)
	if err != nil {
		middleware.OnError(c, err)
		c.Abort()
		return
	}

	if context.Reached {
		middleware.OnLimitReached(c)
		c.Set("reachedLimit", true)
		c.Abort()
		return
	}

	c.Next()
	if middleware.SkipFailedRequests {
		if !c.GetBool("reachedLimit") {
			_ = middleware.Limiter.Decr(c, key)
		}
	}
}
