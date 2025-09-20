package internal

import (
	"fmt"
	"strings"
)

type QueryString map[string]string

func (q QueryString) ToURL() string {
	if q == nil {
		return ""
	}

	query := ""
	for k, v := range q {
		query = fmt.Sprintf("%s&%s=%s", query, k, v)
	}

	if query == "" {
		return query
	}

	return query[1:]
}

type PathComponents []string

func (p PathComponents) ToURL() string {
	return strings.Join(p, "/")
}

func Path(pathsComponents PathComponents, query QueryString) string {
	url := pathsComponents.ToURL()
	queryString := query.ToURL()
	if queryString == "" {
		return url
	}

	return fmt.Sprintf("%s?%s", url, queryString)
}
