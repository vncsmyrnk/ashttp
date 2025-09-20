package internal

type Config struct {
	Domain  string
	Headers map[string]string
}

type DomainAlias string

type Configs map[DomainAlias]Config
