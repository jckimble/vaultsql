// Package vaultsql provides a sql driver for easily supporting hashicorp vault database secrets without having to use sidecar containers
//
// The driver should be used via the database/sql package along with a vault supported driver
//
package vaultsql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/vault/api"
)

func init() {
	for _, driver := range sql.Drivers() {
		db, err := sql.Open(driver, "")
		if err != nil {
			panic(err)
		}
		sql.Register("vault-"+driver, vaultDriver{db.Driver()})
	}
}

type vaultDriver struct {
	driver driver.Driver
}

func (d vaultDriver) OpenConnector(dsn string) (driver.Connector, error) {
	config := api.DefaultConfig()
	if err := config.ReadEnvironment(); err != nil {
		return nil, err
	}
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	vc := &vaultConnector{
		driver:     d.driver,
		rwlock:     &sync.RWMutex{},
		dsn:        dsn,
		client:     client,
		secretPath: os.Getenv("VAULT_SECRET_PATH"),
	}
	return vc, vc.vaultSecretRenewer()
}

func (d vaultDriver) Open(_ string) (driver.Conn, error) {
	return nil, errors.New("Please use OpenConnector")
}

type vaultConnector struct {
	driver     driver.Driver
	rwlock     *sync.RWMutex
	dsn        string
	client     *api.Client
	secretPath string

	data map[string]interface{}
}

func (c vaultConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.driver.Open(c.parseDSN())
}

func (c vaultConnector) Driver() driver.Driver {
	return c.driver
}

func (c *vaultConnector) updateData(data map[string]interface{}) {
	c.rwlock.Lock()
	defer c.rwlock.Unlock()
	if c.data == nil {
		c.data = data
	} else {
		for k, v := range data {
			c.data[k] = v
		}
	}
}

func (c vaultConnector) parseDSN() string {
	c.rwlock.RLock()
	defer c.rwlock.RUnlock()
	dsn := c.dsn
	if c.data != nil {
		for k, v := range c.data {
			if sr, ok := v.(string); ok {
				dsn = strings.ReplaceAll(dsn, "{{"+k+"}}", sr)
			}
		}
	}
	return dsn
}

func (c *vaultConnector) vaultSecretRenewer() error {
	secret, err := c.client.Logical().Read(c.secretPath)
	if err != nil {
		return err
	}
	c.updateData(secret.Data)
	renewer, err := c.client.NewRenewer(&api.RenewerInput{
		Secret: secret,
	})
	if err != nil {
		return err
	}
	go renewer.Renew()
	go func() {
		defer renewer.Stop()
		for {
			select {
			case err := <-renewer.DoneCh():
				if err != nil {
					panic(err)
				} else {
					go c.vaultSecretRenewer()
				}
			case renewal := <-renewer.RenewCh():
				secr := renewal.Secret
				c.updateData(secr.Data)
			}
		}
	}()
	return nil
}
