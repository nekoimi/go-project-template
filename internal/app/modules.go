package app

import (
	"github.com/nekoimi/go-project-template/internal/framework"

	_ "github.com/nekoimi/go-project-template/internal/modules/auth"
	_ "github.com/nekoimi/go-project-template/internal/modules/examplejob"
	_ "github.com/nekoimi/go-project-template/internal/modules/upload"
	_ "github.com/nekoimi/go-project-template/internal/modules/user"
	_ "github.com/nekoimi/go-project-template/internal/modules/websocket"
)

func registeredModules(scopes ...framework.Scope) []framework.Module {
	return framework.Modules(scopes...)
}
