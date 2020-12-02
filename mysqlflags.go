package mysqlflags

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/percona/go-mysql/dsn"
	"github.com/vaughan0/go-ini"
)

// MyOpts mysql connection related flags used with go-flags
type MyOpts struct {
	MySQLDefaultsExtraFile string  `long:"defaults-extra-file" description:"path to defaults-extra-file"`
	MySQLSocket            string  `long:"mysql-socket" description:"path to mysql listen sock"`
	MySQLHost              string  `short:"H" long:"host" default:"localhost" description:"Hostname"`
	MySQLPort              string  `short:"p" long:"port" default:"3306" description:"Port"`
	MySQLUser              string  `short:"u" long:"user" default:"root" description:"Username"`
	MySQLPass              *string `short:"P" long:"password" description:"Password"`
	MySQLDBName            string  `long:"database" default:"" description:"database name connect to"`
	MySQLDSNParams         map[string]string
}

// CreateDSN creates DSN from Opts. omit timeout parameter when timeout is 0
func CreateDSN(opts MyOpts, timeout time.Duration, debug bool) (string, error) {
	dsn, err := dsn.Defaults("")
	if err != nil {
		return "", err
	}

	if opts.MySQLDefaultsExtraFile != "" {
		i, err := ini.LoadFile(opts.MySQLDefaultsExtraFile)
		if err != nil {
			return "", err
		}
		section := i.Section("client")
		user, ok := section["user"]
		if ok {
			dsn.Username = user
		}
		password, ok := section["password"]
		if ok {
			dsn.Password = password
		}
		socket, ok := section["socket"]
		if ok {
			dsn.Socket = socket
		}
		host, ok := section["host"]
		if ok {
			dsn.Hostname = host
		}
		port, ok := section["port"]
		if ok {
			dsn.Port = port
		}
	}
	if opts.MySQLHost != "" {
		dsn.Hostname = opts.MySQLHost
	}
	if opts.MySQLPort != "" {
		dsn.Port = opts.MySQLPort
	}
	if opts.MySQLUser != "" {
		dsn.Username = opts.MySQLUser
	}
	if opts.MySQLPass != nil {
		dsn.Password = *opts.MySQLPass
	}
	if opts.MySQLSocket != "" {
		dsn.Socket = opts.MySQLSocket
	}

	if dsn.Username == "" {
		dsn.Username = os.Getenv("USER")
	}
	dsn.DefaultDb = opts.MySQLDBName
	if timeout > 0 {
		dsn.Params = append(dsn.Params, fmt.Sprintf("timeout=%s", timeout.String()))
	}
	if opts.MySQLDSNParams != nil {
		for k, v := range opts.MySQLDSNParams {
			dsn.Params = append(dsn.Params, fmt.Sprintf("%s=%s", k, v))
		}
	}
	dsnString := dsn.String()
	if debug {
		dsn.Password = "xxxx"
		log.Printf("DSN: %s", dsn.String())
	}
	return dsnString, nil
}

// OpenDB opens MySQL connections from Opts
func OpenDB(opts MyOpts, timeout time.Duration, debug bool) (*sql.DB, error) {
	dsn, err := CreateDSN(opts, timeout, debug)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// QueryMapCol query and converts rows like show status to map[string]string
func QueryMapCol(db *sql.DB, query string, args ...interface{}) (map[string]string, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for rows.Next() {
		var n string
		var v string
		err := rows.Scan(&n, &v)
		if err != nil {
			return nil, err
		}
		result[n] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryMapRows query and converts rows like show slave status to []map[string]string
func QueryMapRows(db *sql.DB, query string, args ...interface{}) ([]map[string]string, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	c := make([]string, len(cols))
	for i, v := range cols {
		c[i] = v
	}
	result := []map[string]string{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		for index := range vals {
			vals[index] = new(sql.RawBytes)
		}
		err = rows.Scan(vals...)
		if err != nil {
			return nil, err
		}
		r := map[string]string{}
		for i := range vals {
			r[c[i]] = string(*vals[i].(*sql.RawBytes))
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, err
}
