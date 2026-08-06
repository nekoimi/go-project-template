package user

import (
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/pkg/resp"
	"github.com/nekoimi/go-project-template/internal/repository"
)

func init() {
	framework.Register(NewModule(), framework.ScopeHTTP)
}

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "user"
}

func (m *Module) Register(ctx *framework.ModuleContext) error {
	if !ctx.ModuleEnabled(m.Name()) {
		return nil
	}

	userRepo := repository.NewUserRepository(ctx.DB)
	userService := NewService(userRepo)
	userHandler := NewHandler(userService, ctx.Logger)

	users := ctx.Protected.Group("/users")
	users.GET("/profile", resp.Handle(userHandler.GetProfile, ctx.Logger))

	return nil
}
