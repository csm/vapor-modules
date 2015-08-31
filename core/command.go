package main

import (
	"container/list"
	"github.com/csm/go-edn/types"
	"github.com/csm/vapor-modules"
	"os/exec"
	"errors"
	"fmt"
	"bytes"
	"strings"
)

type Command struct{}

func (this Command) TakesInput() bool {
	return true
}

// The command module takes as input a map, and runs that command.
//
// Parameters in the input map are:
//   :command, a list or vector, giving the command to run. Required.
//   :env, a map of names to values; if specified, these are added to the environment when running the command; optional
//   :input, a string to pass to the command's stdin; optional
func (this Command) Exec(input types.Value) (output types.Value, err error) {
	var command []string
	var commandMap types.Map
	if m, ok := input.(types.Map); ok {
		commandMap = m
	} else {
		err = errors.New("input must be a map")
		return
	}

	var commandEntry = commandMap[types.Keyword("command")]
	if commandEntry == nil {
		err = errors.New("require at least :command in input")
		return
	} else if l, ok := commandEntry.(*types.List); ok {
		var list = (*list.List)(l)
		command = make([]string, list.Len())
		var i = 0
		for e := list.Front(); e != nil; e = e.Next() {
			if elem, ok := e.Value.(types.String); ok {
				command[i] = string(elem)
				i++
			} else {
				err = errors.New(":command must be a list of strings")
				return
			}
		}
	} else if v, ok := commandEntry.(types.Vector); ok {
		var vect = ([]types.Value)(v)
		command = make([]string, len(vect))
		for i, e := range vect {
			if elem, ok := e.(types.String); ok {
				command[i] = string(elem)
			} else {
				err = errors.New(":command must be a vector of strings")
				return
			}
		}
	} else {
		err = errors.New("expected a list or vector as :command")
		return
	}
	var cmd = exec.Command(command[0], command[1:]...)

	var envEntry = commandMap[types.Keyword("env")]
	if envEntry != nil {
		if e, ok := envEntry.(types.Map); ok {
			var env = make([]string, len(e))
			var i = 0
			for k := range e {
				v := e[k]
				if kk, ok := k.(types.String); ok && len(kk) > 0 {
					if vv, ok := v.(types.String); ok && len(vv) > 0 {
						env[i] = fmt.Sprint(string(kk), "=", string(vv))
					} else {
						err = errors.New(":env values must be nonempty strings")
						return
					}
				} else {
					err = errors.New(":env keys must be nonempty strings")
					return
				}
				i++
			}
			cmd.Env = env
		} else {
			err = errors.New(":env must be a map if specified")
			return
		}
	}

	var stdinEntry = commandMap[types.Keyword("input")]
	if stdinEntry != nil {
		if stdin, ok := stdinEntry.(types.String); ok {
			cmd.Stdin = strings.NewReader(string(stdin))
		} else {
			err = errors.New(":input must be a string if specified")
			return
		}
	}

	var stderr = new(bytes.Buffer)
	var stdout = new(bytes.Buffer)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err = cmd.Run()
	if err != nil {
		output = types.Map {
			types.Keyword("success"): types.Bool(false),
			types.Keyword("stderr"): types.String(stderr.String()),
			types.Keyword("stdout"): types.String(stdout.String()),
			types.Keyword("system-time"): types.Float(cmd.ProcessState.SystemTime().Seconds()),
			types.Keyword("user-time"): types.Float(cmd.ProcessState.UserTime().Seconds()),
		}
		return
	}
	cmd.ProcessState.SysUsage()

	output = types.Map{
		types.Keyword("success"): types.Bool(true),
		types.Keyword("stderr"): types.String(stderr.String()),
		types.Keyword("stdout"): types.String(stdout.String()),
		types.Keyword("system-time"): types.Float(cmd.ProcessState.SystemTime().Seconds()),
		types.Keyword("user-time"): types.Float(cmd.ProcessState.UserTime().Seconds()),
	}
	return
}

func main() {
	vapor.RunModule(Command{})
}