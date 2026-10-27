package config

import "testing"

func TestParsePortalConf(t *testing.T) {
	if err := Parse("../portal.conf"); err != nil {
		t.Fatalf("Parse portal.conf failed: %v", err)
	}
	if *RadiusEnable != true {
		t.Errorf("RadiusEnable = %v, want true", *RadiusEnable)
	}
	if *RadiusAuthPort != 1812 {
		t.Errorf("RadiusAuthPort = %d, want 1812", *RadiusAuthPort)
	}
	if *RadiusSecret != "huawei123" {
		t.Errorf("RadiusSecret = %q, want huawei123", *RadiusSecret)
	}
	if *HuaweiVersion != 2 {
		t.Errorf("HuaweiVersion = %d, want 2", *HuaweiVersion)
	}
	if *HuaweiDomain != "dongjianwei.cn" {
		t.Errorf("HuaweiDomain = %q, want dongjianwei.cn", *HuaweiDomain)
	}
	if *UseRemoteIpAsUserIp != true {
		t.Errorf("UseRemoteIpAsUserIp = %v, want true", *UseRemoteIpAsUserIp)
	}
	if *NasIp != "10.10.100.253" {
		t.Errorf("NasIp = %q, want 10.10.100.253", *NasIp)
	}
	if *MacCacheTimeout != 3600 {
		t.Errorf("MacCacheTimeout = %d, want 3600", *MacCacheTimeout)
	}
	if *HttpPort != 8081 {
		t.Errorf("HttpPort = %d, want 8081", *HttpPort)
	}
}

func TestParseDefaults(t *testing.T) {
	// no file should keep defaults
	if *LoginPage != "./login.html" {
		t.Errorf("LoginPage default = %q", *LoginPage)
	}
	if *HuaweiPort != 50100 {
		t.Errorf("HuaweiPort default = %d", *HuaweiPort)
	}
}
