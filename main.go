package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const exec_sep = "--"

func main() {
	var cmds_to_run []string

	var i int
	var arg string
	var env_name string
	env_m := map[string]string{}
	for i, arg = range os.Args {
		switch env_name {
		case "":
			if !strings.HasPrefix(arg, "-") {
				continue
			}
			env_str := strings.TrimLeft(arg, "-")
			env_list := strings.SplitN(env_str, "=", 2)
			switch len(env_list) {
			case 1:
				env_name = env_list[0]
			case 2:
				env_m[env_list[0]] = env_list[1]
			default:
				log.Fatalf("unreachable!")
			}
		default:
			env_m[env_name] = arg
			env_name = ""
		}

		if strings.TrimSpace(arg) == exec_sep {
			break
		}
	}
	cmds_to_run = os.Args[i+1:]

	if len(cmds_to_run) == 0 {
		log.Fatal("no commnad to run, uses '--' to seperate command to run")
	}

	for k, v := range env_m {
		os.Setenv(k, v)
	}

	cmd := exec.Command(cmds_to_run[0], cmds_to_run[1:]...)
	err := func() error {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}

		tcnt := sync.WaitGroup{}
		tcnt.Add(2)
		go func() {
			defer func() {
				tcnt.Done()
			}()
			b := make([]byte, 1)
			for {
				n, err := stderr.Read(b)
				if err != nil {
					return
				}
				os.Stderr.Write(b[:n])
			}
		}()

		go func() {
			defer func() {
				tcnt.Done()
			}()
			b := make([]byte, 1)
			for {
				n, err := stdout.Read(b)
				if err != nil {
					return
				}
				os.Stdout.Write(b[:n])
			}
		}()

		go func() {
			defer func() {
				stdin.Close()
			}()
			b := make([]byte, 1)
			for {
				n, err := os.Stdin.Read(b)
				if err != nil {
					return
				}
				stdin.Write(b[:n])
			}
		}()

		tcnt.Wait()

		return nil
	}()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
