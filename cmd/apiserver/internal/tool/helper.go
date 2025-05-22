package tool

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

var ExternalCommandToolSignals = []ToolSignal{
	*ToolSignalTemplateInt,
	*ToolSignalTemplateTerm,
	*ToolSignalTemplateKill,
}

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
		for {
			select {
			case <-done:
				return
			case sig := <-kill:
				switch sig {
				case 2:
					cmd.Process.Signal(os.Interrupt)
				case 15:
					cmd.Process.Signal(os.Kill)
				case 9:
					cmd.Process.Signal(os.Kill)
				}
			}
		}
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
	// close done, so that all listeners can receive empty
	// messages and exit
	close(done)
	return 0, nil
}

type SimpleToolRunner func(args map[string]any, in io.Reader, out io.Writer, err io.Writer, kill chan int) error

type RedWriter struct {
	w io.Writer
}

func (r *RedWriter) Write(p []byte) (n int, err error) {
	b := []byte("\033[31m")
	b = append(b, p...)
	b = append(b, []byte("\033[0m")...)
	rn, err := r.w.Write(b)
	inN := rn - len("\033[31m")
	if inN < 0 {
		n = 0
	} else if inN >= len(p) {
		n = len(p)
	} else {
		n = inN
	}
	return
}

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

		err = f(args, tty, tty, &RedWriter{
			w: tty,
		}, kill)

		// there is no need to cosume the resize channel
		if err != nil {
			return 1, err
		} else {
			return 0, nil
		}
	}
}
