package vapor

import (
    "github.com/csm/go-edn"
    "github.com/csm/go-edn/types"
    "os"
    "bufio"
    "fmt"
)

type Module interface {
    Doc() string
    TakesInput() bool
    Exec(types.Value) (types.Value, error)
}

func RunModule(module Module) {
    var output types.Value = nil
    defer func() {
        var result = 0
        if r := recover(); r != nil {
            result = 1
            if output == nil {
                output = types.Map{
                    types.Keyword("success"): types.Bool(false),
                    types.Keyword("error"): types.String(fmt.Sprint(r)),
                }
            }
        }
        if output == nil {
            output = types.Map{
                types.Keyword("success"): types.Bool(true),
            }
        }
        fmt.Println(edn.DumpString(output))
        os.Exit(result)
    }()
    var input types.Value = nil
    var err error = nil
    if module.TakesInput() {
        input, err = edn.ParseReader(bufio.NewReader(os.Stdin))
    }
    if err != nil {
        panic(err)
    }
    output, err = module.Exec(input)
    if err != nil {
        panic(err)
    }
}
