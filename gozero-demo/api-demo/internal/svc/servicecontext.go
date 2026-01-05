package svc

import (
	"api-demo/internal/config"
	"api-demo/internal/middleware"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config config.Config
	Cache  rest.Middleware
	Auth   rest.Middleware
	CORS   rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Cache:  middleware.NewCacheMiddleware().Handle,
		Auth:   middleware.NewAuthMiddleware().Handle,
		CORS:   middleware.NewCORSMiddleware().Handle,
	}
}
