package component

import (
	"io/ioutil"
	"os"
	"strings"
	"time"
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/17331149294/portal/config"
	"github.com/17331149294/portal/i"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
)

// 新增MAC缓存结构体
type MacCache struct {
	Mac       string
	ExpiresAt time.Time // 过期时间
	Username  string  // 新增：用户名
	Ip        string  // 新增：IP地址
}

type AuthInfo struct {
	Name    []byte
	Pwd     []byte
	Mac     net.HardwareAddr
	Timeout uint32
}

type AuthServer struct {
	authing_user map[string]*AuthInfo
	macCache     map[string]MacCache // 新增MAC缓存 map[mac地址]缓存信息
}


// 定时清理过期的MAC缓存（每小时执行一次）
func (a *AuthServer) cleanExpiredMacCache() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		for mac, cache := range a.macCache {
			if now.After(cache.ExpiresAt) {
				delete(a.macCache, mac)
			}
		}
		// 同步更新到文件
		a.saveMacCacheToFile()
	}
}

// 从文件加载MAC缓存（改为导出方法）
func (a *AuthServer) LoadMacCacheFromFile() {
    filePath := "mac.txt"
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        if !os.IsNotExist(err) {
            log.Printf("加载MAC缓存文件失败: %v", err)
        }
        return
    }

    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        parts := strings.Split(line, ",")
        if len(parts) != 4 {  // 适配新格式
            continue
        }
        mac := parts[0]
        expireTime, err := time.Parse(time.RFC3339, parts[1])
        if err != nil {
            continue
        }
        username := parts[2]
        ip := parts[3]
        if time.Now().Before(expireTime) {
            a.macCache[mac] = MacCache{
                Mac:       mac,
                ExpiresAt: expireTime,
                Username:  username,
                Ip:        ip,
            }
        }
    }
}


// 将MAC缓存保存到文件
func (a *AuthServer) saveMacCacheToFile() {
    filePath := "mac.txt"
    var content strings.Builder
    now := time.Now()
    for _, cache := range a.macCache {
        if now.Before(cache.ExpiresAt) {
            // 格式：mac,过期时间,用户名,IP
            content.WriteString(fmt.Sprintf("%s,%s,%s,%s\n",
                cache.Mac,
                cache.ExpiresAt.Format(time.RFC3339),
                cache.Username,
                cache.Ip))
        }
    }
    err := ioutil.WriteFile(filePath, []byte(content.String()), 0644)
    if err != nil {
        log.Printf("保存MAC缓存文件失败: %v", err)
    }
}
// 新增：添加MAC到缓存（有效期60分钟）
func (a *AuthServer) addMacToCache(mac string, username string, ip string) {
    // 从配置读取有效期（秒），转换为时间间隔
    timeout := time.Duration(*config.MacCacheTimeout) * time.Second
    expiresAt := time.Now().Add(timeout)
    a.macCache[mac] = MacCache{
        Mac:       mac,
        ExpiresAt: expiresAt,
        Username:  username,  // 存储用户名
        Ip:        ip,        // 存储IP
    }
    a.saveMacCacheToFile() // 同步到文件
}

var BASIC_SERVICE = new(AuthServer)

func InitBasic() {
	BASIC_SERVICE.authing_user = make(map[string]*AuthInfo)
	BASIC_SERVICE.macCache = make(map[string]MacCache)
	// 启动定时清理过期MAC缓存的协程
	go BASIC_SERVICE.cleanExpiredMacCache()
}

func (a *AuthServer) AuthChap(username []byte, chapid byte, chappwd, chapcha []byte, userip net.IP, usermac net.HardwareAddr) (err error, to uint32) {
	if info, ok := a.authing_user[userip.String()]; ok {
		if bytes.Compare(username, info.Name) == 0 && i.TestChapPwd(chapid, info.Pwd, chapcha, chappwd) {
			to = info.Timeout
			info.Mac = usermac
			return
		}
	} else {
		err = fmt.Errorf("radius auth - no such user ", userip.String())
	}
	return
}

func (a *AuthServer) AuthMac(mac net.HardwareAddr, userip net.IP) (error, uint32) {
    macStr := mac.String()
    // 每次都从文件加载最新数据
    a.macCache = make(map[string]MacCache)  // 清空当前缓存
    a.LoadMacCacheFromFile()                // 重新从文件加载

    if cache, exists := a.macCache[macStr]; exists {
        if time.Now().Before(cache.ExpiresAt) {
            // 输出认证成功信息
            log.Printf("mac认证成功，mac账号为%s，上次认证成功的用户名为%s，IP为%s",
                cache.Mac, cache.Username, cache.Ip)
            return nil, 3600
        } else {
            // 已过期，从文件中删除
            delete(a.macCache, macStr)
            a.saveMacCacheToFile()
        }
    }
    return fmt.Errorf("MAC认证失败: %s 不在缓存中或已过期", macStr), 0
}

func (a *AuthServer) AuthPap(username, userpwd []byte, userip net.IP) (err error, to uint32) {
	if info, ok := a.authing_user[userip.String()]; ok {
		if bytes.Compare(info.Pwd, userpwd) == 0 {
			to = info.Timeout
		}
	} else {
		err = fmt.Errorf("radius auth - no such user ", userip.String())
	}
	return
}

func (a *AuthServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var err error
	if config.IsValidClient(r.RemoteAddr) {
		timeout := r.FormValue("timeout")
		nas := r.FormValue("nasip")
		if *config.NasIp != "" {
			nas = *config.NasIp
		}
		userip_str := r.FormValue("userip")
		username := []byte(r.FormValue("username"))
		userpwd := []byte(r.FormValue("userpwd"))
		var to uint64
		to, err = strconv.ParseUint(timeout, 10, 32)

		if to == 0 && *config.DefaultTimeout != 0 {
			to = *config.DefaultTimeout
		}

		userip := net.ParseIP(userip_str)
		if *config.UseRemoteIpAsUserIp == true {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			userip = net.ParseIP(ip)
		} else if userip == nil {
			u_refer := r.Header.Get("Referer")
			var u *url.URL
			if u, err = url.Parse(u_refer); err == nil {
				if uip := u.Query().Get("userip"); uip != "" {
					userip = net.ParseIP(userip_str)
				} else {
					err = fmt.Errorf("请求解析Referer错误")
				}
			} else {
				err = fmt.Errorf("配置错误！请联系管理员")
			}
		}
		var full_username []byte
		if userip != nil {
			if basip := net.ParseIP(nas); basip != nil {
				log.Printf("got a login request from %s o1n nas %s\n", userip, basip)
				if len(username) == 0 {
					if *config.RandomUser {
						full_username, userpwd = a.RandomUser(userip, basip, *config.HuaweiDomain, uint32(to))
					} else {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
				} else {
					full_username = []byte(string(username) + "@" + *config.HuaweiDomain)
					a.authing_user[userip.String()] = &AuthInfo{username, userpwd, []byte{}, uint32(to)}
				}
				if err = Auth(userip, basip, uint32(to), []byte(full_username), userpwd); err == nil {
					w.Write([]byte(a.authing_user[userip.String()].Mac.String()))
					message := "用户已认证成功: " + string(username)
					w.Write([]byte(message))
					usernameStr := string(username)
					log.Printf("%s 已经认证成功 %s on nas %s\n", usernameStr, userip, basip)
					// 认证成功后，获取MAC并添加到缓存
					macStr := a.authing_user[userip.String()].Mac.String()
					if macStr != "" {
						a.addMacToCache(macStr, string(username), userip.String()) // 添加到缓存
					}
					w.Write([]byte(macStr))

					return
				}
			} else {
				err = fmt.Errorf("NAS IP配置错误")
			}
		}
	} else {
		err = fmt.Errorf("该IP不在配置可允许的用户中")
	}
	if err != nil {
		log.Println("login error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}
}

//处理Logout请求
func (a *AuthServer) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var err error
	nas := r.FormValue("nasip") //TODO
	userip_str := r.FormValue("userip")
	if userip := net.ParseIP(userip_str); userip != nil {
		if basip := net.ParseIP(nas); basip != nil {
			if _, err = Logout(userip, *config.HuaweiSecret, basip); err == nil {
				w.WriteHeader(http.StatusOK)
				message := "用户已成功退出"
				w.Write([]byte(message))
				log.Printf("已经退出 %s on nas %s\n", userip, basip)
				return
			}
		} else {
			err = fmt.Errorf("Parse Ip err from %s", nas)
		}
	} else {
		err = fmt.Errorf("Parse Ip err from %s", userip_str)
	}
}

func (a *AuthServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("Show login page")
	path := filepath.FromSlash(*config.LoginPage)
	http.ServeFile(w, r, path)
}

func (a *AuthServer) RandomUser(userip, nasip net.IP, domain string, timeout uint32) ([]byte, []byte) {
	hash := md5.New()
	hash.Write(userip)
	hash.Write(nasip)
	bts := hash.Sum(nil)
	username := []byte(userip.String())
	app := []byte("@" + domain)
	if len(username)+len(app) > 32 {
		username = username[:32-len(app)]
	}
	fname := append(username, app...)
	userpwd := bts
	a.authing_user[userip.String()] = &AuthInfo{username, userpwd, []byte{}, timeout}
	return fname, userpwd
}

func (a *AuthServer) AcctStart(username []byte, userip net.IP, nasip net.IP, usermac net.HardwareAddr, sessionid string) error {
	return nil
}

func (a *AuthServer) AcctStop(username []byte, userip net.IP, nasip net.IP, usermac net.HardwareAddr, sessionid string) error {
	callBackOffline(*config.CallBackUrl, userip, nasip)
	return nil
}

func (a *AuthServer) NotifyLogout(userip, nasip net.IP) error {
	callBackOffline(*config.CallBackUrl, userip, nasip)
	return nil
}
