package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

type Runner struct {
	Dir string
}

func (r Runner) RunSequential(ctx context.Context, commands []string) error {
	for _, command := range commands {
		if err := r.RunOne(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) RunParallel(ctx context.Context, commands []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, len(commands))
	var wg sync.WaitGroup
	for _, command := range commands {
		command := command
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.RunOne(ctx, command); err != nil {
				cancel()
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) RunOne(ctx context.Context, command string) error {
	fmt.Printf("→ %s\n", command)
	name, args := shell(command)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = r.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s", command)
	}
	return nil
}

func shell(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}
