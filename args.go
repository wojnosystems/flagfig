package flagfig

import "strings"

// ArgsAfterArgWithEqualTo given a token to split on, such as: "--", will take in the optional set of arguments following this token and trim off any arguments *prior* to and including the token argument. This was created to facilitate testing, as this will strip out the set-up arguments for a test harness/debugging call.
// If the token was not found, return the same list of arguments passed to this function
// For example, if you
func ArgsAfterArgWithEqualTo(token string, args ...string) []string {
	dex := 0
	found := false
	for i, arg := range args {
		if strings.Compare(arg, token) == 0 {
			dex = i
			found = true
			break
		}
	}
	if found {
		return args[dex+1:]
	} else {
		return args
	}
}
