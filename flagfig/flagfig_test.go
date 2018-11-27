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

package flagfig

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestUsage(t *testing.T) {
	called := false
	ResetForTesting(func() { called = true })
	if flag.CommandLine.Parse([]string{"-x"}) == nil {
		t.Error("parse did not fail for unknown flag")
	}
	if !called {
		t.Error("did not call Usage for unknown flag")
	}
}

func testParse(f *FlagfigSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	stringFlag := f.String("string", "0", "ENV_NAME", "string value")
	extra := "one-extra-argument"
	args := []string{
		"-string", "hello",
		extra,
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if len(f.Args()) != 1 {
		t.Error("expected one argument, got", len(f.Args()))
	} else if f.Args()[0] != extra {
		t.Errorf("expected argument %q got %q", extra, f.Args()[0])
	}
}

func TestParse(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParse(CommandLine, t)
}

func TestFlagSetParse(t *testing.T) {
	testParse(NewFlagfigSet("test", flag.ContinueOnError), t)
}

func testParseWithEnv(f *FlagfigSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	_ = os.Setenv("ENV_STRING", "hello")
	defer func() { _ = os.Setenv("ENV_STRING", "") }()
	stringFlag := f.String("string", "0", "ENV_STRING", "string value")
	args := make([]string, 0, 0)
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *stringFlag != "hello" {
		t.Error("string environment should be `hello`, is ", *stringFlag)
	}
}

func TestParseWithEnv(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParseWithEnv(CommandLine, t)
}

func testParseOverwriteEnv(f *FlagfigSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	_ = os.Setenv("ENV_STRING", "nothello")
	defer func() { _ = os.Setenv("ENV_STRING", "") }()
	stringFlag := f.String("string", "0", "ENV_STRING", "string value")
	args := []string{
		"-string=hello",
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *stringFlag != "hello" {
		t.Error("string environment should be `hello`, is ", *stringFlag)
	}

}

func TestParseOverwriteEnv(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParseOverwriteEnv(CommandLine, t)
}

func testParseOverwriteFile(f *FlagfigSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}

	fileContent := &struct {
		BoolConfigTrue  bool
		BoolConfigFalse bool
		FloatConfig     float64
		IntConfig       int
		InConfig        string
		Inenv           string
		Inflag          string
		Int64Config     int64
		UintConfig      uint
		Uint64Config    uint64
		DurationConfig  time.Duration
	}{
		BoolConfigTrue:  true,
		BoolConfigFalse: false,
		FloatConfig:     987.654321,
		IntConfig:       1234,
		InConfig:        "configfile",
		Inenv:           "configfile",
		Inflag:          "configfile",
		Int64Config:     234,
		UintConfig:      345,
		Uint64Config:    456,
		DurationConfig:  30 * time.Second,
	}
	marshalled, _ := json.Marshal(fileContent)
	tmpFileName, tfremove := testTempFile(t)
	defer tfremove()
	err := ioutil.WriteFile(tmpFileName, marshalled, 0600)
	if err != nil {
		t.Fatalf("Unable to write temp file: %s", tmpFileName)
	}

	err = os.Setenv("ENV_ENV", "env")
	if err != nil {
		t.Fatal("Unable to edit os.Setenv: ENV_STRING")
	}

	f.AddConfigFile("config-file", "Config file of doom")
	boolConfigTrue := f.Bool("BoolConfigTrue", false, "", "bool value true")
	boolConfigFalse := f.Bool("BoolConfigFalse", true, "", "bool value false")
	stringConfig := f.String("InConfig", "0", "", "string value")
	intConfig := f.Int("IntConfig", 1, "", "int value")
	floatConfig := f.Float64("FloatConfig", 0.5, "", "float value")
	stringEnv := f.String("Inenv", "0", "ENV_ENV", "string value")
	stringFlag := f.String("Inflag", "0", "ENV_STRING", "string value")
	int64Config := f.Int64("Int64Config", 0, "", "string value")
	uintConfig := f.Uint("UintConfig", 0, "", "string value")
	uint64Config := f.Uint64("Uint64Config", 0, "", "string value")
	durationConfig := f.Duration("DurationConfig", 1, "", "string value")
	args := []string{
		"-config-file=" + tmpFileName,
		"-Inflag=flag",
	}
	if err = f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *stringFlag != "flag" {
		t.Error("stringFlag flag should be `flag`, is ", *stringFlag)
	}
	if *stringEnv != "env" {
		t.Error("stringEnv environment should be `env`, is ", *stringEnv)
	}
	if *stringConfig != "configfile" {
		t.Error("stringConfig flag should be `configfile`, is ", *stringConfig)
	}
	if *intConfig != fileContent.IntConfig {
		t.Error("intConfig flag should be 1234, is ", *intConfig)
	}
	if *floatConfig != fileContent.FloatConfig {
		t.Error("floatConfig flag should be 987.654321, is ", *floatConfig)
	}
	if *boolConfigTrue != fileContent.BoolConfigTrue {
		t.Error("boolConfigTrue flag should be true, is ", *boolConfigTrue)
	}
	if *boolConfigFalse != fileContent.BoolConfigFalse {
		t.Error("boolConfigFalse flag should be false, is ", *boolConfigFalse)
	}
	if *int64Config != fileContent.Int64Config {
		t.Error("int64Config flag should be ", fileContent.Int64Config, " is ", *int64Config)
	}
	if *uintConfig != fileContent.UintConfig {
		t.Error("uintConfig flag should be ", fileContent.UintConfig, " is ", *uintConfig)
	}
	if *uint64Config != fileContent.Uint64Config {
		t.Error("uint64Config flag should be ", fileContent.Uint64Config, " is ", *uint64Config)
	}
	if *durationConfig != fileContent.DurationConfig {
		t.Error("durationConfig flag should be ", fileContent.DurationConfig, " is ", *durationConfig)
	}
}

func TestParseOverwriteFile(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParseOverwriteFile(CommandLine, t)
}
