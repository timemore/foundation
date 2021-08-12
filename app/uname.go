package app

import (
	"bytes"
	"fmt"
	"net"
	"runtime"
	"strings"
	"syscall"
)

func unameString() (string, error) {
	os := runtime.GOOS
	switch os {
	case "windows", "darwin":
		return getNetworkHardwareName(), nil
	}

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

func getNetworkHardwareName() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	var currentIP, currentNetworkHardwareName string

	for _, address := range addrs {

		// check the address type and if it is not a loopback the display it
		// = GET LOCAL IP ADDRESS
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentIP = ipnet.IP.String()
			}
		}
	}
	// get all the system's or local machine's network interfaces
	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {

		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				// only interested in the name with current IP address
				if strings.Contains(addr.String(), currentIP) {
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}

	// extract the hardware information base on the interface name
	// capture above
	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)

	if err != nil {
		return ""
	}

	macAddress := netInterface.HardwareAddr

	// verify if the MAC address can be parsed properly
	hwAddr, err := net.ParseMAC(macAddress.String())

	if err != nil {
		return ""
	}
	return hwAddr.String()
}
