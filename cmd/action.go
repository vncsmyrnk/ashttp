package main

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/ashttp/internal/config"
	internalhttp "github.com/ashttp/internal/http"
)

type Action struct {
	URLAlias          string
	HTTPMethod        string
	URLPathComponents []string
	Options           map[string]string
}

var acceptedMethods = func() []string {
	methods := []string{http.MethodGet, http.MethodDelete}
	for i := range methods {
		methods[i] = strings.ToLower(methods[i])
	}
	return methods
}()

var errInvalidFormat = errors.New("invalid format")

func NewAction(args []string) (Action, error) {
	if len(args) < 2 {
		return Action{}, fmt.Errorf("%w: empty arguments", errInvalidFormat)
	}

	httpMethod := args[1]
	if err := validateHTTPMethod(httpMethod); err != nil {
		return Action{}, fmt.Errorf("unsuported http method: %w", err)
	}

	request := Action{
		URLAlias:          args[0],
		HTTPMethod:        httpMethod,
		URLPathComponents: make([]string, 0, len(args)),
		Options:           make(map[string]string, len(args)),
	}

	readingFlag := false
	lastFlag := ""
	for _, arg := range args[2:] {
		if strings.HasPrefix(arg, "--") {
			readingFlag = true
			lastFlag = strings.TrimPrefix(arg, "--")
			request.Options[lastFlag] = ""
			continue
		}

		if readingFlag {
			request.Options[lastFlag] = arg
			continue
		}

		request.URLPathComponents = append(request.URLPathComponents, arg)
	}

	return request, nil
}

func (a Action) Request() internalhttp.Request {
	pathCompnents := internalhttp.PathComponents(a.URLPathComponents)

	return internalhttp.Request{
		Path:      strings.Join(pathCompnents, "/"),
		Method:    a.HTTPMethod,
		Arguments: a.Options,
	}
}

func (a Action) Setting() (config.Setting, error) {
	settings, err := config.GetSettings()
	if err != nil {
		return config.Setting{}, err
	}

	urlAlias := config.URLAlias(a.URLAlias)
	if config, ok := settings[urlAlias]; ok {
		return config, nil
	}

	return config.Setting{}, fmt.Errorf(
		"no config found for %s, make sure it exists at %s", urlAlias, config.GetDefaultConfigPath())
}

func validateHTTPMethod(method string) error {
	if !slices.Contains(acceptedMethods, method) {
		return fmt.Errorf("invalid http method, only %s are supported", strings.Join(acceptedMethods, ", "))
	}
	return nil
}
