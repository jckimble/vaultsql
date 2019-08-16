// Copyright © 2019 James Kimble
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "gitlab.com/jckimble/vaultsql"

	"github.com/spf13/cobra"

	"io/ioutil"
	"path/filepath"
)

// rootCmd represents the base command when called without any command's
var rootCmd = &cobra.Command{
	Use:   "vaultmysql",
	Short: "A mysql/mysqldump cmd for vault",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		secretPath, _ := cmd.PersistentFlags().GetString("path")
		if secretPath != "" {
			os.Setenv("VAULT_SECRET_PATH", secretPath)
		}
		if os.Getenv("VAULT_SECRET_PATH") == "" {
			return fmt.Errorf("Vault Secret Path must be set")
		}
		token, _ := cmd.PersistentFlags().GetString("token")
		if token != "" {
			os.Setenv("VAULT_TOKEN", token)
		}
		if os.Getenv("VAULT_TOKEN") == "" {
			data, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".vault-token"))
			if err != nil {
				return fmt.Errorf("Error reading vault token: %s", err)
			}
			os.Setenv("VAULT_TOKEN", string(data))
		}
		return nil
	},
}

func wrapExec(f func(cmd *cobra.Command, args []string, db *sql.DB) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			return err
		}
		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}
		db, err := sql.Open("vault-mysql", fmt.Sprintf("{{username}}:{{password}}@tcp(%s:%d)/", host, port))
		if err != nil {
			return fmt.Errorf("Unable to Connect: %s", err)
		}
		return f(cmd, args, db)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("path", "", "Vault Secret Path")
	rootCmd.PersistentFlags().StringP("token", "t", "", "Vault Token")
	rootCmd.PersistentFlags().String("host", "127.0.0.1", "Mysql Host")
	rootCmd.PersistentFlags().Int("port", 3306, "Mysql Port")
}
