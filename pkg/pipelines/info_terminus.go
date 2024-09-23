package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/terminus"
)

func PrintTerminusInfo() {
	var cli = &terminus.GetTerminusVersion{}
	terminusVersion, err := cli.Execute()
	if err != nil {
		fmt.Printf("Terminus might not be installed.\n")
		return
	}

	fmt.Printf("%s\n", terminusVersion)
}
