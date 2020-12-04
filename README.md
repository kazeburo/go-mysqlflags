# go-mysqlflags

Utility for mysql related flags and connect and exec show statement to mysql.

# Usage

```
import "github.com/kazeburo/go-mysqlflags"
```

### use with go-flags and Connect to DB

Create DSN for connecting database.

```
type opts struct {
	mysqlflags.MyOpts
	Timeout time.Duration `long:"timeout" default:"10s" description:"Timeout to connect mysql"`
}

psr := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
_, err := psr.Parse()

dsn, err := mysqlflags.CreateDSN(opts.MyOpts, opts.Timeout, false)
```

When `my_print_defaults` command is available, mysqlflags.CreateDSN reads mysql client parameter from that. The parameters given with go-flags overwrite them.

Open Database with mysqlflags.

```
db, err := mysqlflags.OpenDB(opts.MyOpts, opts.Timeout, false)
if err != nil {
	log.Printf("couldn't connect DB: %v", err)
	return 1
}
defer db.Close()
```

### Exec show statement and get result as struct

show global status / show variables

```
type threads struct {
	Running   int64 `mysqlvar:"Threads_running"`
	Connected int64 `mysqlvar:"Threads_connected"`
	Cached    int64 `mysqlvar:"Threads_cached"`
}

type connections struct {
	Max       int64 `mysqlvar:"max_connections"`
	CacheSize int64 `mysqlvar:"thread_cache_size"`
}

var threads threads
err = mysqlflags.Query(db, "SHOW GLOBAL STATUS").Scan(&threads)
if err != nil {
	return err
}

var connections connections
err = mysqlflags.Query(db, "SHOW VARIABLES").Scan(&connections)
if err != nil {
	return err
}
```

show slave status (single source)

```
type slave struct {
	IORunning   mysqlflags.Bool `mysqlvar:"Slave_IO_Running"`
	SQLRunning  mysqlflags.Bool `mysqlvar:"Slave_SQL_Running"`
    LastSQLError string `mysqlvar:"Last_SQL_Error"`
}

var slave slave
err := mysqlflags.Query(db, "SHOW SLAVE STATUS").Scan(&slave)

f !slave.IORunning.Yes() || !slave.SQLRunning.Yes() {
    fmt.Errorf("something wrong is replication IO:%s SQL:%s Error:%s",
        slave.IORunning, slave.SQLRunning, slave.LastSQLError);
}

```


show slave status (multi source)

```
type slave struct {
	IORunning   mysqlflags.Bool `mysqlvar:"Slave_IO_Running"`
	SQLRunning  mysqlflags.Bool `mysqlvar:"Slave_SQL_Running"`
	ChannelName *string         `mysqlvar:"Channel_Name"` // use pointer for optinal field
	Behind      int64           `mysqlvar:"Seconds_Behind_Master"`
}


var slaves []slave
err := mysqlflags.Query(db, "SHOW SLAVE STATUS").Scan(&slaves)
for _, slave := range slaves {
    ..
}
```

