package di

import (
	"github.com/samber/do"
)

type Container = do.Injector

var Root *Container

func init() {
	Root = do.New()
}
