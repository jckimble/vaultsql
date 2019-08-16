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
	"database/sql"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// mysqldumpCmd represents the mysqldump command
var mysqldumpCmd = &cobra.Command{
	Use:   "mysqldump",
	Short: "VaultSQL version of mysqldump",
	RunE:  wrapExec(mysqldump),
}

func init() {
	rootCmd.AddCommand(mysqldumpCmd)

	mysqldumpCmd.Flags().BoolP("databases", "B", false, "Dump several databases. Note the difference in usage; in this case no tables are given. All name arguments are regarded as database names. 'USE db_name;' will be included in the output.")
	mysqldumpCmd.Flags().BoolP("all-databases", "A", false, "Dump all the databases. This will be same as --databases with all databases selected.")
	mysqldumpCmd.Flags().Bool("add-drop-database", false, "Add a DROP DATABASE before each create.")
	mysqldumpCmd.Flags().Bool("add-drop-table", false, "Add a DROP TABLE before each create.")
}

func mysqldump(cmd *cobra.Command, args []string, db *sql.DB) error {
	databasesFlag, err := cmd.Flags().GetBool("databases")
	if err != nil {
		return err
	}
	allDatabasesFlag, err := cmd.Flags().GetBool("all-databases")
	if err != nil {
		return err
	}
	addDropDatabase, err := cmd.Flags().GetBool("add-drop-database")
	if err != nil {
		return err
	}
	addDropTable, err := cmd.Flags().GetBool("add-drop-table")
	if err != nil {
		return err
	}
	if !allDatabasesFlag {
		if len(args) < 1 {
			return fmt.Errorf("Database name is required")
		}
	}
	if allDatabasesFlag {
		dbs, err := runListQuery(db, "SHOW DATABASES;")
		if err != nil {
			return err
		}
		if err := backupDBs(db, dbs, addDropDatabase, addDropTable); err != nil {
			return err
		}
	} else if databasesFlag {
		if err := backupDBs(db, args, addDropDatabase, addDropTable); err != nil {
			return err
		}
	} else {
		if len(args) == 1 {
			if err := backupDB(db, args[0], addDropTable); err != nil {
				return err
			}
		} else {
			if err := useDatabase(db, args[0]); err != nil {
				return err
			}
			for _, table := range args[1:] {
				if addDropTable {
					fmt.Printf("DROP TABLE IF EXISTS `%s`;\n\n", table)
				}
				if err := backupTable(db, table); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func runListQuery(db *sql.DB, query string) ([]string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []string{}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		list = append(list, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func backupDBs(db *sql.DB, dbs []string, dropDB, dropTable bool) error {
	for _, dbname := range dbs {
		if dropDB {
			fmt.Printf("DROP DATABASE IF EXISTS `%s`;\n\n", dbname)
		}
		fmt.Printf("CREATE DATABASE IF NOT EXISTS `%s`;\n\n", dbname)
		fmt.Printf("USE `%s`;\n\n", dbname)
		if err := backupDB(db, dbname, dropTable); err != nil {
			return err
		}
	}
	return nil
}

func backupDB(db *sql.DB, dbname string, dropTable bool) error {
	if err := useDatabase(db, dbname); err != nil {
		return err
	}
	tables, err := runListQuery(db, "SHOW TABLES;")
	if err != nil {
		return err
	}
	for _, table := range tables {
		if dropTable {
			fmt.Printf("DROP TABLE IF EXISTS `%s`;\n\n", table)
		}
		if err := backupTable(db, table); err != nil {
			return err
		}
	}
	return nil
}

func backupTable(db *sql.DB, table string) error {
	row := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`;", table))
	var name, create string
	if err := row.Scan(&name, &create); err != nil {
		return err
	}
	fmt.Printf("%s;\n\n", create)

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`;", table))
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))
	valueptr := make([]interface{}, len(columns))
	for rows.Next() {
		placeholders := []string{}
		for i := range values {
			valueptr[i] = &values[i]
		}
		if err := rows.Scan(valueptr...); err != nil {
			return err
		}
		for i, val := range values {
			switch columnTypes[i].ScanType().Name() {
			case "NullString", "RawBytes":
				placeholders = append(placeholders, fmt.Sprintf("'%s'", val))
			case "NullInt64", "NullFloat64", "NullBool":
				placeholders = append(placeholders, fmt.Sprintf("%s", val))
			default:
				panic("Unknown Type: " + columnTypes[i].ScanType().Name())
			}
		}
		fmt.Printf("INSERT INTO `%s`(`%s`) VALUES(%s);\n", table, strings.Join(columns, "`,`"), strings.Join(placeholders, ","))
	}
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Printf("\n\n")
	return nil
}
