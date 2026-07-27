// Package terminal provides a WebSocket-based PTY shell terminal.
// It spawns a local shell (bash/sh) using a pseudo-terminal so that
// interactive programs work correctly — no SSH required.
package terminal

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"

	"github.com/gorilla/websocket"
)

// WSMessage is the JSON envelope exchanged with the browser xterm.js client.
type WSMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// openPTY opens a PTY master/slave pair on Linux.
func openPTY() (master, slave *os.File, err error) {
	fd, e := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_CLOEXEC, 0)
	if e != nil {
		return nil, nil, e
	}
	master = os.NewFile(uintptr(fd), "/dev/ptmx")

	// Unlock slave PTY
	var n int32 // 0 = unlock
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&n))); errno != 0 {
		master.Close()
		return nil, nil, errno
	}

	// Get slave device number
	var ptn uint32
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, master.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&ptn))); errno != 0 {
		master.Close()
		return nil, nil, errno
	}

	slavePath := "/dev/pts/" + uitoa(ptn)
	slave, e = os.OpenFile(slavePath, os.O_RDWR|syscall.O_NOCTTY, 0)
	if e != nil {
		master.Close()
		return nil, nil, e
	}
	return master, slave, nil
}

func resizePTY(f *os.File, rows, cols uint16) {
	ws := &winsize{Row: rows, Col: cols}
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(ws)))
}

func uitoa(n uint32) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}

// Handle bridges a WebSocket connection to a local PTY shell session.
// cwd, if non-empty and valid, sets the shell's initial working directory.
// It blocks until the connection or shell exits.
func Handle(conn *websocket.Conn, cwd string) {
	mu := &sync.Mutex{}
	send := func(v interface{}) {
		mu.Lock()
		defer mu.Unlock()
		conn.WriteJSON(v)
	}

	master, slave, err := openPTY()
	if err != nil {
		send(WSMessage{Type: "error", Data: "Failed to open PTY: " + err.Error()})
		return
	}
	defer master.Close()

	shell := "/bin/bash"
	if _, e := os.Stat(shell); e != nil {
		shell = "/bin/sh"
	}

	// 以登录 shell 启动（argv[0] = "-bash"），使 shell 读取 /etc/profile、~/.profile、~/.bash_profile，
	// 从而与 SSH 登录会话一致地加载用户配置的 PATH 等环境变量。
	cmd := exec.Command(shell, "-l")
	cmd.Args[0] = "-" + filepath.Base(shell)
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)
	// 若调用方提供了合法存在的目录，则终端以该目录作为初始工作目录，
	// 使终端与用户在文件管理器中浏览的位置保持一致。
	if cwd != "" {
		if info, err := os.Stat(cwd); err == nil && info.IsDir() {
			cmd.Dir = cwd
		}
	}
	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0, // fd 0 = stdin inside child = slave
	}

	if err := cmd.Start(); err != nil {
		slave.Close()
		send(WSMessage{Type: "error", Data: "Failed to start shell: " + err.Error()})
		return
	}
	slave.Close() // parent closes its copy; child keeps it via inheritance

	send(WSMessage{Type: "connected"})

	// PTY → WebSocket goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := master.Read(buf)
			if n > 0 {
				send(WSMessage{Type: "output", Data: string(buf[:n])})
			}
			if err != nil {
				if err != io.EOF {
					send(WSMessage{Type: "error", Data: err.Error()})
				}
				send(WSMessage{Type: "closed"})
				return
			}
		}
	}()

	// WebSocket → PTY
	wsMsgs := make(chan []byte, 64)
	go func() {
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				close(wsMsgs)
				return
			}
			wsMsgs <- raw
		}
	}()

	for {
		select {
		case <-done:
			cmd.Process.Kill()
			cmd.Wait()
			return
		case raw, ok := <-wsMsgs:
			if !ok {
				cmd.Process.Kill()
				cmd.Wait()
				return
			}
			var msg WSMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case "input":
				master.Write([]byte(msg.Data))
			case "resize":
				rows, cols := msg.Rows, msg.Cols
				if rows == 0 {
					rows = 24
				}
				if cols == 0 {
					cols = 80
				}
				resizePTY(master, rows, cols)
			}
		}
	}
}
