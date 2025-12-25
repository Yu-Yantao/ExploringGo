package svc

import (
	"api/internal/config"
	"api/internal/middleware"

	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config config.Config
	demo   rest.Middleware
	demo2  rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		demo:   middleware.NewDemoMiddleware().Handle,
		demo2:  middleware.NewDemo2Middleware().Handle,
	}
}
