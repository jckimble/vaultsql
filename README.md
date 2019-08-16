# VaultSQL
[![pipeline status](https://gitlab.com/jckimble/vaultsql/badges/master/pipeline.svg)](https://gitlab.com/jckimble/vaultsql/commits/master)

VaultSQL is a sql driver for easily supporting Hashicorp vault database secrets without having to use sidecar containers

---
* [Install](#install)
* [Configuration](#configuration)
* [Usage](#usage)
* [License](#license)

---

## Install
```sh
go get -u gitlab.com/jckimble/vaultsql
```

## Configuration
Uses all environment variables Hashicorp Vault uses along with `VAULT_SECRET_PATH` to set the database credentials path. 

## Usage
VaultSQL discovers any loaded driver then wraps it, making it easy to use for any driver that Hashicorp Vault supports
```go
package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "gitlab.com/jckimble/vaultsql"
	"log"
	"time"
)

func main() {
	db, err := sql.Open("vault-mysql", "{{username}}:{{password}}@tcp(127.0.0.1:3306)/database")
	if err != nil {
		panic(err)
	}
	//TODO: Work with database here
}
```

## License

Copyright 2019 James Kimble

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
