package docker

import (
	"net"
	"os"
	"os/exec"
	"syscall"
)

func run(addr net.Addr, args []string) int {
	pargs := []string{
		"-H",
		addr.String(),
	}
	for _, arg := range args {
		pargs = append(pargs, arg)
	}

	cmd := exec.Command("docker", pargs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			if ws, ok := err.ProcessState.Sys().(syscall.WaitStatus); ok {
				return ws.ExitStatus()
			}
		}
		return 1
	}

	return 0
}

// Run ...
func Run(addr net.Addr, args []string) {
	os.Exit(run(addr, args))
}
