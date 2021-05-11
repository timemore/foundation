package app

import (
	"bytes"
	"fmt"
	"syscall"
)

func unameString() (string, error) {
	utsname := syscall.Utsname{}
	err := syscall.Uname(&utsname)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s %s %s %s",
		utsStringToString(utsname.Sysname),
		utsStringToString(utsname.Nodename),
		utsStringToString(utsname.Release),
		utsStringToString(utsname.Version),
		utsStringToString(utsname.Machine),
	), nil
}

func utsStringToString(utsStr [65]int8) string {
	s := make([]byte, len(utsStr))
	i := 0
	for _, c := range utsStr {
		s[i] = byte(c)
		i++
	}
	s = s[:bytes.IndexByte(s, 0)]
	return string(s)
}
