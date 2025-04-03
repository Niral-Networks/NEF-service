//go:build !debug
// +build !debug

package util

import (
	"github.com/free5gc/path_util"
)

var NefLogPath = path_util.Free5gcPath("NEF-service/nefsslkey.log")
var NefPemPath = path_util.Free5gcPath("free5gc/support/TLS/nef.pem")
var NefKeyPath = path_util.Free5gcPath("free5gc/support/TLS/nef.key")

//var DefaultNefConfigPath = path_util.Free5gcPath("free5gc/config/nefcfg.yaml")

var (
	DefaultNefConfigPath = "./etc/niralos/nefcfg.yaml"
)
