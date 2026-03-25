package shared

import "regexp"

// AddrRegex matches a 64-character hexadecimal string (SHA-256)
var AddrRegex = regexp.MustCompile("^[0-9a-fA-F]{64}$")