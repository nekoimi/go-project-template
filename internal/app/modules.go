package app

import (
	"github.com/nekoimi/go-project-template/internal/framework"
	"github.com/nekoimi/go-project-template/internal/modules/auth"
	"github.com/nekoimi/go-project-template/internal/modules/examplejob"
	"github.com/nekoimi/go-project-template/internal/modules/upload"
	"github.com/nekoimi/go-project-template/internal/modules/user"
	modulews "github.com/nekoimi/go-project-template/internal/modules/websocket"
)

func builtinModules() []framework.Module {
	return []framework.Module{
		auth.NewModule(),
		user.NewModule(),
		upload.NewModule(),
		modulews.NewModule(),
		examplejob.NewModule(),
	}
}

func builtinSchedulerModules() []framework.Module {
	return []framework.Module{
		examplejob.NewModule(),
	}
}
