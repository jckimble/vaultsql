# VaultMySQL

VaultMySQL is a tool for accessing mysql databases that use Vault database secrets for authentication

---
* [Install](#install)
* [Configuration](#configuration)
* [Run](#run)
* [License](#license)

---

## Install
```sh
go get -u gitlab.com/jckimble/vaultsql/tools/vaultmysql
```

## Configuration
Uses all environment variables VaultSQL uses.

## Run
```sh
$vaultmysql
A mysql/mysqldump cmd for vault

Usage:
  vaultmysql [command]

Available Commands:
  help        Help about any command
  mysql       Uses vault token to login to mysql
  mysqldump   VaultSQL version of mysqldump

Flags:
  -h, --help           help for vaultmysql
      --host string    Mysql Host (default "127.0.0.1")
      --path string    Vault Secret Path
      --port int       Mysql Port (default 3306)
  -t, --token string   Vault Token

Use "vaultmysql [command] --help" for more information about a command.
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
