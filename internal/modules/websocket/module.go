package websocket

import (
	"context"

	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/module"
	ws "github.com/nekoimi/go-project-template/internal/websocket"
)

func init() {
	module.Register(NewModule(), module.ScopeHTTP)
}

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "websocket"
}

func (m *Module) Register(ctx *framework.ModuleContext) error {
	if !ctx.ModuleEnabled(m.Name()) || !ctx.Config.Websocket.Enabled {
		return nil
	}

	handler := ws.NewWSHandler(
		ctx.WSManager,
		ctx.Config.JWT.Secret,
		ctx.Logger,
		ctx.Config.Server.AllowedOrigins,
		ctx.Config.Websocket,
	)

	ctx.Router.GET("/ws/v1/chat", handler.Upgrade)
	return nil
}

func (m *Module) Boot(ctx context.Context, moduleCtx *framework.ModuleContext) error {
	if moduleCtx == nil || !moduleCtx.ModuleEnabled(m.Name()) || !moduleCtx.Config.Websocket.Enabled {
		return nil
	}
	go moduleCtx.WSManager.Run()
	return nil
}

func (m *Module) Shutdown(ctx context.Context, moduleCtx *framework.ModuleContext) error {
	if moduleCtx == nil || !moduleCtx.ModuleEnabled(m.Name()) || !moduleCtx.Config.Websocket.Enabled {
		return nil
	}
	moduleCtx.WSManager.Shutdown()
	return nil
}
