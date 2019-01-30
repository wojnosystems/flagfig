# Installation

go.mod:
```
require github.com/wojnosystems/flagfig v1.0.5
```

# Why?

Good question, I found myself writing flag/switch/config file/environment configuration code over and over again. Go has a really cool flag library, but it doesn't do environment or configuration files. So flagFig does all 3.

Additionally, the Nester interface makes it easy to embed/compose configurations so that you can build up your configs for other services (this is the part I kept re-writing). Just create your configuration as you would, then implement the RegisterFlags and AfterParse (if you desire), then use the power of Go's composition/embedding to avoid writing code over again.

This is immensely helpful for adding test-specific flags that should only appear in testing as well as creating re-usable flags in case you need multiple instances of the same type of configuration, as well as facilitation of the creation of a library of configuration object types.

# Usage



```go
import (
	"flag"
	"github.com/wojnosystems/flagfig"
)

// MySQLConfig is a quick and re-usable way of pulling in the configuration for a MySQL server
// Use NewMySQLConfig, do not instantiate yourself
type MySQLConfig struct {
	flagfig.Nester
	Config mysql.Config

	UserNameCfg, PasswordCfg, DbNameCfg, HostCfg flagfig.ConfigurableConfig
	userTmp, passwordTmp, nameTmp, hostTmp       *string
}
```

The variables named like *Cfg's allow this object to be re-used. You to change the flag and environment name of the object before calling RegisterFlags. By default, we'll specify what they should be in the "constructor" for this configuration object:

```go
// Initializes a MySQLConfig. You can modify the names of the fields prior to calling RegisterFlags
func NewMySQLConfig() *MySQLConfig {
	return &MySQLConfig{
		Config: mysql.Config{
			AllowNativePasswords: true,
		},
		UserNameCfg: flagfig.ConfigurableConfig{
			FlagName: "dbUser",
			EnvName:  "DB_USER",
		},
		PasswordCfg: flagfig.ConfigurableConfig{
			FlagName: "dbPassword",
			EnvName:  "DB_PASSWORD",
		},
		DbNameCfg: flagfig.ConfigurableConfig{
			FlagName: "dbName",
			EnvName:  "DB_NAME",
		},
		HostCfg: flagfig.ConfigurableConfig{
			FlagName: "dbHost",
			EnvName:  "DB_HOST",
		},
	}
}

// Call this before parse. This will allow you to customize the names of fields, if you desire
func (c *MySQLConfig) RegisterFlags( flags *flagfig.FlagfigSet ) {
	c.userTmp = flags.String(c.UserNameCfg.FlagName, "", c.UserNameCfg.EnvName, "the user to connect to MySQL with")
	c.passwordTmp = flags.String(c.PasswordCfg.FlagName, "", c.PasswordCfg.EnvName, "the password to connect to MySQL with")
	c.nameTmp = flags.String(c.DbNameCfg.FlagName, "", c.DbNameCfg.EnvName, "the database/schema to connect to")
	c.hostTmp = flags.String(c.HostCfg.FlagName, ":3306", c.HostCfg.EnvName, "the host address in the form of HOSTNAME:PORT or the unix socket address in the form of: /path/to/mysql.sock")
}

// AfterParsed is run after Parsing the flags. Any clean up or value coalescing that needs to be done should be done in this
func (c *MySQLConfig) AfterParsed() (err error) {
	c.Config.Net = "tcp"
	c.Config.Addr = *c.hostTmp
	if strings.HasPrefix(c.Config.Addr, "/") {
		c.Config.Net = "unix"
	}
	c.Config.User = *c.userTmp
	c.Config.Passwd = *c.passwordTmp
	c.Config.DBName = *c.nameTmp
	return nil
}
```

# Overriding Flag and Environment names

What if you want to use the same database config twice? You can't just use it as-is because the flags will collide. Fear not! Anyone can overwrite these values by doing:

```go
type SpecialConfig struct {
	flagfig.Nester
	
	Db1 *MySQLConfig
	Db2 *MySQLConfig
}

func NewSpecialConfig() *SpecialConfig {
	s := &SpecialConfig{
		Db1: NewMySQLConfig(),
		Db2: NewMySQLConfig(),
	}
	s.Db2.UserNameCfg.Flagname = "dbUser2"
	s.Db2.UserNameCfg.EnvName = "DB_USER2"
	s.Db2.PasswordCfg.Flagname = "dbPassword2"
	s.Db2.PasswordCfg.EnvName = "DB_PASSWORD2"
	s.Db2.DbNameCfg.Flagname = "dbName2"
	s.Db2.DbNameCfg.EnvName = "DB_NAME2"
	s.Db2.HostCfg.Flagname = "dbHost2"
	s.Db2.HostCfg.EnvName = "DB_HOST2"
	return s
}

func main() {
	cfg := NewSpecialConfig()
	MustParseNested(flag.PanicOnError, []Nester{cfg.Db1, cfg.Db2}, os.Args[1:])
}
```

Now your program will parse out two database configurations, one for the dbUser, and the other for the dbUser2, etc.