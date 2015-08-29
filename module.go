package vapor

import (
	"github.com/csm/go-edn"
	"github.com/csm/go-edn/types"
	"os"
	"bufio"
	"fmt"
)

type Module interface {
	TakesInput() bool
	Exec(types.Value) (types.Value, error)
}

func RunModule(module Module) {
	var input types.Value = nil
	var err error = nil
	if module.TakesInput() {
		input, err = edn.ParseReader(bufio.NewReader(os.Stdin))
	}
	if err != nil {
		os.Exit(1)
	}
	var output types.Value = nil
	output, err = module.Exec(input)
	if err != nil {
		if output != nil {
			fmt.Print(edn.DumpString(output))
		}
		os.Exit(1)
	}
	if output != nil {
		fmt.Println(edn.DumpString(output))
	}
	os.Exit(0)
}
