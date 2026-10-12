package mysql

import "fmt"

// DSN sets clientFoundRows so RowsAffected counts matched rows, not changed
// ones. Without it, an update that writes a column's current value reports
// zero rows, and the repos turn that into "not found".
func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&clientFoundRows=true",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}
