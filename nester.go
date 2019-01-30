package flagfig

import (
	"flag"
)

// Nester creates an expected interface to enable configuration objects to be nested and composed relatively painlessly.
// Stop writing your config code multiple times and start composing instead!
//
// A common pattern is to have nested configurations. Your main application may have a master configuration, that has one or MORE database configurations. When creating a command-line cli or other util, you may need some, but not all of those configuration objects. This interface lets you nest and compose those interfaces as needed. Here's an example of how to use it:
//

//// myConfig is the inner configuration. It goes into the MyServerConfig
//type myConfig struct {
//	NesterBase
//
//	MySecretSquare int
//
//	MyCoolString *string
//	MySecretNumber *int
//
//	MyCoolStringConf, MySecretNumberConf ConfigurableConfig
//}
//// Create a MyDbConfig with pre-defined names for fields. Put these in the New func to allow programmers to override this later
//// These ConfigurableConfigs are optional. You can hard-code the flags if you wish, but it's a good idea to be flexible.
//func newMyConfig() *myConfig {
//	return &myConfig{
//		MyCoolStringConf: ConfigurableConfig{
//			FlagName: "myCoolString",
//			EnvName: "COOL_STRING",
//		},
//		MySecretNumberConf: ConfigurableConfig{
//			FlagName: "mySecretNumber",
//			EnvName: "SECRET_NUMBER",
//		},
//	}
//}
//// RegisterFlags should always look like this. No heavy-lifting should be here, wait for AfterParsed or do that in your own code
//func (c *myConfig) RegisterFlags( flags *FlagfigSet ) {
//	c.MyCoolString = flags.String(c.MyCoolStringConf.FlagName, "ice cold, baby", c.MyCoolStringConf.EnvName, "a very groovy string for use with cool things")
//	c.MySecretNumber = flags.Int(c.MySecretNumberConf.FlagName, 42, c.MySecretNumberConf.EnvName, "the answer to everything")
//}
//// AfterParsed allows you to do some after-parsing processing/cleanup
//func (c *myConfig) AfterParsed() (err error) {
//	c.MySecretSquare = *c.MySecretNumber*(*c.MySecretNumber)
//	return nil
//}
//
//// Now that you have your inner-config, create your container configuration:
//type myServerConfig struct {
//	// This is also embeddable ;)
//	NesterBase
//
//	// This can be a pointer if you want, but this is a reference to your inner configuration
//	NestedConfig *myConfig
//
//	FullName string
//
//	FirstName *string
//	LastName *string
//
//	FirstNameConf, LastNameConf ConfigurableConfig
//}
//func newMyServerConfig() *myServerConfig {
//	return &myServerConfig{
//		FirstNameConf: ConfigurableConfig{
//			FlagName: "firstName",
//			EnvName: "FIRST_NAME",
//		},
//		LastNameConf: ConfigurableConfig{
//			FlagName: "lastName",
//			EnvName: "LAST_NAME",
//		},
//		NestedConfig: newMyConfig(),
//	}
//}
//// RegisterFlags should always look like this. No heavy-lifting should be here, wait for AfterParsed or do that in your own code
//func (c *myServerConfig) RegisterFlags( flags *FlagfigSet ) {
//	c.FirstName = flags.String(c.FirstNameConf.FlagName, "Chris", c.FirstNameConf.EnvName, "what do people call you?")
//	c.LastName = flags.String(c.LastNameConf.FlagName, "", c.LastNameConf.EnvName, "your last name?")
//}
//// AfterParsed is doing some error checking and some re-formatting to make it easier on ourselves later
//func (c *myServerConfig) AfterParsed() (err error) {
//	if len(*c.FirstName) == 0 || len(*c.LastName) == 0 {
//		return errors.New("hey, you don't get your gun until you tell me your name!")
//	}
//	sb := strings.Builder{}
//	sb.WriteString(*c.FirstName)
//	sb.WriteString(" ")
//	sb.WriteString(*c.LastName)
//	c.FullName = sb.String()
//	return nil
//}
//
//// Now, to parse these as a set, just do:
//func main() {
//	cfg := NewMyServerConfig()
//	MustParseNested(flag.PanicOnError, []Nester{cfg,cfg.NestedConfig}, os.Args[1:] )
//	fmt.Println("hey ", cfg.FullName, " guess my secret number! Here's a hint: ", cfg.NestedConfig.MySecretSquare)
//}
//
//
// Alternatively to using the ParseNested, you may, optionally, hide your inner configuration objects, but when RegisterFlags is called, you must call RegisterFlags on any embedded objects that support that method to ensure that their configuration data is picked up. Otherwise, you'll get a Panic or other error about missing or extra flags.

type Nester interface {
	// RegisterFlags is where you register your custom flags for your nested configuration
	// use the @param flags provided to attach your value listeners within your configuration object
	RegisterFlags(flags *FlagfigSet)
	// AfterParsed should be called right after Parse succeeds. Some configurations require that some processing takes place after the flags are parsed into values to make it easier to use the configuration in a final state later. Implementing this method is optional and should only occur if you need it. Do not execute any long-running or network processes in this. This should only be used to transform command-line/file/environment configurations into usable configuration values for the application
	AfterParsed() (err error)
}

// NesterBase is the basic case for a nested configuration. Nested configurations that wish to "inherit" this behavior should embed this type
type NesterBase struct {
}

// AfterParsed default is do nothing
func (n *NesterBase) AfterParsed() (err error) {
	return nil
}

// ParseNested will register the flags for each nestedConfig, then execute Parse on the flags and then run AfterParsed on every nestedConfig provided
func ParseNested(handling flag.ErrorHandling, nestedConfigs []Nester, args []string) (err error) {
	flags := NewFlagfigSet("", handling)
	for _, nc := range nestedConfigs {
		nc.RegisterFlags(flags)
	}
	err = flags.Parse(args)
	if err != nil {
		return err
	}
	for _, nc := range nestedConfigs {
		err = nc.AfterParsed()
		if err != nil {
			return err
		}
	}
	return nil
}

// MustParseNested does the same as ParseNested, but panics on error instead of returning the error
func MustParseNested(handling flag.ErrorHandling, nestedConfigs []Nester, args []string) {
	err := ParseNested(handling, nestedConfigs, args)
	if err != nil {
		panic(err)
	}
}
