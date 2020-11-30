package mysqlflags

import (
	"testing"
	"time"

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
	assert.Equal(t, "testuser:testpass@tcp(example.com:33306)/?interpolateParams=true&timeout=1s", dsn)
}
