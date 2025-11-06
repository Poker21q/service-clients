package config

type Postgres struct {
	host     string
	port     string
	user     string
	password string
	db       string
}

func (c Postgres) Host() string {
	return c.host
}

func (c Postgres) Port() string {
	return c.port
}

func (c Postgres) User() string {
	return c.user
}

func (c Postgres) Password() string {
	return c.password
}

func (c Postgres) DB() string {
	return c.db
}
