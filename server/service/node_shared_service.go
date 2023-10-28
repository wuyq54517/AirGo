package service

import (
	"AirGo/global"
	"AirGo/model"
	"AirGo/utils/encrypt_plugin"
	"AirGo/utils/net_plugin"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ParseVMessLink(link string) *model.NodeShared {
	// vmess 例子 {"add":"AirGo","aid":"0","alpn":"h2,http/1.1","fp":"qq","host":"www.baidu.com","id":"e0d5fe65-a5d1-4b8a-8d40-ed92a6a35d8b","net":"ws","path":"/path","port":"6666","ps":"到期时间:2024-03-06  |  剩余流量:20.00GB","scy":"auto","sni":"www.baidu.com","tls":"tls","type":"","v":"2"}
	// vmess 例子 {"add":"AirGo","aid":"0","alpn":"","fp":"","host":"www.baidu.com","id":"e0d5fe65-a5d1-4b8a-8d40-ed92a6a35d8b","net":"ws","path":"/path","port":"6666","ps":"到期时间:2024-03-06  |  剩余流量:20.00GB","scy":"auto","sni":"","tls":"reality","type":"","v":"2"}
	node := new(model.NodeShared)
	node.Enabled = true
	node.NodeType = "vmess"
	node.IsSharedNode = true
	if strings.ToLower(link[:8]) == "vmess://" {
		link = link[8:]
	} else {
		return nil
	}
	if len(link) == 0 {
		return nil
	}
	jsonStr := encrypt_plugin.SubBase64Decode(link)
	if jsonStr == "" {
		return nil
	}
	var mapResult map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &mapResult)
	if err != nil {
		return nil
	}
	if version, ok := mapResult["v"]; ok {
		node.V = fmt.Sprintf("%v", version)
	}
	if ps, ok := mapResult["ps"]; ok {
		node.Remarks = fmt.Sprintf("%v", ps) //别名
	} else {
		return nil
	}
	if addr, ok := mapResult["add"]; ok {
		node.Address = fmt.Sprintf("%v", addr) //地址
	} else {
		return nil
	}
	if scy, ok := mapResult["scy"]; ok {
		node.Scy = fmt.Sprintf("%v", scy) //加密方式 auto,none,chacha20-poly1305,aes-128-gcm,zero
	} else {
		node.Scy = "auto"
	}
	if port, ok := mapResult["port"]; ok {
		value, err := strconv.ParseInt(fmt.Sprintf("%v", port), 10, 64)
		if err == nil {
			node.Port = value //端口
		} else {
			return nil
		}
	} else {
		return nil
	}

	if id, ok := mapResult["id"]; ok {
		node.UUID = fmt.Sprintf("%v", id) //uuid
	} else {
		return nil
	}
	if aid, ok := mapResult["aid"]; ok {
		if value, err := strconv.ParseInt(fmt.Sprintf("%v", aid), 10, 64); err == nil {
			node.Aid = value //额外id
		} else {
			return nil
		}
	} else {
		return nil
	}
	if net, ok := mapResult["net"]; ok {
		node.Network = fmt.Sprintf("%v", net) //传输协议
	} else {
		return nil
	}
	if type1, ok := mapResult["type"]; ok {
		node.Type = fmt.Sprintf("%v", type1)
	} else {
		return nil
	}

	//获取混淆
	if host, ok := mapResult["host"]; ok {
		node.Host = fmt.Sprintf("%v", host)
	} else {
		return nil
	}

	if path, ok := mapResult["path"]; ok {
		node.Path = fmt.Sprintf("%v", path)
	} else {
		return nil
	}
	if tls, ok := mapResult["tls"]; ok {
		node.Security = fmt.Sprintf("%v", tls)
	} else {
		return nil
	}
	if sni, ok := mapResult["sni"]; ok {
		node.Sni = fmt.Sprintf("%v", sni)
	}
	if alpn, ok := mapResult["alpn"]; ok {
		node.Alpn = fmt.Sprintf("%v", alpn)
	}
	return node
}

func ParseVLessLink(link string) *model.NodeShared {
	// vless例子 vless://d342d11e-d424-4583-b36e-524ab1f0afa7@1.6.1.1:443?path=%2F%3Fed%3D2048&security=tls&flow=xtls-rprx-vision-udp443&encryption=none&alpn=h2,http/1.1&host=v2.airgoo.link&fp=randomized&type=ws&sni=v2.airgoo.link#v2.airgoo.link
	// vless例子 vless://d342d11e-d424-4583-b36e-524ab1f0afa7@1.6.1.4:443?path=%2F%3Fed%3D2048&security=reality&flow=xtls-rprx-vision-udp443&encryption=none&pbk=ppkk&host=v2.airgoo.link&fp=randomized&spx=ssxx&type=ws&sni=v2.airgoo.link&sid=ssdd#v2.airgoo.link
	// [scheme:][//[userinfo@]host][/]path[?query][#fragment]
	u, err := url.Parse(link)
	if err != nil {
		return nil
	}
	if u.User == nil || u.Scheme != "vless" {
		return nil
	}
	node := new(model.NodeShared)
	node.Enabled = true
	node.NodeType = "vless"
	node.IsSharedNode = true

	//remarks
	node.Remarks = u.Fragment
	if node.Remarks == "" {
		node.Remarks = u.Host
	}
	//address
	node.Address = u.Hostname()
	//port
	node.Port, err = strconv.ParseInt(u.Port(), 10, 64)
	if err != nil {
		return nil
	}
	//uuid
	node.UUID = u.User.Username()

	//解析参数
	urlQuery := u.Query()
	if urlQuery.Get("flow") != "" {
		node.VlessFlow = urlQuery.Get("flow")
	}
	if urlQuery.Get("encryption") != "" {
		node.VlessEncryption = urlQuery.Get("encryption")
	}
	if urlQuery.Get("type") != "" {
		node.Network = urlQuery.Get("type")
	}
	if urlQuery.Get("security") != "" {
		node.Security = urlQuery.Get("security")
	}
	//获取混淆
	if urlQuery.Get("host") != "" {
		node.Host = urlQuery.Get("host")
	} else {
		return nil
	}

	if urlQuery.Get("path") != "" {
		node.Path = urlQuery.Get("path")
	}

	if urlQuery.Get("sni") != "" {
		node.Sni = urlQuery.Get("sni")
	}
	if urlQuery.Get("alpn") != "" {
		node.Alpn = urlQuery.Get("alpn")
	}
	if urlQuery.Get("allowInsecure") != "" {
		node.AllowInsecure = true
	}
	return node
}

func ParseTrojanLink(link string) *model.NodeShared {
	u, err := url.Parse(link)
	if err != nil {
		return nil
	}
	if u.User == nil || u.Scheme != "trojan" {
		return nil
	}
	node := new(model.NodeShared)
	node.Enabled = true
	node.NodeType = "trojan"
	node.IsSharedNode = true
	//remarks
	node.Remarks = u.Fragment
	if node.Remarks == "" {
		node.Remarks = u.Host
	}
	//address
	node.Address = u.Hostname()
	//port
	node.Port, err = strconv.ParseInt(u.Port(), 10, 64)
	if err != nil {
		return nil
	}
	//uuid
	node.UUID = u.User.Username()

	//解析参数
	urlQuery := u.Query()
	if urlQuery.Get("network") != "" {
		node.Network = urlQuery.Get("network")
	}
	if urlQuery.Get("type") != "" {
		node.Type = urlQuery.Get("type")
	}
	//获取混淆
	if urlQuery.Get("host") != "" {
		node.Host = urlQuery.Get("host")
	} else {
		return nil
	}
	if urlQuery.Get("path") != "" {
		node.Path = urlQuery.Get("path")
	}
	if urlQuery.Get("tls") != "" {
		node.Security = urlQuery.Get("tls")
	}
	if urlQuery.Get("sni") != "" {
		node.Sni = urlQuery.Get("sni")
	}
	if urlQuery.Get("alpn") != "" {
		node.Alpn = urlQuery.Get("alpn")
	}
	if urlQuery.Get("allowInsecure") != "" {
		node.AllowInsecure = true
	}

	return node
}

func ParseSSLink(link string) *model.NodeShared {
	ss, err := url.Parse(link)
	var node model.NodeShared
	if err != nil {
		global.Logrus.Error(err.Error())
		return nil
	}
	node.NodeType = "shadowsocks"
	node.Remarks = ss.Fragment
	node.Address = ss.Hostname()
	node.Port, err = strconv.ParseInt(ss.Port(), 10, 64)

	p, _ := SubBase64Decode(ss.User.String())
	arr := strings.SplitN(p, ":", 2)
	node.Scy = arr[0]
	node.UUID = arr[1] //Passwd存到uuid字段

	return &node
}

func ParseSubUrl(urlStr string) *[]model.NodeShared {
	//去掉前后空格
	urlStr = strings.TrimSpace(urlStr)
	//订阅url
	if !strings.HasPrefix(urlStr, "vmess") && !strings.HasPrefix(urlStr, "vless") && !strings.HasPrefix(urlStr, "trojan") {
		if _, err := url.ParseRequestURI(urlStr); err == nil {
			rsp, err := net_plugin.ClientWithDNS("223.6.6.6", 5*time.Second).Get(urlStr)
			if err != nil {
				return nil
			}
			defer rsp.Body.Close()
			subLink := net_plugin.ReadDate(rsp)
			if len(subLink) == 0 {
				return nil
			}
			urlStr = subLink
		}
	}
	// base64编码
	if urlStrBase64Decode, err := SubBase64Decode(urlStr); err == nil {
		urlStr = urlStrBase64Decode
	}
	list := strings.Fields(urlStr) //节点url数组
	var Nodes []model.NodeShared
	for _, v := range list {
		data := ParseOne(v)
		if data == nil {
			continue
		}
		Nodes = append(Nodes, *data)
	}
	return &Nodes
}

// 解析一条节点,vmess vless trojan
func ParseOne(link string) *model.NodeShared {
	//fmt.Println("解析一条链接", link)
	u, err := url.Parse(link)
	if err != nil {
		return nil
	}
	switch u.Scheme {
	case "vmess":
		if obj := ParseVMessLink(link); obj != nil {
			return obj
		}
	case "vless":
		if obj := ParseVLessLink(link); obj != nil {
			return obj
		}
	case "trojan":
		if obj := ParseTrojanLink(link); obj != nil {
			return obj
		}
	case "ss":
		if obj := ParseSSLink(link); obj != nil {
			return obj
		}
	}
	return nil
}

// 对节点base64格式进行解析
func SubBase64Decode(str string) (string, error) {
	i := len(str) % 4
	switch i {
	case 1:
		str = str[:len(str)-1]
	case 2:
		str += "=="
	case 3:
		str += "="
	}
	//str = strings.Split(str, "//")[1]
	var data []byte
	var err error
	if strings.Contains(str, "-") || strings.Contains(str, "_") {
		data, err = base64.URLEncoding.DecodeString(str)

	} else {
		data, err = base64.StdEncoding.DecodeString(str)
	}
	if err != nil {
		fmt.Println(err)
	}
	return string(data), err
}