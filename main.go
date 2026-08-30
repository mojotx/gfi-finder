// Command gfi-finder finds "good first issue" candidates in a GitHub repository.
package main

import (
	"fmt"
	"os"

	"github.com/mojotx/gfi-finder/internal/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
