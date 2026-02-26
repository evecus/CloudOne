package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds all runtime configuration for CloudOne.
// It is persisted to data/conf.ini and read on every startup.
type Config struct {
	Host       string
	Port       string
	StorageDir string
	Username   string
	Password   string // bcrypt hash
	Lang       string
	JWTSecret  string
}

func Default() *Config {
	return &Config{
		Host:       "0.0.0.0",
		Port:       "6677",
		StorageDir: "./data/storage",
		Username:   "",
		Password:   "",
		Lang:       "zh",
		JWTSecret:  "",
	}
}

// Load reads conf.ini from the given path.
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
		case "storage.dir":
			c.StorageDir = val
		case "account.username":
			c.Username = val
		case "account.password":
			c.Password = val
		case "general.lang":
			c.Lang = val
		case "general.jwt_secret":
			c.JWTSecret = val
		}
	}
	return c, scanner.Err()
}

// Save writes the current config to conf.ini at the given path.
func Save(path string, c *Config) error {
	content := fmt.Sprintf(
		"# CloudOne Configuration File\n"+
			"[server]\n"+
			"port = %s\n"+
			"host = %s\n\n"+
			"[storage]\n"+
			"dir = %s\n\n"+
			"[account]\n"+
			"username = %s\n"+
			"password = %s\n\n"+
			"[general]\n"+
			"lang = %s\n"+
			"jwt_secret = %s\n",
		c.Port, c.Host, c.StorageDir, c.Username, c.Password, c.Lang, c.JWTSecret,
	)
	return os.WriteFile(path, []byte(content), 0600)
}
