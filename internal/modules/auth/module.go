package auth

import (
	"context"
	"time"

	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/module"
	"github.com/nekoimi/go-project-template/internal/pkg/resp"
	"github.com/nekoimi/go-project-template/internal/repository"
	"go.uber.org/zap"
)

func init() {
	module.Register(NewModule(), module.ScopeHTTP)
}

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "auth"
}

func (m *Module) Register(ctx *framework.ModuleContext) error {
	if !ctx.ModuleEnabled(m.Name()) {
		return nil
	}

	userRepo := repository.NewUserRepository(ctx.DB)
	jwtExpire := time.Duration(ctx.Config.JWT.ExpireHours) * time.Hour
	authService := NewService(userRepo, ctx.DB, ctx.Config.JWT.Secret, jwtExpire, ctx.Events)
	authHandler := NewHandler(authService, ctx.Logger)

	ctx.Subscribe(EventUserRegistered, func(eventCtx context.Context, event framework.Event) error {
		payload, ok := event.Payload.(UserRegisteredEvent)
		if !ok {
			ctx.Logger.Warn("invalid user registered event payload")
			return nil
		}
		ctx.Logger.Info("user registered",
			zap.String("user_id", payload.UserID),
			zap.String("username", payload.Username),
			zap.String("email", payload.Email),
		)
		return nil
	})

	auth := ctx.API.Group("/auth")
	auth.POST("/register", resp.Handle(authHandler.Register, ctx.Logger))
	auth.POST("/login", resp.Handle(authHandler.Login, ctx.Logger))

	return nil
}
