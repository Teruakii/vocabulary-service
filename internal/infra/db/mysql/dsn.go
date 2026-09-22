package mysql

import "fmt"

func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&clientFoundRows=true",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}
