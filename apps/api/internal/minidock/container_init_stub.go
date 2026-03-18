//go:build !linux

package minidock

func HandleContainerInitCommand() bool {
	return false
}
