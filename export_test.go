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