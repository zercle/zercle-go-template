//go:build !health && !all
// +build !health,!all

package container

import (
	"github.com/samber/do/v2"
)

// HealthRegistrationHook is called from NewContainer
func HealthRegistrationHook(i do.Injector) {
	// Empty implementation - health handler not included in this build
}
