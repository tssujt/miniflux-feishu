//go:build wireinject
// +build wireinject

package main

import (
	"miniflux-feishu/internal/handlers"
	"miniflux-feishu/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	services.NewFeishuService,
	handlers.NewWebhookHandler,
	NewRouter,
)

func NewRouter(webhookHandler *handlers.WebhookHandler) *gin.Engine {
	return setupRouter(webhookHandler)
}

func InitializeApp() (*gin.Engine, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
