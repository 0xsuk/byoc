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
		run()
	case "child":
		child()
	default:
		panic("bad command")

	}
}

func run() {
	fmt.Printf("Creating namespace %v as %v\n", os.Args[2:], os.Getpid())

	//Create namespace by running go run main.go run child ...args
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS, //TODO ? //issue new pid
		Unshareflags: syscall.CLONE_NEWNS,                                               //so that "mount | grep proc" => "proc on /home/container/proc type proc (rw,relatime)" is not visible
	}

	err := cmd.Run()
	must(err)
}

func child() {
	fmt.Printf("Running %v as %v\n", os.Args[2:], os.Getpid())

	syscall.Sethostname([]byte("container"))
	//isolate process in container from parent process
	syscall.Chroot("/home/container") //! DO cp /bin/bash /home/container/bin/bash  && YOU HAVE TO DO https://unix.stackexchange.com/a/416556
	syscall.Chdir("/")                //because of above line, cwd in host process does not exist in child process, ending up directory undefined
	//     mount ./proc to /proc, type is proc, default flag 0, no data
	syscall.Mount("proc", "/proc", "proc", 0, "") //! make sure ./proc exists
	defer syscall.Unmount("/proc", 0)

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	must(err)

	fmt.Println("Done")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
