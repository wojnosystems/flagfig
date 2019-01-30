package flagfig

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// myConfig is the inner configuration. It goes into the MyServerConfig
type myConfig struct {
	NesterBase

	MySecretSquare int

	MyCoolString   *string
	MySecretNumber *int

	MyCoolStringConf, MySecretNumberConf ConfigurableConfig
}

// Create a MyDbConfig with pre-defined names for fields. Put these in the New func to allow programmers to override this later
// These ConfigurableConfigs are optional. You can hard-code the flags if you wish, but it's a good idea to be flexible.
func newMyConfig() *myConfig {
	return &myConfig{
		MyCoolStringConf: ConfigurableConfig{
			FlagName: "myCoolString",
			EnvName:  "COOL_STRING",
		},
		MySecretNumberConf: ConfigurableConfig{
			FlagName: "mySecretNumber",
			EnvName:  "SECRET_NUMBER",
		},
	}
}

// RegisterFlags should always look like this. No heavy-lifting should be here, wait for AfterParsed or do that in your own code
func (c *myConfig) RegisterFlags(flags *FlagfigSet) {
	c.MyCoolString = flags.String(c.MyCoolStringConf.FlagName, "ice cold, baby", c.MyCoolStringConf.EnvName, "a very groovy string for use with cool things")
	c.MySecretNumber = flags.Int(c.MySecretNumberConf.FlagName, 42, c.MySecretNumberConf.EnvName, "the answer to everything")
}

// AfterParsed allows you to do some after-parsing processing/cleanup
func (c *myConfig) AfterParsed() (err error) {
	c.MySecretSquare = *c.MySecretNumber * (*c.MySecretNumber)
	return nil
}

// Now that you have your inner-config, create your container configuration:
type myServerConfig struct {
	// This is also embeddable ;)
	NesterBase

	// This can be a pointer if you want, but this is a reference to your inner configuration
	NestedConfig *myConfig

	FullName string

	FirstName *string
	LastName  *string

	FirstNameConf, LastNameConf ConfigurableConfig
}

func newMyServerConfig() *myServerConfig {
	return &myServerConfig{
		FirstNameConf: ConfigurableConfig{
			FlagName: "firstName",
			EnvName:  "FIRST_NAME",
		},
		LastNameConf: ConfigurableConfig{
			FlagName: "lastName",
			EnvName:  "LAST_NAME",
		},
		NestedConfig: newMyConfig(),
	}
}

// RegisterFlags should always look like this. No heavy-lifting should be here, wait for AfterParsed or do that in your own code
func (c *myServerConfig) RegisterFlags(flags *FlagfigSet) {
	c.FirstName = flags.String(c.FirstNameConf.FlagName, "Chris", c.FirstNameConf.EnvName, "what do people call you?")
	c.LastName = flags.String(c.LastNameConf.FlagName, "", c.LastNameConf.EnvName, "your last name?")
}

// AfterParsed is doing some error checking and some re-formatting to make it easier on ourselves later
func (c *myServerConfig) AfterParsed() (err error) {
	if len(*c.FirstName) == 0 || len(*c.LastName) == 0 {
		return errors.New("hey, you don't get your gun until you tell me your name!")
	}
	sb := strings.Builder{}
	sb.WriteString(*c.FirstName)
	sb.WriteString(" ")
	sb.WriteString(*c.LastName)
	c.FullName = sb.String()
	return nil
}

func TestNesterBase(t *testing.T) {
	fakeArgs := []string{"app.out", "-firstName", "Chris", "-lastName", "Wojno", "-mySecretNumber", "71"}
	cfg := newMyServerConfig()
	err := ParseNested(flag.PanicOnError, []Nester{cfg, cfg.NestedConfig}, fakeArgs[1:])
	if err != nil {
		t.Error("did not expect an error, but got: ", err)
	}
	expected := "hey Chris Wojno, guess my secret number! Here's a hint: 5041"
	actual := fmt.Sprint("hey ", cfg.FullName, ", guess my secret number! Here's a hint: ", strconv.Itoa(cfg.NestedConfig.MySecretSquare))
	if strings.Compare(expected, actual) != 0 {
		t.Error("strings did not match, expected '", expected, "' but got '", actual, "'")
	}
}
