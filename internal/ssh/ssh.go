package ssh

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"os"
	"path/filepath"
)

const knownHostsFile = "data/known_hosts"

var khMu sync.Mutex

func ensureKnownHostsFile() error {
	if err := os.MkdirAll(filepath.Dir(knownHostsFile), 0750); err != nil {
		return err
	}
	f, err := os.OpenFile(knownHostsFile, os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	return f.Close()
}

// tofuHostKeyCallback 首次连接记录指纹（TOFU），后续严格校验防 MITM
func tofuHostKeyCallback(hostname string, remote net.Addr, key gossh.PublicKey) error {
	khMu.Lock()
	defer khMu.Unlock()

	if err := ensureKnownHostsFile(); err != nil {
		return fmt.Errorf("failed to ensure known_hosts: %w", err)
	}

	checkCb, err := knownhosts.New(knownHostsFile)
	if err != nil {
		return fmt.Errorf("failed to load known_hosts: %w", err)
	}

	err = checkCb(hostname, remote, key)
	if err == nil {
		return nil
	}

	var keyErr *knownhosts.KeyError
	if ke, ok := err.(*knownhosts.KeyError); ok {
		keyErr = ke
	}
	if keyErr != nil && len(keyErr.Want) == 0 {
		// 首次连接：TOFU 记录
		line := knownhosts.Line([]string{hostname}, key)
		f, werr := os.OpenFile(knownHostsFile, os.O_APPEND|os.O_WRONLY, 0600)
		if werr != nil {
			return fmt.Errorf("failed to write known_hosts: %w", werr)
		}
		defer f.Close()
		if _, werr = fmt.Fprintln(f, line); werr != nil {
			return fmt.Errorf("failed to append known_hosts: %w", werr)
		}
		return nil
	}
	return fmt.Errorf("SSH host key mismatch for %s: possible MITM attack", hostname)
}

// Config SSH 连接配置，密码和私钥二选一
type Config struct {
	Host       string
	Port       int
	Username   string
	Password   string // 密码认证
	PrivateKey []byte // 私钥认证（PEM 格式）
}

// Session 表示一个活跃的 SSH 终端会话
type Session struct {
	client  *gossh.Client
	session *gossh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
}

// Connect 建立 SSH 连接并启动 shell
func Connect(cfg Config) (*Session, error) {
	var authMethods []gossh.AuthMethod

	if len(cfg.PrivateKey) > 0 {
		signer, err := gossh.ParsePrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, gossh.PublicKeys(signer))
	}

	if cfg.Password != "" {
		authMethods = append(authMethods, gossh.Password(cfg.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method: please set password or private key in SSH settings")
	}

	clientConfig := &gossh.ClientConfig{
		User:            cfg.Username,
		Auth:            authMethods,
		HostKeyCallback: tofuHostKeyCallback,
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client, err := gossh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	modes := gossh.TerminalModes{
		gossh.ECHO:          1,
		gossh.TTY_OP_ISPEED: 14400,
		gossh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", 24, 80, modes); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to request pty: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to get stdin: %w", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to get stdout: %w", err)
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to get stderr: %w", err)
	}

	if err := session.Shell(); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	return &Session{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
	}, nil
}

func (s *Session) Write(data []byte) (int, error) { return s.stdin.Write(data) }

func (s *Session) Resize(rows, cols uint32) error {
	return s.session.WindowChange(int(rows), int(cols))
}

func (s *Session) ReadLoop(outCh chan<- []byte, errCh chan<- error) {
	pipe := func(r io.Reader, isStderr bool) {
		buf := make([]byte, 32*1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				outCh <- data
			}
			if err != nil {
				if !isStderr {
					if err != io.EOF {
						errCh <- err
					} else {
						errCh <- nil
					}
				}
				return
			}
		}
	}
	go pipe(s.stdout, false)
	go pipe(s.stderr, true)
}

func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
	if s.client != nil {
		s.client.Close()
	}
}
