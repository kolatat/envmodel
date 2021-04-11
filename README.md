# envmodel

Managing configuration data from environment variables using object model.

## Usage

Declare model and parse.

```go
package main

import (
	"log"
	"time"

	"github.com/kolatat/envmodel"
)

// Declare the object model
type Config struct {
	Name    string
	Timeout time.Duration
	Debug   bool
}

func main() {
	var cfg Config
	// Parse environment variables into your object
	if _, err := envmodel.Parse(&cfg, &envmodel.Option{
		AppName: "MYAPP",
		DotEnv:  true,
	}); err != nil {
		log.Fatalf("parsing config: %s", err)
	}
}
```

The parser will read the field names, and convert the PascalCase names into UPPERCASE_SNAKE_CASE names.

### Options

#### Namespace

`Namespace` - prefixes all key names with the specified namespace. For example, when using `Namespace: "DB"`:

    HostName string     => "DB_HOST_NAME"
    Database string     => "DB_DATABASE"

#### DotEnv

`DotEnv: true` - Additionally loads environment variables from ~.env files. See https://github.com/joho/godotenv.

`AppName: "MYAPP"` - The environment variable `{AppName}_ENV`, is checked to establish the runtime environment of the
app. If it is empty or unset, a `development` environment is assumed.

DotEnv will check for and load the following files:

1. .env.{environment}.local
2. .env.local
3. .env.{environment}
4. .env

Missing files will be skipped, and existing variables take precedence.

### Struct Tag

Tag key name: `env`

```go
package main

import "time"

type Config struct {
	Name    string        `env:"DISPLAY"`
	Timeout time.Duration `env:"TIMEOUT,required,default:5s"`
	Wtv     interface{}   `env:"-"`
}
```

The `env` tag is a comma separated list of attributes. The first attribute is always the name of the environment
variable to parse from. Use `-` as the name and the field will be ignored by the parser.

## References

- https://github.com/joho/godotenv
- https://github.com/kelseyhightower/envconfig
