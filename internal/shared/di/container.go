package di

import (
	"github.com/samber/do"
)

// Container is an alias to do.Injector, the dependency injection container.
type Container = do.Injector

// Root is the global dependency injection container used throughout the application.
var Root *Container

func init() {
	Root = do.New()
}
