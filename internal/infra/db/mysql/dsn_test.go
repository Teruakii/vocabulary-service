package mysql

import (
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDSN_ReportsMatchedRows(t *testing.T) {
	parsed, err := mysql.ParseDSN(Config{Host: "db", Port: "3306", User: "u", DBName: "d"}.DSN())
	require.NoError(t, err)

	assert.True(t, parsed.ClientFoundRows)
	assert.True(t, parsed.ParseTime)
}
