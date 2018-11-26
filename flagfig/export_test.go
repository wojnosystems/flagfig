// Copyright 2018 Chris Wojno. All rights reserved.

package flagfig

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

// ResetForTesting clears all flag state and sets the usage function as directed.
// After calling ResetForTesting, parse errors in flag handling will not
// exit the program.
func ResetForTesting(usage func()) {
	CommandLine = NewFlagfigSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.Usage = usage
	flag.Usage = usage
}

func testTempFile(t *testing.T) (string,func()) {
	tf, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	return tf.Name(), func() { os.Remove(tf.Name())}
}