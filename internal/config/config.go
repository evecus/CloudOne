package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config 只保留服务监听地址配置。
// 账号密码、JWT Secret、WebDAV 凭证等敏感数据全部存入加密的 .db 文件，
// 不再写入任何纯文本配置文件。
type Config struct {
	Host string
	Port string
}

func Default() *Config {
	return &Config{
		Host: "0.0.0.0",
		Port: "6677",
	}
}

// Load 只读取 [server] 段的 host 和 port。
func Load(path string) (*Config, error) {
	c := Default()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return c, err
	}
	defer f.Close()

	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if idx := strings.Index(val, " #"); idx != -1 {
			val = strings.TrimSpace(val[:idx])
		}
		switch section + "." + key {
		case "server.port":
			c.Port = val
		case "server.host":
			c.Host = val
		}
	}
	return c, scanner.Err()
}

// Save 只写入 host 和 port，不写入任何敏感信息。
func Save(path string, c *Config) error {
	content := fmt.Sprintf(
		"# CloudOne Configuration File\n"+
			"[server]\n"+
			"port = %s\n"+
			"host = %s\n",
		c.Port, c.Host,
	)
	return os.WriteFile(path, []byte(content), 0600)
}
