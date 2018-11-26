// Copyright 2018 Chris Wojno. All rights reserved.

package flagfig

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"testing"
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
	defer func() {_ = os.Setenv("ENV_STRING", "")}()
	stringFlag := f.String("string", "0", "ENV_STRING", "string value")
	args := make([]string,0,0)
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
	defer func() {_ = os.Setenv("ENV_STRING", "")}()
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

	tmpFileName, tfremove := testTempFile(t)
	defer tfremove()
	fileContent := make(map[string]string)
	fileContent["inconfig"] = "configfile"
	fileContent["inenv"] = "configfile"
	fileContent["inflag"] = "configfile"
	marshalled, _ := json.Marshal(fileContent)
	err := ioutil.WriteFile(tmpFileName, marshalled, 0600 )
	if err != nil {
		t.Fatalf("Unable to write temp file: %s", tmpFileName)
	}

	err = os.Setenv("ENV_ENV", "env")
	if err != nil {
		t.Fatal("Unable to edit os.Setenv: ENV_STRING")
	}

	f.AddConfigFile("config-file","Config file of doom")
	stringConfig := f.String("inconfig", "0", "", "string value")
	stringEnv := f.String("inenv", "0", "ENV_ENV", "string value")
	stringFlag := f.String("inflag", "0", "ENV_STRING", "string value")
	args := []string{
		"-config-file=" + tmpFileName,
		"-inflag=flag",
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
}

func TestParseOverwriteFile(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParseOverwriteFile(CommandLine, t)
}