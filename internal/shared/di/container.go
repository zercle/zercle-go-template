package di

import (
	"github.com/samber/do/v2"
)

// Container is a dependency injection container.
type Container = do.Injector

// Root is the global application DI container.
var Root do.Injector

func init() {
	Root = do.New()
}
