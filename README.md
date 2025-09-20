# ashttp

A command-line HTTP client tool that simplifies API requests using configurable domain aliases and a CLI syntax.

ashttp is a Go-based HTTP client that allows you to make API requests using predefined domain aliases instead of full URLs. It's designed to streamline your workflow when working with multiple APIs by providing a simple, intuitive command-line interface with configuration-based domain management.

```bash
ashttp httpbin users 456 profile --include "posts,comments"

# Will be equivalent to:
# curl https://httpbin.dev/anything/users/456/profile?include=posts,comments
```

ashttp was created to solve the common overhead problem of managing multiple API endpoints with all sorts of authorization and specific headers.

## Usage

```bash
ashttp <domain-alias> [path-components...] [--query-key query-value...]
```

## Configuration

The configuration file is automatically created at `~/.config/ashttp/config.json` with a default httpbin example:

```json
{
  "httpbin": {
    "url": "https://httpbin.dev/anything",
    "defaultHeaders": {
      "authorization": "123"
    }
  }
}
```

Using this configuration, the command below demonstrates how ashttp translates to the equivalent curl request:

```bash
ashttp httpbin users 456 profile --include "posts,comments"

# Will be equivalent to:
# curl https://httpbin.dev/anything/users/456/profile?include=posts,comments \
#    -H "authorization: 123"
```

## Installation

```bash
go install github.com/ashttp@latest
```
