//go:build debug
// +build debug

package util

import (
	"free5gc/lib/path_util"
)

var NefLogPath = path_util.Gofree5gcPath("NEF-service/nefsslkey.log")
var NefPemPath = path_util.Gofree5gcPath("free5gc/support/TLS/_debug.pem")
var NefKeyPath = path_util.Gofree5gcPath("free5gc/support/TLS/_debug.key")
var DefaultNefConfigPath = path_util.Gofree5gcPath("free5gc/config/nefcfg.yaml")
