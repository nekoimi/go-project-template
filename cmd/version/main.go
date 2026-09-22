package main

import (
	"fmt"

	"github.com/nekoimi/go-project-template/internal/buildinfo"
)

func main() {
	fmt.Printf("version=%s commit=%s build_time=%s\n", buildinfo.Version, buildinfo.Commit, buildinfo.BuildTime)
}
