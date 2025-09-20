package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ashttp/internal/config"
	internalhttp "github.com/ashttp/internal/http"
)

type Action struct {
	DomainAlias       string
	URLPathComponents []string
	URLQuery          map[string]string
}

func NewAction(args []string) (Action, error) {
	if len(args) == 0 {
		return Action{}, fmt.Errorf("empty arguments")
	}

	request := Action{
		DomainAlias:       args[0],
		URLPathComponents: make([]string, 0, len(args)),
		URLQuery:          make(map[string]string, len(args)),
	}

	readingFlag := false
	lastFlag := ""
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "--") {
			readingFlag = true
			lastFlag = strings.TrimPrefix(arg, "--")
			request.URLQuery[lastFlag] = ""
			continue
		}

		if readingFlag {
			request.URLQuery[lastFlag] = arg
			continue
		}

		request.URLPathComponents = append(request.URLPathComponents, arg)
	}

	return request, nil
}

func (a Action) Request() internalhttp.Request {
	pathCompnents := internalhttp.PathComponents(a.URLPathComponents)
	queryString := internalhttp.QueryString(a.URLQuery)

	return internalhttp.Request{
		Path:       internalhttp.Path(pathCompnents, queryString),
		HTTPMethod: http.MethodGet,
	}
}

func (a Action) Config() (config.Config, error) {
	configs, err := config.GetConfigs()
	if err != nil {
		return config.Config{}, err
	}

	domainAlias := config.DomainAlias(a.DomainAlias)
	if config, ok := configs[domainAlias]; ok {
		return config, nil
	}

	return config.Config{}, fmt.Errorf("failed to get config")
}
