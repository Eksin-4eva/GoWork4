package svc

import (
	"gobili/gateway/internal/app"
	"gobili/gateway/internal/config"
)

type ServiceContext struct {
	Config config.Config
	App    *app.App
}

func NewServiceContext(c config.Config) *ServiceContext {
	application, err := app.NewApp(c)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		App:    application,
	}
}
