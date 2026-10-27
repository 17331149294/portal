package config

import (
	"net"
	"os"
	"strconv"
	"strings"
)

// valueKind describes the typed storage of a registered config variable.
type valueKind int

const (
	vString valueKind = iota
	vBool
	vInt
	vInt64
	vUint64
)

type varEntry struct {
	kind valueKind
	ptr  interface{}
}

// registry maps "section.key" -> how to set the pointed-to variable.
var registry = map[string]varEntry{}

// Bool registers a bool config variable with a default value and returns its pointer.
func Bool(name string, def bool) *bool {
	p := new(bool)
	*p = def
	registry[name] = varEntry{vBool, p}
	return p
}

// Int registers an int config variable with a default value and returns its pointer.
func Int(name string, def int) *int {
	p := new(int)
	*p = def
	registry[name] = varEntry{vInt, p}
	return p
}

// Int64 registers an int64 config variable with a default value and returns its pointer.
func Int64(name string, def int64) *int64 {
	p := new(int64)
	*p = def
	registry[name] = varEntry{vInt64, p}
	return p
}

// Uint64 registers a uint64 config variable with a default value and returns its pointer.
func Uint64(name string, def uint64) *uint64 {
	p := new(uint64)
	*p = def
	registry[name] = varEntry{vUint64, p}
	return p
}

// String registers a string config variable with a default value and returns its pointer.
func String(name string, def string) *string {
	p := new(string)
	*p = def
	registry[name] = varEntry{vString, p}
	return p
}

// Parse reads a TOML config file and applies its values to the registered variables.
func Parse(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	kv, err := parseTOML(data)
	if err != nil {
		return err
	}
	for name, e := range registry {
		raw, ok := kv[name]
		if !ok {
			continue
		}
		switch e.kind {
		case vString:
			*(e.ptr.(*string)) = unquote(raw)
		case vBool:
			if v, err := strconv.ParseBool(raw); err == nil {
				*(e.ptr.(*bool)) = v
			}
		case vInt:
			if v, err := strconv.Atoi(raw); err == nil {
				*(e.ptr.(*int)) = v
			}
		case vInt64:
			if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
				*(e.ptr.(*int64)) = v
			}
		case vUint64:
			if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
				*(e.ptr.(*uint64)) = v
			}
		}
	}
	return nil
}

var (
	RadiusEnable        = Bool("radius.enabled", false)
	RadiusAuthPort      = Int("radius.port", 1812)
	RadiusAccPort       = Int("radius.acc_port", 1813)
	RadiusSecret        = String("radius.secret", "huawei123")
	HttpPort            = Int("http.port", 8080)
	HttpWhiteList       = String("http.white_list", "")
	RandomUser          = Bool("basic.random_user", true)
	LogFile             = String("basic.logfile", "")
	CallBackUrl         = String("basic.callback_logout", "")
	UseRemoteIpAsUserIp = Bool("basic.remote_ip_as_user_ip", false)
	HuaweiPort          = Int("huawei.port", 50100)
	HuaweiVersion       = Int("huawei.version", 1)
	HuaweiTimeout       = Int("huawei.timeout", 15)
	HuaweiSecret        = String("huawei.secret", "testing123")
	HuaweiNasPort       = Int("huawei.nas_port", 2000)
	HuaweiDomain        = String("huawei.domain", "huawei.com")
	LoginPage           = String("basic.login_page", "./login.html")
	NasIp               = String("basic.nas_ip", "")
	DefaultTimeout      = Uint64("basic.default_timeout", 0)
	AuthType            = String("radius.auth_type", "random")
	MacCacheTimeout     = Int64("basic.mac_cache_timeout", 3600)
)

func IsValid() bool {
	return true
}

func IsValidClient(addr string) bool {
	if *HttpWhiteList == "" {
		return true
	}
	if ip, _, err := net.SplitHostPort(addr); err == nil {
		if strings.Contains(*HttpWhiteList, ip) {
			return true
		}
	}
	return false
}
