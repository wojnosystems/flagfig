/*
Copyright 2018 Chris Wojno.
Attribution 4.0 International (CC BY 4.0)
All rights reserved.
You do not have to comply with the license for elements of the material in the public domain or where your use is
permitted by an applicable exception or limitation.

No warranties are given. The license may not give you all of the permissions necessary for your intended use.
For example, other rights such as publicity, privacy, or moral rights may limit how you use the material.

See LICENSE file for the full license
*/

/*
	Package flagfig implements configuration file, environment, and command-line flag parsing

	Usage

	Define flags using:
		flagfig.String(), flagfig.Int(), flagfig.Float64(), flagfig.Duration(), flagfig.Uint(), flagfig.Uint64()
		flagfig.Bool(),flagfig.Int64()
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
	and only work with string keys and values: string, float. So you can use this:

	{
		flag1: "value",
        flag2: "anothervalue",
		flag3: 1234
        flag4: 1234.56
        duration: 10000000000
	}

	But you cannot use a file like this:

	{
		flag1: {
			complexItem: 4
		}
	}


	Hack Alert

	This package is an extreme hack of the GoLang flag package. I tried to re-use as much as possible, but without
    the exported data values, I had to get creative with the time conversions.


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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var CommandLine = NewFlagfigSet(os.Args[0], flag.ExitOnError)

const (
	intType = iota
	stringType
	boolType
	floatType
	int64Type
	uintType
	uint64Type
	durationType
)

// FlagurationSet
type FlagfigSet struct {
	flag.FlagSet
	configFilePaths []*string
	flagTypes       map[string]int
	envNames        map[string]string
}

func NewFlagfigSet(name string, errorHandling flag.ErrorHandling) *FlagfigSet {
	fs := &FlagfigSet{}
	fs.FlagSet = *flag.NewFlagSet(name, errorHandling)
	fs.configFilePaths = make([]*string, 0, 1)
	fs.envNames = make(map[string]string)
	fs.flagTypes = make(map[string]int)
	return fs
}

func Parse() {
	_ = CommandLine.Parse(os.Args[1:])
}

func (f *FlagfigSet) Parse(arguments []string) (err error) {
	err = f.FlagSet.Parse(arguments)
	if err == nil {
		err = f.Collate()
	}
	return
}

// AddConfigFile adds a configuration file flag to the command line
// When Parse() is called, this file will be added to the list of files to parse when looking for configuration values
// name is the flagname
func AddConfigFile(name, usage string) *string {
	return CommandLine.AddConfigFile(name, usage)
}
func (f *FlagfigSet) AddConfigFile(name, usage string) *string {
	p := new(string)
	f.configFilePaths = append(f.configFilePaths, p)
	f.FlagSet.StringVar(p, name, "", usage)
	return p
}

// Collate combines the values from config files, environment variables, and flags as a single value.
// Assumes that the command flags are already parsed
func (f *FlagfigSet) Collate() (err error) {
	unVisitedFlags := make(map[string]*flag.Flag)
	allFlags := make(map[string]bool)
	f.FlagSet.VisitAll(func(fl *flag.Flag) {
		allFlags[fl.Name] = false
	})
	f.FlagSet.Visit(func(fl *flag.Flag) {
		allFlags[fl.Name] = true
	})
	for name, visited := range allFlags {
		if !visited {
			unVisitedFlags[name] = f.FlagSet.Lookup(name)
		}
	}

	err = f.readConfigurationFiles(unVisitedFlags)
	if err != nil {
		return
	}

	for _, fl := range unVisitedFlags {
		// Find the Env value
		envVal := ""
		// Blank envName means skip ENV lookup, for safety
		if envName, ok := f.envNames[fl.Name]; ok {
			envVal = os.Getenv(envName)
			if len(envVal) != 0 {
				err = f.FlagSet.Set(fl.Name, envVal)
			}
		}

	}
	return
}

func Bool(name string, defaultValue bool, envName, usage string) *bool {
	return CommandLine.Bool(name, defaultValue, envName, usage)
}

func (f *FlagfigSet) Bool(name string, defaultValue bool, envName, usage string) *bool {
	p := new(bool)
	f.envNames[name] = envName
	f.flagTypes[name] = boolType
	f.FlagSet.BoolVar(p, name, defaultValue, usage)
	return p
}

func String(name, defaultValue, envName, usage string) *string {
	return CommandLine.String(name, defaultValue, envName, usage)
}

func (f *FlagfigSet) String(name, defaultValue, envName, usage string) *string {
	p := new(string)
	f.envNames[name] = envName
	f.flagTypes[name] = stringType
	f.FlagSet.StringVar(p, name, defaultValue, usage)
	return p
}

func Int(name string, defaultValue int, envName, usage string) *int {
	return CommandLine.Int(name, defaultValue, envName, usage)
}
func (f *FlagfigSet) Int(name string, defaultValue int, envName, usage string) *int {
	p := new(int)
	f.envNames[name] = envName
	f.flagTypes[name] = intType
	f.FlagSet.IntVar(p, name, defaultValue, usage)
	return p
}

func Float64(name string, defaultValue float64, envName, usage string) *float64 {
	return CommandLine.Float64(name, defaultValue, envName, usage)
}
func (f *FlagfigSet) Float64(name string, defaultValue float64, envName, usage string) *float64 {
	p := new(float64)
	f.envNames[name] = envName
	f.flagTypes[name] = floatType
	f.FlagSet.Float64Var(p, name, defaultValue, usage)
	return p
}

func Int64(name string, defaultValue int64, envName, usage string) *int64 {
	return CommandLine.Int64(name, defaultValue, envName, usage)
}

func (f *FlagfigSet) Int64(name string, defaultValue int64, envName, usage string) *int64 {
	p := new(int64)
	f.envNames[name] = envName
	f.flagTypes[name] = int64Type
	f.FlagSet.Int64Var(p, name, defaultValue, usage)
	return p
}

func Uint(name string, defaultValue uint, envName, usage string) *uint {
	return CommandLine.Uint(name, defaultValue, envName, usage)
}

func (f *FlagfigSet) Uint(name string, defaultValue uint, envName, usage string) *uint {
	p := new(uint)
	f.envNames[name] = envName
	f.flagTypes[name] = uintType
	f.FlagSet.UintVar(p, name, defaultValue, usage)
	return p
}

func Uint64(name string, defaultValue uint64, envName, usage string) *uint64 {
	return CommandLine.Uint64(name, defaultValue, envName, usage)
}

func (f *FlagfigSet) Uint64(name string, defaultValue uint64, envName, usage string) *uint64 {
	p := new(uint64)
	f.envNames[name] = envName
	f.flagTypes[name] = uint64Type
	f.FlagSet.Uint64Var(p, name, defaultValue, usage)
	return p
}

func Duration(name string, defaultValue time.Duration, envName, usage string) *time.Duration {
	return CommandLine.Duration(name, defaultValue, envName, usage)
}

func (f *FlagfigSet) Duration(name string, defaultValue time.Duration, envName, usage string) *time.Duration {
	p := new(time.Duration)
	f.envNames[name] = envName
	f.flagTypes[name] = durationType
	f.FlagSet.DurationVar(p, name, defaultValue, usage)
	return p
}

// readConfigurationFiles in order and records the values, overriding each in turn
// Files are read just once and only the final value is stored
func (f *FlagfigSet) readConfigurationFiles(unvisitedFlags map[string]*flag.Flag) (err error) {
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
					if _, ok := unvisitedFlags[key]; ok {
						switch v := val.(type) {
						case bool:
							if v {
								_ = f.FlagSet.Set(key, "true")
							} else {
								_ = f.FlagSet.Set(key, "false")
							}
						case string:
							_ = f.FlagSet.Set(key, v)
						case int:
							_ = f.FlagSet.Set(key, strconv.Itoa(v))
						case int64:
							_ = f.FlagSet.Set(key, strconv.FormatInt(v, 10))
						case uint:
							_ = f.FlagSet.Set(key, strconv.FormatUint(uint64(v), 10))
						case uint64:
							_ = f.FlagSet.Set(key, strconv.FormatUint(v, 10))
						case float64:
							// So, every number in JSON is actually a float64...
							switch f.flagTypes[key] {
							case intType:
								_ = f.FlagSet.Set(key, fmt.Sprintf("%.0f", v))
							case uintType:
								_ = f.FlagSet.Set(key, fmt.Sprintf("%.0f", v))
							case int64Type:
								_ = f.FlagSet.Set(key, fmt.Sprintf("%.0f", v))
							case uint64Type:
								_ = f.FlagSet.Set(key, fmt.Sprintf("%.0f", v))
							case floatType:
								_ = f.FlagSet.Set(key, fmt.Sprintf("%f", v))
							case durationType:
								s := strings.TrimSpace(fmt.Sprintf("%18.0fns", v))
								//fmt.Println(key, ":",s)
								_ = f.FlagSet.Set(key, s)
							}
						default:
							log.Fatalf("Unsupported Config file type %t", v)
						}
					}
				}
			}
		}
	}
	return
}
