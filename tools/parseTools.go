package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"stressTest/global"
	"stressTest/logger"
	"strings"
)

func GetEnvVarOrDefault(key, defaultValue string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		logger.Error(fmt.Sprintf("未找到环境变量:%s， 使用默认值:%s.", key, defaultValue))
		return defaultValue
	}
	logger.Debug(fmt.Sprintf("环境变量%s的值为:%s.", key, val))
	return val
}

func GetAccountToFile(filePath string) []string {
	rs, _ := PathExists(filePath)
	if !rs {
		logger.Error("账号文件路径错误")
		return nil
	}

	ext := filepath.Ext(global.YokaConfig.AccountFileName)
	switch ext {
	case ".txt":
		content, err := os.ReadFile(global.YokaConfig.AccountFileName)
		if err != nil {
			logger.Error("读取账号文件失败")
			return nil
		}
		return strings.Split(strings.Replace(string(content), "\r", "", -1), "\n")
	}
	return nil
}
