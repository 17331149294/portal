package main

import (
	"flag"
	"fmt"
	"github.com/17331149294/portal/component"
	"github.com/17331149294/portal/config"
	"path/filepath"
)

func main() {
	//http server
	//radius avp server
	//radius accounting server

	// go func() {
	// 	res, err := v1.Challenge(net.IPv4(192, 168, 10, 254), *config.PortalSecret, net.IPv4(192, 168, 56, 2))
	// }()

	path := flag.String("config", "./portal.conf", "设置配置文件的路径")
	component.InitBasic()
	flag.Parse()
	*path = filepath.FromSlash(*path)
	if err := config.Parse(*path); err == nil {
		if config.IsValid() {
			component.InitLogger()
			// 调用导出方法加载MAC缓存
			component.BASIC_SERVICE.LoadMacCacheFromFile()
			go component.StartHuawei()
			if *config.RadiusEnable {
				go component.StartRadiusAuth()
				go component.StartRadiusAcc()
			}
			component.StartHttp()
		}
	} else {
		fmt.Println(err)
	}
}
