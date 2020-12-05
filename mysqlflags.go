package mysqlflags

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kazeburo/mapstructure"
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

type dsn struct {
	Username  string
	Password  string
	Hostname  string
	Port      string
	Socket    string
	DefaultDB string
	Params    map[string]string
}

func newDSN() dsn {
	dsn := dsn{}
	dsn.Params = map[string]string{}
	return dsn
}

func (dsn *dsn) String() string {
	dsnString := ""

	if dsn.Socket != "" {
		dsnString = fmt.Sprintf("%s:%s@unix(%s)",
			dsn.Username,
			dsn.Password,
			dsn.Socket,
		)
	} else {
		if dsn.Hostname == "" {
			dsn.Hostname = "localhost"
		}
		if dsn.Port == "" {
			dsn.Port = "3306"
		}
		dsnString = fmt.Sprintf("%s:%s@tcp(%s:%s)",
			dsn.Username,
			dsn.Password,
			dsn.Hostname,
			dsn.Port,
		)
	}

	dsnString += "/" + dsn.DefaultDB
	p := []string{}
	for k, v := range dsn.Params {
		p = append(p, fmt.Sprintf("%s=%s", k, v))
	}
	params := strings.Join(p, "&")
	if params != "" {
		dsnString += "?" + params
	}
	return dsnString
}

var defautlsRegexp = regexp.MustCompile("^--(.+?)=(.*)")

func parseMySQLDefaults(r io.Reader) dsn {
	dsn := newDSN()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		res := defautlsRegexp.FindAllSubmatch(b, -1)
		if res != nil {
			key := string(res[0][1])
			val := string(res[0][2])
			switch key {
			case "user":
				dsn.Username = val
			case "password":
				dsn.Password = val
			case "socket":
				dsn.Socket = val
			case "host":
				dsn.Hostname = val
			case "port":
				dsn.Port = val
			}
		}
	}
	if dsn.Socket != "" {
		dsn.Port = ""
		dsn.Hostname = ""
	}
	if dsn.Hostname == "" && dsn.Socket == "" {
		dsn.Hostname = "localhost"
	}
	return dsn
}

func defaults() dsn {
	params := [][]string{
		{"-s", "client"},
		{"client"},
	}
	for _, param := range params {
		outout, err := exec.Command("my_print_defaults", param...).Output()
		if err != nil {
			continue
		}
		return parseMySQLDefaults(bytes.NewReader(outout))
	}
	return newDSN()
}

// CreateDSN creates DSN from Opts. omit timeout parameter when timeout is 0
func CreateDSN(opts MyOpts, timeout time.Duration, debug bool) (string, error) {
	dsn := defaults()

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
	dsn.DefaultDB = opts.MySQLDBName
	if timeout > 0 {
		dsn.Params["timeout"] = timeout.String()
	}
	if opts.MySQLDSNParams != nil {
		for k, v := range opts.MySQLDSNParams {
			dsn.Params[k] = v
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

// Bool mysql variables boolean
type Bool struct {
	original string
}

// True true when the original string
func (b Bool) String() string {
	return b.original
}

// Yes return true when the original is Yes or yes or YES
func (b Bool) Yes() bool {
	switch b.original {
	case "Yes", "yes", "YES", "On", "ON", "on":
		return true
	default:
		return false
	}
}

// QueryMap has rows map and error
type QueryMap struct {
	err    error
	result []map[string]string
}

// Scan converts rows to Struct
func (qm *QueryMap) Scan(dest interface{}) error {
	if qm.err != nil {
		return qm.err
	}

	destRv := reflect.ValueOf(dest)
	if destRv.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer: %v", dest)
	}

	destRv = destRv.Elem()
	var input interface{}
	if destRv.Kind() != reflect.Slice {
		if len(qm.result) == 0 {
			return fmt.Errorf("no sql result")
		}
		input = qm.result[0]
	} else {
		input = qm.result
	}

	decodeHook := func(f, t reflect.Type, v interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t.Kind() == reflect.Bool {
			if typed, ok := v.(string); ok {
				switch typed {
				case "Yes", "yes", "YES", "On", "ON", "on":
					return "True", nil
				default:
					return "False", nil
				}
			}
		} else if f.Kind() == reflect.String && t.String() == "mysqlflags.Bool" {
			if typed, ok := v.(string); ok {
				return Bool{typed}, nil
			}
		}

		return v, nil
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		ErrorUnsetFields: true,
		Result:           dest,
		TagName:          "mysqlvar",
		DecodeHook:       decodeHook,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	err = decoder.Decode(input)
	if err != nil {
		return err
	}
	return nil
}

// Query does exec show statement and return QueryMap for Scan
func Query(db *sql.DB, query string, args ...interface{}) *QueryMap {
	rows, err := db.Query(query, args...)
	if err != nil {
		return &QueryMap{err: err}
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return &QueryMap{err: err}
	}
	c := make([]string, len(cols))
	for i, v := range cols {
		c[i] = v
	}
	if len(cols) == 2 && c[0] == "Variable_name" && c[1] == "Value" {
		// show status | show variables
		return queryCol(c, rows)
	}
	return queryRow(c, rows)
}

func queryCol(c []string, rows *sql.Rows) *QueryMap {
	r := map[string]string{}
	for rows.Next() {
		var n string
		var v string
		err := rows.Scan(&n, &v)
		if err != nil {
			return &QueryMap{err: err}
		}
		r[n] = v
	}
	if err := rows.Err(); err != nil {
		return &QueryMap{err: err}
	}
	result := []map[string]string{}
	result = append(result, r)
	return &QueryMap{result: result, err: nil}
}

func queryRow(c []string, rows *sql.Rows) *QueryMap {
	result := []map[string]string{}
	for rows.Next() {
		vals := make([]interface{}, len(c))
		for index := range vals {
			vals[index] = new(sql.RawBytes)
		}
		err := rows.Scan(vals...)
		if err != nil {
			return &QueryMap{err: err}
		}
		r := map[string]string{}
		for i := range vals {
			r[c[i]] = string(*vals[i].(*sql.RawBytes))
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return &QueryMap{err: err}
	}
	return &QueryMap{result: result, err: nil}
}
