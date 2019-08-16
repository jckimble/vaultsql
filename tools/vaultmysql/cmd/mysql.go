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

	"bufio"
	"bytes"
	"database/sql"
	"github.com/abiosoft/readline"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// mysqlCmd represents the mysql command
var mysqlCmd = &cobra.Command{
	Use:   "mysql",
	Short: "Uses vault token to login to mysql",
	RunE:  wrapExec(mysql),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("Can only use one database")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mysqlCmd)
}

func isPipe() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return false
	}
	return true
}

type StringReader interface {
	ReadString(byte) (string, error)
}

func isExec(line string) bool {
	chk := strings.ToLower(line)
	execs := []string{"create", "update", "insert", "delete", "drop", "use"}
	for _, exec := range execs {
		if strings.HasPrefix(chk, exec) {
			return true
		}
	}
	return false
}

func useDatabase(db *sql.DB, name string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(fmt.Sprintf("USE `%s`;", name)); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func runQueries(db *sql.DB, reader StringReader) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for {
		line, err := reader.ReadString([]byte(";")[0])
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("Parsing Error: %s", err)
			}
			break
		}
		line = strings.TrimSpace(line)
		if isExec(line) {
			_, err := tx.Exec(line)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			rows, err := tx.Query(line)
			if err != nil {
				tx.Rollback()
				return err
			}
			columns, err := rows.Columns()
			if err != nil {
				rows.Close()
				return err
			}
			values := make([]interface{}, len(columns))
			valueptr := make([]interface{}, len(columns))
			out := make([]string, len(columns))
			fmt.Println(strings.Join(columns, "\t| "))
			for rows.Next() {
				for i := range values {
					valueptr[i] = &values[i]
				}
				rows.Scan(valueptr...)
				for i, val := range values {
					if str, ok := val.(string); ok {
						out[i] = str
					} else {
						out[i] = toString(val)
					}
				}
				fmt.Println(strings.Join(out, "\t| "))
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return err
			}
			rows.Close()
		}
	}
	return tx.Commit()
}

func toString(v interface{}) string {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return fmt.Sprintf("%+v", v)
}

func mysql(cmd *cobra.Command, args []string, db *sql.DB) error {
	if len(args) == 1 {
		if err := useDatabase(db, args[0]); err != nil {
			return err
		}
	}
	if !isPipe() {
		rl, err := readline.New("> ")
		if err != nil {
			return fmt.Errorf("Unable to start interactive mode: %s", err)
		}
		defer rl.Close()
		buf := bytes.NewBuffer([]byte(""))
		for {
			line, err := rl.Readline()
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			if line == "quit" {
				return nil
			}
			buf.WriteString(line)
			if strings.HasSuffix(line, ";") {
				if err := runQueries(db, buf); err != nil {
					fmt.Println(err)
				}
				buf.Reset()
			}
		}
		return nil
	}
	reader := bufio.NewReader(os.Stdin)
	return runQueries(db, reader)
}
