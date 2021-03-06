package cmd

import (
	"fmt"

	"github.com/akerl/prospectus/checks"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for state changes",
	RunE:  checkRunner,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	f := checkCmd.Flags()
	f.BoolP("all", "a", false, "Show all items, regardless of state")
	f.Bool("json", false, "Print output as JSON")
}

func checkRunner(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	flagAll, err := flags.GetBool("all")
	if err != nil {
		return err
	}
	flagJSON, err := flags.GetBool("json")
	if err != nil {
		return err
	}

	params := []string{"."}
	if len(args) != 0 {
		params = args
	}

	c, err := checks.NewSet(params)
	if err != nil {
		return err
	}
	results := c.Execute()
	if err != nil {
		return err
	}
	if !flagAll {
		results = results.Changed()
	}

	var output string
	if flagJSON {
		output, err = results.JSON()
		if err != nil {
			return err
		}
	} else {
		output = results.String()
	}
	fmt.Println(output)
	return nil
}
