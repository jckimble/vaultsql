package vaultsql

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	//"database/sql"
	"database/sql/driver"
)

type testDriver string

func (d testDriver) Open(dsn string) (driver.Conn, error) {
	if string(d) != dsn {
		return nil, errors.New(string(d) + " != " + dsn)
	}
	return nil, nil
}

func TestVaultSQL(t *testing.T) {
	ts := newTestServer()
	os.Setenv("VAULT_ADDR", ts.URL)
	os.Setenv("VAULT_SECRET_PATH", "database/creds/test")
	vd := vaultDriver{testDriver("user:pass@tcp(127.0.0.1:3306)/database")}
	conn, err := vd.OpenConnector("{{username}}:{{password}}@tcp(127.0.0.1:3306)/database")
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Connect(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func newTestServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
  "request_id": "e0e5a6c1-5e69-5cf3-c9d2-020af192de36",
  "lease_id": "database/creds/readonly/7aa462ab-98cb-fdcb-b226-f0a0d37644cc",
  "renewable": true,
  "lease_duration": 20,
  "data": {
    "password": "pass",
    "username": "user"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}`)
	}))
	return ts
}
