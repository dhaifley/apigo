# apictl
A command line interface utility for API access.

## Usage

```
Usage: apictl [<option>] <command> <resource> [<query>]

Options:
  --help = Display this usage message
  --version = Display the command version
  --config.endpoint = Base endpoint URL of the API request
  --config.format = (json|yaml) Format of the command input and output
  --config.headers = Optional, HTTP headers to include with the API request
  --config.tls = Optional, TLS options to use for the API request
  
Commands:
  get
  post, create
  put, update
  patch
  delete
  option, head

Resources:
  Any resource or ID provided by the API. Multiple parameters will be combined
as path segments in the API request.

Query Parameters:
  Any parameters beginning with -- will be sent as query parameters with the API
request. For example, --param=value will be sent as ?param=value. Common query
parameters are:
  --search = Search expression
  --size = Number of results to request
  --skip = Offset starting point
  --sort = List of fields to sort by, descending fields have a "-" prefix
  --summary = List of fields to summarize by
```

## Examples
```sh
$ apictl --config.format='yaml' \
--config.endpoint='https://example.com/api/v1' \
--config.tls='{"InsecureSkipVerify":true}' \
--config.headers='{"Authorization":["token"]}' \
get user
```

```sh
created_at: 1721923211
created_by: null
data: null
email: dev@test.com
first_name: null
last_name: null
status: active
updated_at: 1721923211
updated_by: null
user_id: dev@test.com
```

## Building

```sh
$ go build -o apictl main.go
```
