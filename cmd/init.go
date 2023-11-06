/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Env struct {
	Label string `yaml:"label"`
	DBURL string `yaml:"db_url"`
}

type Config struct {
	MigrationsPath string `yaml:"migrations_path"`
	Envs           []Env  `yaml:"env"`
}

func (c *Config) GetEnv(label string) Env {
	for _, env := range c.Envs {
		if env.Label == label {
			return env
		}
	}

	panic("Invalid environment")
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize forgeSQL",
	Long:    "Initialize forgeSQL creating a .forgesql.yml file in the current directory.",
	Example: "forgesql init",
	Run: func(cmd *cobra.Command, args []string) {
		var migrationsPath string
		var envsSeparetedByComma string
		var envs []Env

		fmt.Println("Migrations path (default: migrations): ")
		fmt.Scanln(&migrationsPath)

		dir, _ := os.Getwd()

		if len(os.Args) > 2 {
			dir = os.Args[2]
			fmt.Println(dir)
		}

		if migrationsPath == "" {
			migrationsPath = fmt.Sprintf("%s/migrations", dir)
		}

		fmt.Println("Environments separated by comma (default: dev): ")
		fmt.Scanln(&envsSeparetedByComma)

		if len(envsSeparetedByComma) == 0 {
			envs = []Env{
				{
					Label: "dev",
					DBURL: "sqlite://db.sqlite",
				},
			}
		} else {
			envsSplited := strings.Split(envsSeparetedByComma, ",")
			for _, env := range envsSplited {
				var dbURL string

				fmt.Printf("DB URL for %s (default: sqlite://db.sqlite): ", env)
				fmt.Scanln(&dbURL)

				if dbURL == "" {
					dbURL = "sqlite://db.sqlite"
				}

				envs = append(envs, Env{
					Label: env,
					DBURL: dbURL,
				})
			}
		}

		config := Config{
			MigrationsPath: migrationsPath,
			Envs:           envs,
		}

		// Write yaml file
		b, err := yaml.Marshal(config)
		if err != nil {
			panic(err)
		}

		// Create yaml file based in migrations_path
		pathname := fmt.Sprintf("%s/.forgesql.yml", dir)
		if err := os.WriteFile(pathname, b, 0755); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
