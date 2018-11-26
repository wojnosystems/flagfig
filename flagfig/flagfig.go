// Copyright 2018 Chris Wojno. All rights reserved.

/*
	Package flagfig implements configuration file, environment, and command-line flag parsing

	Usage

	Define flags using: (more types are planned to be added later)
		flagfig.String()
	then follow that with:
		flagfig.Parse()

	Most of this behavior was modelled after GoLang's flag package.

	Flaguration is a simple way to easily combine flags, environment settings, and configuration files in the Go-way
	Once configured, it will read the configuration files from the command line, read the params from the command line,
	and overwrite values in the following, Go-way:
	1. Load configuration files (last value set wins)
	2. Load Env variable (if program opts to do it, if left off, variables will not be queried)
	3. Load the flags from the command line

	Each step will overwrite the previous step, so command-line flags always win. Env variables always take precidence
	over configuration files, etc.

	Use this in the same manner as Golang's flag package.

	Examples

		httpAddr := flagfig.String("httpaddr", DefaultHttpAddress, "MYAPP_HTTP_ADDR", "http address [" + DefaultHttpAddress + "]")
		httpsAddr := flagfig.String("httpsaddr", DefaultHttpsAddress, "MYAPP_HTTPS_ADDR","https address [" + DefaultHttpsAddress + "]")
		certPath := flagfig.String("tlscertpath", DefaultTLSCertPath, "MYAPP_TLS_CERT_PATH","file path to tls cert ]" + DefaultTLSCertPath + "]")
		tlsKeyPath := flagfig.String("tlskeypath", "", "MYAPP_TLS_KEY_PATH","file path to tls key (required)")
		flagfig.AddConfigFile("config","file path to configuration JSON file")
		flagfig.Parse()

	Running:
		go run -config=/path/to/config

	Will load the configuration file path. The environment variables will be loaded, then the flags will be installed, if set.

	Configuration File format

	At the time of this writing, this library ONLY handles JSON files that are flat (have only a single object)
	and only work with string keys and string values. So you can use this:

	{
		flag1: "value",
        flag2: "anothervalue"
	}

	But you cannot use a file like this:

	{
		flag1: {
			complexItem: 4
		}
	}

	Again, it only supports string keys and values; and only a single level of strings with values

	Environment Variables

	If you do not define an environment variable name, it will not be parsed. This allows you to not include parsing
	an environment variable if you do not with to use it. If you don't want to parse it, just toss in an empty string ("")

	That's a stupid name...

	flagfig is a portmanteau of flag and config... If you have to explain it, I guess...
 */
package flagfig

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var CommandLine = NewFlagfigSet(os.Args[0], flag.ExitOnError)

// FlagurationSet
type FlagfigSet struct {
	flag.FlagSet
	configFilePaths    []*string
	fileConfigurations map[string]string
	envNames map[string]string

	stringValues map[string]*string
}

func NewFlagfigSet(name string, errorHandling flag.ErrorHandling) *FlagfigSet {
	fs := &FlagfigSet{	}
	fs.FlagSet = *flag.NewFlagSet(name, errorHandling)
	fs.configFilePaths = make([]*string,0,1)
	fs.envNames = make(map[string]string)
	fs.stringValues = make(map[string]*string)
	return fs
}

func Parse() {
	_ = CommandLine.Parse(os.Args[1:])
}

func (f *FlagfigSet)Parse( arguments []string ) (err error) {
	err = f.FlagSet.Parse(arguments)
	if err == nil {
		f.Collate()
	}
	return
}

// AddConfigFile adds a configuration file flag to the command line
// When Parse() is called, this file will be added to the list of files to parse when looking for configuration values
// name is the flagname
func AddConfigFile( name, usage string ) *string {
	return CommandLine.AddConfigFile( name, usage )
}
func (f *FlagfigSet)AddConfigFile( name, usage string ) *string {
	p := new(string)
	f.configFilePaths = append( f.configFilePaths, p )
	f.FlagSet.StringVar( p, name, "", usage)
	return p
}

// Collate combines the values from config files, environment variables, and flags as a single value.
// Assumes that the command flags are already parsed
func (f *FlagfigSet) Collate() {
	unVisitedFlags := make([]*flag.Flag,0,10)
	allFlags := make(map[string]bool)
	f.FlagSet.VisitAll(func (fl *flag.Flag) {
		allFlags[fl.Name] = false
	})
	f.FlagSet.Visit(func (fl *flag.Flag) {
		allFlags[fl.Name] = true
	})
	for name, visited := range allFlags {
		if !visited {
			unVisitedFlags = append(unVisitedFlags, f.FlagSet.Lookup(name))
		}
	}

	for _, fl := range unVisitedFlags {
		// No value read from flag
		var strVal *string

		if fileVal, ok := f.ConfigFileStringValue(fl.Name); ok {
			strVal = &fileVal
		}
		// Find the Env value
		envVal := ""
		// Blank envName means skip ENV lookup, for safety
		if envName, ok := f.envNames[fl.Name]; ok {
			envVal = os.Getenv(envName)
			if len(envVal) != 0 {
				strVal = &envVal
			}
		}

		if strVal != nil {
			// Val never set, fallback to default
			*f.stringValues[fl.Name] = *strVal
		}
	}
}

func String( name, defaultValue, envName, usage string ) *string {
	return CommandLine.String(name, defaultValue, envName, usage)
}

func (f *FlagfigSet)String( name, defaultValue, envName, usage string ) *string {
	if _, ok := f.stringValues[name]; !ok {
		f.stringValues[name] = new(string)
		f.envNames[name] = envName
	}
	// Finally, nothing else is left, use the
	f.FlagSet.StringVar(f.stringValues[name], name, defaultValue, usage)
	return f.stringValues[name]
}

// readConfigurationFiles in order and records the values, overriding each in turn
// Files are read just once and only the final value is stored
func (f *FlagfigSet)readConfigurationFiles() {
	if f.fileConfigurations == nil {
		f.fileConfigurations = make(map[string]string)

		for _, filePath := range f.configFilePaths {
			if filePath != nil && len(*filePath) != 0 {
				dat, err := ioutil.ReadFile(*filePath)
				if err != nil {
					panic(err)
				}
				var jsonDat map[string]interface{}
				err = json.Unmarshal(dat, &jsonDat)
				if err != nil {
					// Skip this file
					log.Printf("Unable to JSON Decode file: '%s' because: %s", *filePath, err)
				} else {
					// Process file's contents
					for key, val := range jsonDat {
						switch v := val.(type) {
						case string:
							f.fileConfigurations[key] = v
						default:
							log.Printf("Unsupported type %T for configuration named %s in '%s'; skipping", v, key, *filePath)
						}
					}
				}
			}
		}
	}
}

// ConfigFileStringValue gets the string value for the flagname name
func (f *FlagfigSet)ConfigFileStringValue( name string ) (val string, ok bool) {
	f.readConfigurationFiles()
	val, ok = f.fileConfigurations[name]
	return
}