package tools

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"stressTest/logger"
	"syscall"
)

var (
	XmlConfig ConfigInfo
)

type ConfigInfo struct {
	ConfigName xml.Name   `xml:"ROOT"`
	BaseConfig BaseConfig `xml:"BASE"`
	TagsConfig TagsConfig `xml:"Tags"`
}

type BaseConfig struct {
	BaseConfigInfo []BaseConfigInfo `xml:"CONF"`
}

type BaseConfigInfo struct {
	StepConfXMLBaseConfigAttr string `xml:"ATTR,attr"`
	StepConfXMLBaseConfigName string `xml:"NAME,attr"`
}

type TagsConfig struct {
	TagsConfigInfo []TagsConfigInfo `xml:"tag"`
}

type TagsConfigInfo struct {
	TagValue string `xml:"VALUE,attr"`
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetStrPathExists(path string) (strPath string) {
	//strPath = filepath.Dir(os.Args[0]) + "/" + path
	strPath = path
	rs, _ := PathExists(strPath)
	if rs {
		return strPath
	}
	//log.Println(strPath + " is not Exists")
	str, _ := os.Getwd()
	strPath = str + "/src/stableTest/" + path
	rs, _ = PathExists(strPath)
	if rs {
		return strPath
	}
	//log.Println(strPath + " is not Exists")
	strPath = str + "/stableTest/" + path
	rs, _ = PathExists(strPath)
	if rs {
		return strPath
	}
	//log.Println(strPath + " is not Exists")
	strPath = str + "/" + path
	rs, _ = PathExists(strPath)
	if rs {
		return strPath
	}
	log.Println(filepath.Dir(os.Args[0]) + "/" + path + " is not Exists")
	log.Println(str + "/src/stableTest/" + path + " is not Exists")
	log.Println(str + "/stableTest/" + path + " is not Exists")
	log.Println(str + "/" + path + " is not Exists")
	//panic("GetPath Error, check the Path.")
	return ""
}

func MkdirPath() {
	dic := ""
	dic = GetStrPathExists("/logs/")
	if dic == "" {
		dic = GetStrPathExists("./../logs/")
		if dic == "" {
			//panic("Write log path Error, check the path.")
			err := os.Mkdir("./logs", os.ModePerm)
			if err != nil {
				return
			}
			dic = GetStrPathExists("/logs/")
		}
	}
}

func ReadConf(fileName string) {
	flag.Parse()
	// logger.Debug("fileName:", fileName)

	if fileName != "" {
		filePath := GetStrPathExists(fileName)
		if filePath == "" {
			logger.Error("The path to the XML file is wrong.")
			panic("The path to the XML file is wrong.")
		}
		// logger.Debug("Read document configuration.")
		content, err := os.ReadFile(fileName)
		if err != nil {
			logger.Error("Path does not exist")
			panic(err)
		}
		err = xml.Unmarshal(content, &XmlConfig)
		if err != nil {
			logger.Errorf("Unmarshal failure:", err)
			panic(err)
		}
		// logger.Debug("xml file configuration details:", StepResult)
	}
}

func GetStringMd5(s string) string {
	_md5 := md5.New()
	_md5.Write([]byte(s))
	md5Str := hex.EncodeToString(_md5.Sum(nil))
	return md5Str
}

// GetStringByBytes 从[]byte中获得字符串
func GetStringByBytes(bytes []byte) (out string) {
	for i, c := range bytes {
		if c == 0 {
			return string(bytes[:i])
		}
	}

	return string(bytes)
}

func GetMacAddrs() (macAddrs string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddrs
	}

	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) == 0 {
			continue
		}

		macAddrs += macAddr
		break
	}
	return macAddrs
}

func StringHash(s string) uint32 {
	//initCRCTable()

	var hash uint32
	var b uint32

	for _, c := range s {

		b = uint32(c)

		hash = ((hash >> 8) & 0x00FFFFFF) ^ crcTable[(hash^b)&0x000000FF]
	}

	return hash
}

var crcTable = make([]uint32, 256)

func SetupSignalHandler(cleanupFunc func()) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		os.Interrupt,    // Ctrl+C
		syscall.SIGTERM, // kill <pid>
		syscall.SIGHUP,  // 终端断开
		syscall.SIGQUIT, // Ctrl+\
	)

	go func() {
		sig := <-interrupt
		log.Printf("接收到信号: %v", sig)

		// 执行清理操作
		if cleanupFunc != nil {
			cleanupFunc()
		}

		// 根据信号类型处理
		switch sig {
		case syscall.SIGTERM:
			log.Println("接收到终止信号，优雅退出")
		case os.Interrupt:
			log.Println("用户中断程序")
		case syscall.SIGHUP:
			log.Println("终端连接断开")
		}

		os.Exit(0)
	}()
}
