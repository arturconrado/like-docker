//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	if len(os.Args) < 4 {
		usageAndExit()
	}
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		usageAndExit()
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func child() {
	rootfs := filepath.Clean(os.Args[2])
	must(syscall.Sethostname([]byte("minidock-100")))
	must(syscall.Mount("", "/", "", uintptr(syscall.MS_REC|syscall.MS_PRIVATE), ""))
	must(syscall.Mount(rootfs, rootfs, "", uintptr(syscall.MS_BIND|syscall.MS_REC), ""))
	must(os.MkdirAll(filepath.Join(rootfs, "oldrootfs"), 0o700))
	must(syscall.PivotRoot(rootfs, filepath.Join(rootfs, "oldrootfs")))
	must(os.Chdir("/"))
	must(syscall.Unmount("/oldrootfs", syscall.MNT_DETACH))
	_ = os.RemoveAll("/oldrootfs")

	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func usageAndExit() {
	fmt.Println("usage: container100 run <rootfs> <command> [args...]")
	os.Exit(1)
}
