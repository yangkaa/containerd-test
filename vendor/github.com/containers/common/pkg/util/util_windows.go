// +build windows

package util

import (
	"github.com/pkg/errors"
)

// getRuntimeDir returns the runtime directory
func GetRuntimeDir() (string, error) {
	return "", errors.New("this function is not implemented for windows")
}
