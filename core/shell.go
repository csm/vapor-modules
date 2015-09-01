package main

import (
    "github.com/csm/go-edn/types"
    "github.com/csm/vapor-modules"
    "bytes"
    "errors"
    "fmt"
    "os/exec"
    "strings"
)

type Shell struct{}

func (this Shell) Doc() string {
    return "Run a command with a shell. Input is a map that at least contains a :command argument, a string giving the command to run. Optional is the :shell argument, which can be the path of the shell to run; default is /bin/sh. :input may be set to a string, which will be sent to the shell's standard input. :env may be a map of environment variable names to values, which will be passed as the environment to the command."
}

func (this Shell) TakesInput() bool {
    return true
}

func (this Shell) Exec(input types.Value) (output types.Value, err error) {
    var commandMap types.Map
    if m, ok := input.(types.Map); ok {
        commandMap = m
    } else {
        err = errors.New("input must be a map")
        return
    }

    var command string
    var commandEntry = commandMap[types.Keyword("command")]
    if commandEntry == nil {
        err = errors.New(":command argument is required")
        return
    } else if c, ok := commandEntry.(types.String); ok {
        command = string(c)
    } else {
        err = errors.New(":command must be a string")
        return
    }

    var shell = "/bin/sh"
    shellEntry := commandMap[types.Keyword("shell")]
    if shellEntry != nil {
        if s, ok := shellEntry.(types.String); ok {
            shell = string(s)
        } else {
            err = errors.New(":shell must be a string if specified")
            return
        }
    }

    cmd := exec.Command(shell, "-c", command)

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

    inputEntry := commandMap[types.Keyword("input")]
    if inputEntry != nil {
        if i, ok := inputEntry.(types.String); ok {
            cmd.Stdin = strings.NewReader(string(i))
        } else {
            err = errors.New(":input must be a string if specified")
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
            types.Keyword("err"): types.String(stderr.String()),
            types.Keyword("out"): types.String(stdout.String()),
            types.Keyword("system-time"): types.Float(cmd.ProcessState.SystemTime().Seconds()),
            types.Keyword("user-time"): types.Float(cmd.ProcessState.UserTime().Seconds()),
        }
        return
    }

    output = types.Map{
        types.Keyword("success"): types.Bool(true),
        types.Keyword("err"): types.String(stderr.String()),
        types.Keyword("out"): types.String(stdout.String()),
        types.Keyword("system-time"): types.Float(cmd.ProcessState.SystemTime().Seconds()),
        types.Keyword("user-time"): types.Float(cmd.ProcessState.UserTime().Seconds()),
    }
    return
}

func main() {
    vapor.RunModule(Shell{})
}
