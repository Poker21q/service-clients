package config

type Auth struct {
	secret string
}

func (a Auth) Secret() string {
	return a.secret
}
