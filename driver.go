package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
)

var Pool *mockDriver

func init() {
	Pool = &mockDriver{
		conns: make(map[string]*sqlmock),
	}
	sql.Register("sqlmock", Pool)
}

type mockDriver struct {
	sync.Mutex
	counter int
	conns   map[string]*sqlmock
}

func (d *mockDriver) Open(dsn string) (driver.Conn, error) {
	d.Lock()
	defer d.Unlock()

	c, ok := d.conns[dsn]
	if !ok {
		return c, fmt.Errorf("expected a connection to be available, but it is not")
	}

	c.opened++
	return c, nil
}

// New creates sqlmock database connection
// and a mock to manage expectations.
// Pings db so that all expectations could be
// asserted.
func New() (*sql.DB, Sqlmock, error) {
	Pool.Lock()
	dsn := fmt.Sprintf("sqlmock_db_%d", Pool.counter)
	Pool.counter++

	smock := &sqlmock{dsn: dsn, drv: Pool, ordered: true}
	Pool.conns[dsn] = smock
	Pool.Unlock()

	return smock.open()
}

// NewWithDSN creates sqlmock database connection
// with a specific DSN and a mock to manage expectations.
// Pings db so that all expectations could be asserted.
//
// This method is introduced because of sql abstraction
// libraries, which do not provide a way to initialize
// with sql.DB instance. For example GORM library.
//
// Note, it will error if attempted to create with an
// already used dsn
//
// It is not recommended to use this method, unless you
// really need it and there is no other way around.
func NewWithDSN(dsn string) (*sql.DB, Sqlmock, error) {
	Pool.Lock()
	if _, ok := Pool.conns[dsn]; ok {
		Pool.Unlock()
		return nil, nil, fmt.Errorf("cannot create a new mock database with the same dsn: %s", dsn)
	}
	smock := &sqlmock{dsn: dsn, drv: Pool, ordered: true}
	Pool.conns[dsn] = smock
	Pool.Unlock()

	return smock.open()
}
