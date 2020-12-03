package mysqlflags

import (
	"reflect"
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

func TestQueryMapCol(t *testing.T) {
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

	r, err := QueryMapCol(db, "SHOW GLOBAL STATUS")
	assert.NoError(t, err)
	want := map[string]string{
		"Uptime":         "941",
		"Current_tls_ca": "ca.pem",
	}
	if !reflect.DeepEqual(r, want) {
		t.Fatalf("expected: %v\nactual: %v", want, r)
	}
}

func TestQueryMapRows(t *testing.T) {
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

	r, err := QueryMapRows(db, "SHOW SLAVE STATUS")
	assert.NoError(t, err)
	want := []map[string]string{
		{"Master_Host": "db1",
			"Master_User": "user1",
			"Master_Port": "3306"},
		{"Master_Host": "db2",
			"Master_User": "user2",
			"Master_Port": "3306"},
		{"Master_Host": "db3",
			"Master_User": "user3",
			"Master_Port": "3306"},
	}
	if !reflect.DeepEqual(r, want) {
		t.Fatalf("expected: %v\nactual: %v", want, r)
	}
}

type QueryColSt struct {
	Uptime int    `mysqlval:"Uptime"`
	Tlsca  string `mysqlval:"Current_tls_ca"`
}

type QueryColStExtra struct {
	Uptime int    `mysqlval:"Uptime"`
	Hoge   string `mysqlval:"Hoge"`
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
}

type QueryRowSt struct {
	Host string `mysqlval:"Master_Host"`
	User string `mysqlval:"Master_User"`
	Port int    `mysqlval:"Master_Port"`
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
