package flagfig

// ConfigurableConfig allows for individual configurations to be renamed as needed
// When creating configurations, structure them with a function to "RegisterFlags" that occurs before the call to Parse, but after the New function is called. This way, you can allow implementors of your configurations to rename the flags that will be used while providing smart defaults
// Here's a quick example that creates a custom configuration with a value field named "value1" that allows an implementer to change its name
//
//type MyConfig struct {
//	flags *FlagfigSet
//
//	value1 *string
//	value1Config ConfigurableConfig
//}
//
//func NewMyConfig(upstreamFlags *FlagfigSet) *MyConfig {
//	return &MyConfig{
//		flags: upstreamFlags,
//		value1Config: ConfigurableConfig{
//			FlagName: "value1",
//			EnvName: "ENV_VALUE1",
//		},
//	}
//}
//
//func DefaultMyConfig() *MyConfig {
//	return NewMyConfig(NewFlagfigSet("my-config", flag.ContinueOnError))
//}
//
//func (m *MyConfig) RegisterFlags() {
//	m.value1 = m.flags.String( m.value1Config.FlagName, "default Value", m.value1Config.EnvName, "this is value1" )
//}
//
//func (m *MyConfig) Parse(args ...string) (err error) {
//	return m.flags.Parse(args)
//}
//
//type MyParentConfig struct {
//	flags *FlagfigSet
//
//	myConfig1 *MyConfig
//	myConfig2 *MyConfig
//}
//
//func NewMyParentConfig() MyParentConfig {
//	pc := MyParentConfig{
//		flags: NewFlagfigSet("parent-config", flag.ContinueOnError),
//	}
//	pc.myConfig1 = NewMyConfig(pc.flags)
//	pc.myConfig2 = NewMyConfig(pc.flags)
//	pc.myConfig2.value1Config.FlagName = "value2"
//	pc.myConfig2.value1Config.EnvName = "ENV_VALUE2"
//	pc.myConfig1.RegisterFlags()
//	pc.myConfig2.RegisterFlags()
//	return pc
//}
//
//func (m *MyParentConfig) Parse(args ...string) (err error) {
//	return m.flags.Parse(args)
//}
//
//When Parse is called, myConfig1 will have the value of: -value1 and myConfig2 will have the value of: -value2 from the command line and the configuration file. ENV_VALUE1 and ENV_VALUE2 will also be set for their respective values.
//
//You, of course, do not have to use this structure and can continue to define configuration flags as normal, but this enables greater flexibility for, say, if you need to use multiple databases in an application.
type ConfigurableConfig struct {
	// FlagName represents the name of the command line flag as well as the configuration file flag
	FlagName string
	// EnvName is the environment name for this value
	EnvName string
}

// NewConfigurableConfig convenience method for making ConfigurableConfig's
func NewConfigurableConfig(flagName, envName string) ConfigurableConfig {
	return ConfigurableConfig{
		FlagName: flagName,
		EnvName:  envName,
	}
}
