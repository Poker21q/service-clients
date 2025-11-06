package config

type Server struct {
	host string
	port string
}

func (c Server) Host() string {
	return c.host
}

func (c Server) Port() string {
	return c.port
}
