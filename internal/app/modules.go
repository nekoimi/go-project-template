package app

import (
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/module"

	_ "github.com/nekoimi/go-project-template/internal/modules/auth"
	_ "github.com/nekoimi/go-project-template/internal/modules/examplejob"
	_ "github.com/nekoimi/go-project-template/internal/modules/upload"
	_ "github.com/nekoimi/go-project-template/internal/modules/user"
	_ "github.com/nekoimi/go-project-template/internal/modules/websocket"
)

func registeredModules() []framework.Module {
	return module.Modules(module.ScopeHTTP, module.ScopeScheduler)
}

func registeredSchedulerModules() []framework.Module {
	return module.Modules(module.ScopeScheduler)
}
