package tool

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

func RunExternalCommand(args []string, env []string, in io.Reader, out io.Writer, kill chan int, resize chan ResizeArg) (int, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env...)

	f, er := pty.Start(cmd)
	if er != nil {
		return 1, er
	}

	done := make(chan struct{})

	// Copy data from in to PTY
	go func() {
		for {
			_, err := io.Copy(f, in)
			if err != nil {
				return
			}
		}
	}()

	// Copy data from PTY to out
	go func() {
		for {
			_, err := io.Copy(out, f)
			if err != nil {
				return
			}
		}
	}()

	go func() {
		<-kill
		cmd.Process.Signal(os.Interrupt)
	}()

	go func() {
		for {
			select {
			case r := <-resize:
				pty.Setsize(f, &pty.Winsize{
					Rows: uint16(r.Height),
					Cols: uint16(r.Width),
				})
			case <-done:
				return
			}
		}
	}()

	cmd.Wait()
	// if done is listened, send a signal to exit
	select {
	case done <- struct{}{}:
	default:
	}
	return 0, nil
}

type SimpleToolRunner func(args map[string]any, in io.Reader, out io.Writer, err io.Writer, kill chan int) error

func CanioalizeWrapper(f SimpleToolRunner) ToolRunner {
	return func(args map[string]any, in io.Reader, out io.Writer, kill chan int, resize chan ResizeArg) (int, error) {
		ptmx, tty, err := pty.Open()

		if err != nil {
			return 1, err
		}

		defer func() {
			// FIXME: ensure ptmx flush
			<-time.After(2 * time.Second)
			ptmx.Close()
			tty.Close()
		}()

		// Copy data from PTY to out
		go func() {
			for {
				_, err := io.Copy(out, ptmx)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}()

		// Copy data from in to PTY
		go func() {
			for {
				_, err := io.Copy(ptmx, in)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}()

		err = f(args, tty, tty, tty, kill)

		// there is no need to cosume the resize channel
		if err != nil {
			return 1, err
		} else {
			return 0, nil
		}
	}
}
