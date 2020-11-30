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

// Opts mysql connection related flags used with go-flags
type Opts struct {
	MySQLDefaultsExtraFile string  `long:"defaults-extra-file" description:"path to defaults-extra-file"`
	MySQLSocket            string  `long:"mysql-socket" description:"path to mysql listen sock"`
	MySQLHost              string  `short:"H" long:"host" default:"localhost" description:"Hostname"`
	MySQLPort              string  `short:"p" long:"port" default:"3306" description:"Port"`
	MySQLUser              string  `short:"u" long:"user" default:"root" description:"Username"`
	MySQLPass              *string `short:"P" long:"password" description:"Password"`
	MySQLDBName            string  `long:"database" default:"queue_mercari" description:"database name connect to"`
}

// CreateDSN creates DSN from Opts
func CreateDSN(opts Opts, timeout time.Duration, debug bool) (string, error) {
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
	dsn.Params = append(dsn.Params, "interpolateParams=true")
	dsn.Params = append(dsn.Params, fmt.Sprintf("timeout=%s", timeout.String()))
	dsnString := dsn.String()
	if debug {
		dsn.Password = "xxxx"
		log.Printf("DSN: %s", dsn.String())
	}
	return dsnString, nil
}

// OpenDB opens MySQL connections from Opts
func OpenDB(opts Opts, timeout time.Duration, debug bool) (*sql.DB, error) {
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
