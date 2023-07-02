package drive

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

// XuguDriver is exported to make the driver directly accessible
type XuguDriver struct{}

/* Register Driver */
func init() {

	/* Register makes a database driver available by the provided name.
	 * If Register is called twice with the same name or if driver is nil,
	 * it panics.
	 */
	sql.Register("xugusql", &XuguDriver{})
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	time.Local = timezone
}

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a
// database name and connection information.
//
// Most users will open a database via a driver-specific connection
// helper function that returns a *DB. No database drivers are included
// in the Go standard library. See https://golang.org/s/sqldrivers for
// a list of third-party drivers.
//
// Open may just validate its arguments without creating a connection
// to the database. To verify that the data source name is valid, call
// Ping.
// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connections. Thus, the Open
// function should be called just once. It is rarely necessary to
// close a DB.
func (db XuguDriver) Open(dsn string) (driver.Conn, error) {
	conn := &connector{dsn: dsn}
	return conn.Connect(context.Background())
}

func (db XuguDriver) OpenConnector(dsn string) (driver.Connector, error) {
	return &connector{dsn: dsn}, nil
}
