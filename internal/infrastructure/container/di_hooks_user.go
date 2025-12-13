//go:build !user && !all
// +build !user,!all

package container

import (
	"github.com/samber/do/v2"
)

// UserRegistrationHook is called from NewContainer
func UserRegistrationHook(i do.Injector) {
	// Empty implementation - user handler not included in this build
}
