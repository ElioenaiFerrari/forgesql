/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var config Config

func init() {
	file, err := os.ReadFile(".forgesql.yml")
	if os.IsNotExist(err) {
		return
	}

	if err := yaml.Unmarshal(file, &config); err != nil {
		panic(err)
	}
}

func getDBDriverAndURL(connection string) (string, string) {
	sqliteExp := regexp.MustCompile("sqlite://")
	mysqlRegex := regexp.MustCompile("mysql://")
	postgresRegex := regexp.MustCompile("postgres://")

	switch {
	case sqliteExp.MatchString(connection):
		return "sqlite3", connection[len("sqlite://"):]
	case mysqlRegex.MatchString(connection):
		return "mysql", connection[len("mysql://"):]
	case postgresRegex.MatchString(connection):
		return "postgres", connection[len("postgres://"):]
	default:
		panic("Invalid database URL")
	}
}

func upOrDown(ctx context.Context, flags *pflag.FlagSet, operation string) {
	env, _ := flags.GetString("environment")

	currentEnv := config.GetEnv(env)

	driver, url := getDBDriverAndURL(currentEnv.DBURL)

	db, err := sql.Open(driver, url)
	if err != nil {
		panic(err)
	}

	migrationsPath := fmt.Sprintf("%s/%s", config.MigrationsPath, currentEnv.Label)

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		fmt.Println("No migrations found")
		return
	}

	file, err := os.Open(migrationsPath)
	if err != nil {
		panic(err)
	}

	files, err := file.Readdir(-1)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		isValid, _ := regexp.MatchString(fmt.Sprintf(`\.(%s)\.sql$`, operation), f.Name())
		if !isValid {
			continue
		}

		pathname := fmt.Sprintf("%s/%s", migrationsPath, f.Name())
		content, err := os.ReadFile(pathname)
		if err != nil {
			panic(err)
		}

		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}

		_, err = tx.ExecContext(ctx, string(content))
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", pathname)
		tx.Commit()
	}
}

func generateCmd(ctx context.Context, flags *pflag.FlagSet) {
	name, err := flags.GetString("name")
	if err != nil {
		panic(err)
	}

	env, err := flags.GetString("environment")
	if err != nil {
		panic(err)
	}

	currentEnv := config.GetEnv(env)

	migrationsPath := fmt.Sprintf("%s/%s", config.MigrationsPath, currentEnv.Label)
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		os.MkdirAll(migrationsPath, 0755)
	}

	timestamp := time.Now().Unix()

	upFilename := fmt.Sprintf("%s/%d_%s.up.sql", migrationsPath, timestamp, name)
	downFilename := fmt.Sprintf("%s/%d_%s.down.sql", migrationsPath, timestamp, name)

	upFile, err := os.Create(upFilename)
	if err != nil {
		panic(err)
	}
	defer upFile.Close()

	downFile, err := os.Create(downFilename)
	if err != nil {
		panic(err)
	}
	defer downFile.Close()
}

// migrationCmd represents the migration command
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Migrations",
	Long: `
	Migrations are a convenient way to alter your database schema over time in a
	way that can easily be replicated and shared with others.

	ForgeSQL uses a simple convention to generate migrations files. Each migration
	file name contains a timestamp which allows ForgeSQL to determine the order of
	the migrations.`,
	Example: `
	forgesql migration --generate --name create_example
	forgesql migration --up --environment dev
	forgesql migration --down --environment dev`,
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		ctx := cmd.Context()
		generate, _ := flags.GetBool("generate")

		if generate {
			generateCmd(ctx, flags)
			return
		}

		up, _ := flags.GetBool("up")
		if up {
			upOrDown(ctx, flags, "up")
			return
		}

		down, _ := flags.GetBool("down")
		if down {
			upOrDown(ctx, flags, "down")
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(migrationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	migrationCmd.Flags().StringP("environment", "e", "dev", "migration env")
	migrationCmd.Flags().BoolP("generate", "g", false, "generate migration")
	migrationCmd.Flags().BoolP("up", "u", false, "up migrations")
	migrationCmd.Flags().BoolP("down", "d", false, "down migrations")
	migrationCmd.Flags().StringP("name", "n", "create_example", "migration name")
}
