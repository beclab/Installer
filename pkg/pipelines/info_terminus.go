package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/terminus"
)

func PrintTerminusInfo() {
	var cli = &terminus.GetTerminusVersion{}
	terminusVersion, err := cli.Execute()
	if err != nil {
		fmt.Printf("Terminus: not installed\n")
		return
	}

	fmt.Printf("Terminus: %s\n", terminusVersion)
}
