package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("wat should I do")
	}
}

func parent() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

}

func child() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("pwd", err)
		return
	}
	path := pwd + "/ubuntu-base-16.04.6-base-arm64"

	syscall.Sethostname([]byte("container"))
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	if err := syscall.Mount(path, path, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		fmt.Println("Mount", err)
		return
	}

	if err := os.MkdirAll(path+"/.old", 0700); err != nil {
		fmt.Println("mkdir", err)
		return
	}

	err = syscall.PivotRoot(path, path+"/.old")
	if err != nil {
		fmt.Println("pivot root ", err)
		return
	}

	syscall.Chdir("/")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}

	syscall.Unmount("/proc", 0)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
