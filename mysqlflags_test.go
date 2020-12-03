package mysqlflags

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateDSN(t *testing.T) {
	timeout := 1 * time.Second
	password := "testpass"
	opts := MyOpts{
		MySQLDefaultsExtraFile: "",
		MySQLSocket:            "",
		MySQLHost:              "example.com",
		MySQLPort:              "33306",
		MySQLUser:              "testuser",
		MySQLPass:              &password,
		MySQLDBName:            "",
	}
	dsn, err := CreateDSN(opts, timeout, false)
	assert.NoError(t, err)
	assert.Equal(t, "testuser:testpass@tcp(example.com:33306)/?timeout=1s", dsn)

	opts.MySQLDSNParams = map[string]string{
		"readTimeout": timeout.String(),
	}
	dsn, err = CreateDSN(opts, timeout, false)
	assert.NoError(t, err)
	assert.Equal(t, "testuser:testpass@tcp(example.com:33306)/?timeout=1s&readTimeout=1s", dsn)

	dsn, err = CreateDSN(opts, 0*time.Second, false)
	assert.NoError(t, err)
	assert.Equal(t, "testuser:testpass@tcp(example.com:33306)/?readTimeout=1s", dsn)
}

type QueryColSt struct {
	Uptime int    `mysqlvar:"Uptime"`
	Tlsca  string `mysqlvar:"Current_tls_ca"`
}

type QueryColStExtra struct {
	Uptime int    `mysqlvar:"Uptime"`
	Hoge   string `mysqlvar:"Hoge"`
}

type QueryColStBool struct {
	Uptime  int  `mysqlvar:"Uptime"`
	Running bool `mysqlvar:"Running"`
	Live    bool `mysqlvar:"Live"`
}

func TestQueryCol(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to init db mock")
	}
	defer db.Close()

	columns := []string{"Variable_name", "Value"}
	mock.ExpectQuery("SHOW").
		WillReturnRows(
			sqlmock.NewRows(columns).
				AddRow("Uptime", 941).
				AddRow("Current_tls_ca", "ca.pem"),
		)

	var result QueryColSt
	err = Query(db, "SHOW GLOBAL STATUS").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 941, result.Uptime)
	assert.Equal(t, "ca.pem", result.Tlsca)

	mock.ExpectQuery("SHOW").
		WillReturnRows(
			sqlmock.NewRows(columns).
				AddRow("Uptime", 941).
				AddRow("Current_tls_ca", "ca.pem"),
		)

	var result2 QueryColStExtra
	err = Query(db, "SHOW GLOBAL STATUS").Scan(&result2)
	assert.Error(t, err)

	mock.ExpectQuery("SHOW").
		WillReturnRows(
			sqlmock.NewRows(columns).
				AddRow("Uptime", 941).
				AddRow("Running", "Yes").
				AddRow("Live", "Nes"),
		)
	var result3 QueryColStBool
	err = Query(db, "SHOW GLOBAL STATUS").Scan(&result3)
	assert.NoError(t, err)
	assert.Equal(t, true, result3.Running)
	assert.Equal(t, false, result3.Live)

}

type QueryRowSt struct {
	Host string `mysqlvar:"Master_Host"`
	User string `mysqlvar:"Master_User"`
	Port int    `mysqlvar:"Master_Port"`
}

func TestQueryRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to init db mock")
	}
	defer db.Close()

	columns := []string{"Master_Host", "Master_User", "Master_Port"}
	mock.ExpectQuery("SHOW").
		WillReturnRows(
			sqlmock.NewRows(columns).
				AddRow("db1", "user1", 3306).
				AddRow("db2", "user2", 3306).
				AddRow("db3", "user3", 3306),
		)

	var result1 QueryRowSt
	err = Query(db, "SHOW SLAVE STATUS").Scan(&result1)
	assert.NoError(t, err)
	assert.Equal(t, "db1", result1.Host)
	assert.Equal(t, "user1", result1.User)
	assert.Equal(t, 3306, result1.Port)

	mock.ExpectQuery("SHOW").
		WillReturnRows(
			sqlmock.NewRows(columns).
				AddRow("db1", "user1", 3306).
				AddRow("db2", "user2", 3306).
				AddRow("db3", "user3", 3306),
		)

	var result2 []QueryRowSt
	err = Query(db, "SHOW SLAVE STATUS").Scan(&result2)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(result2))
	assert.Equal(t, result2[0].Host, "db1")
	assert.Equal(t, result2[0].User, "user1")
	assert.Equal(t, result2[0].Port, 3306)
	assert.Equal(t, result2[2].Host, "db3")
	assert.Equal(t, result2[2].User, "user3")
	assert.Equal(t, result2[2].Port, 3306)

}
