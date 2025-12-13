//go:build !post && !all
// +build !post,!all

package container

import (
	"github.com/samber/do/v2"
)

// PostRegistrationHook is called from NewContainer
func PostRegistrationHook(i do.Injector) {
	// Empty implementation - post handler not included in this build
}
