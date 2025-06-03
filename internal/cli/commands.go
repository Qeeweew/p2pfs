package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// RootCmd is the base command for the p2pfs CLI.
var RootCmd = &cobra.Command{
	Use:   "p2pfs",
	Short: "p2pfs CLI",
}

func init() {
	RootCmd.AddCommand(addCmd, getCmd, pinCmd, catCmd, lsCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a file to the P2P file system",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add command not implemented")
	},
}

var getCmd = &cobra.Command{
	Use:   "get [cid] [output]",
	Short: "Retrieve a file by CID",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get command not implemented")
	},
}

var pinCmd = &cobra.Command{
	Use:   "pin [cid]",
	Short: "Pin a block locally",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pin command not implemented")
	},
}

var catCmd = &cobra.Command{
	Use:   "cat [cid]",
	Short: "Print block raw data",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cat command not implemented")
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls [cid]",
	Short: "List links in a DAG node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ls command not implemented")
	},
}
