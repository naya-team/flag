package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/naya-team/flag"
	"github.com/naya-team/flag/connection"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "flagger",
		Short: "Flagger for service",
		Long:  `This cli for now only contains eraser command.`,
	}

	eraserCmd = &cobra.Command{
		Use:   "eraser",
		Short: "Eraser for flagger",
		Run: func(cmd *cobra.Command, args []string) {

			// get argument from flag
			// get -dsn from flag
			dsn := cmd.PersistentFlags().Lookup("dsn").Value.String()

			driver := cmd.PersistentFlags().Lookup("db-driver").Value.String()

			// get -flags from flag
			var _flags []string
			flags := cmd.PersistentFlags().Lookup("flags").Value.String()
			if flags != "" {
				_flags = strings.Split(flags, ",")
			}

			if len(_flags) == 0 && dsn == "" {
				log.Fatal("cannot get dsn from flag, dsn required if args flags not set, use --dsn for set dsn")
			}
			// get -root-path from flag
			rootPath := cmd.PersistentFlags().Lookup("root-path").Value.String()
			if rootPath == "" {
				rootPath = "."
			}

			var db *sql.DB
			switch driver {
			case "mysql":
				db = connection.NewMysqlConnection(dsn)
			case "postgres":
				db = connection.NewPostgresConnection(dsn)
			}

			flag.Eraser(_flags, rootPath, db)
		},
	}
)

func init() {
	eraserCmd.PersistentFlags().String("db-driver", "", "mysql or postgres")
	eraserCmd.PersistentFlags().String("dsn", "", "dsn for database connection")
	eraserCmd.PersistentFlags().String("flags", "", "flags that will be deleted, separated by comma")
	eraserCmd.PersistentFlags().String("root-path", "", "root path for search flagger")
	rootCmd.AddCommand(eraserCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
