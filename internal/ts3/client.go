package ts3

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents a connection to a TeamSpeak 3 ServerQuery interface.
type Client struct {
	conn       net.Conn
	sshClient  *ssh.Client
	sshSession *ssh.Session
	reader     *bufio.Reader
	writer     io.Writer
	addr       string
	user       string
	pass       string
	proto      string
	timeout    time.Duration
}

// NewClient creates a new TS3 client.
func NewClient(host string, port int, user, pass, proto string, timeout time.Duration) (*Client, error) {
	if proto == "" {
		proto = "tcp"
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	client := &Client{
		addr:    addr,
		user:    user,
		pass:    pass,
		proto:   proto,
		timeout: timeout,
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) connect() error {
	var err error
	if c.proto == "ssh" {
		config := &ssh.ClientConfig{
			User: c.user,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Support host key verification
			Timeout:         c.timeout,
		}
		c.sshClient, err = ssh.Dial("tcp", c.addr, config)
		if err != nil {
			return err
		}

		c.sshSession, err = c.sshClient.NewSession()
		if err != nil {
			c.sshClient.Close()
			return err
		}

		stdin, err := c.sshSession.StdinPipe()
		if err != nil {
			c.sshSession.Close()
			c.sshClient.Close()
			return err
		}
		stdout, err := c.sshSession.StdoutPipe()
		if err != nil {
			c.sshSession.Close()
			c.sshClient.Close()
			return err
		}

		if err := c.sshSession.Shell(); err != nil {
			c.sshSession.Close()
			c.sshClient.Close()
			return err
		}

		c.writer = stdin
		c.reader = bufio.NewReader(stdout)

	} else {
		dialer := net.Dialer{Timeout: c.timeout}
		c.conn, err = dialer.Dial("tcp", c.addr)
		if err != nil {
			return err
		}
		c.writer = c.conn
		c.reader = bufio.NewReader(c.conn)

		// Read initial welcome message
		// TS3 usually sends "TS3"
		// We can just consume until the first prompt, or just assume it's there.
		// Actually, for TCP, we need to login explicitly.
		// For SSH, login is handled by SSH protocol.

		// Read welcome banner
		// "TS3"
		// "Welcome to the..."
		// We should read until we get a clean state?
		// Actually, standard TCP query starts with "TS3\n\rWelcome..."
		// We can just login.
		// Wait, login command is `login user pass`.
	}

	// For TCP, we need to login.
	if c.proto == "tcp" {
		// Read banner lines (usually 2-3 lines)
		// We can try to read explicitly, or just send login.
		// If we send login immediately, we might have junk in buffer.
		// Let's flush? No, we need to read.
		// TS3 TCP banner:
		// TS3
		// Welcome to the ...
		// <blank line>?

		// Let's just consume the first two lines?
		c.conn.SetReadDeadline(time.Now().Add(c.timeout))
		line1, err := c.reader.ReadString('\n')
		if err != nil {
			c.Close()
			return err
		}
		if !strings.Contains(line1, "TS3") {
			// Warn but proceed?
		}

		// Usually there is a second line.
		// "Welcome to the..."
		// But let's verify if we need to consume it.
		// Some servers might not send it?
		// Standard TS3 sends it.
		// We can use a peek?
		// Let's try to read it.
		line2, err := c.reader.ReadString('\n')
		if err != nil {
			// Might be EOF if server is weird.
		}
		_ = line2 // Ignore

		// Login
		if _, err := c.Execute(fmt.Sprintf("login %s %s", Escape(c.user), Escape(c.pass))); err != nil {
			c.Close()
			return fmt.Errorf("login failed: %w", err)
		}
	} else {
        // SSH login is already done, but we might need to register for events or just ready to go.
        // SSH doesn't require "login" command usually because we logged in via SSH.
        // But for consistency we might need to select server?
        // Wait, "use" command is needed later.
    }

	return nil
}

// Close closes the connection.
func (c *Client) Close() error {
	var err error
	if c.sshSession != nil {
		c.sshSession.Close()
	}
	if c.sshClient != nil {
		err = c.sshClient.Close()
	}
	if c.conn != nil {
		err = c.conn.Close()
	}
	return err
}

// Execute sends a command and returns the response string.
func (c *Client) Execute(cmd string) (string, error) {
	// Send command
	if c.proto == "tcp" {
		c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	}
	if _, err := fmt.Fprintf(c.writer, "%s\n", cmd); err != nil {
		return "", err
	}

	// Read response
	var responseBuilder strings.Builder
	for {
		if c.proto == "tcp" {
			c.conn.SetReadDeadline(time.Now().Add(c.timeout))
		}
		// ReadString('\n') might block forever if server hangs, so timeouts are crucial.
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "error id=") {
			// Parse error
			// Format: error id=0 msg=ok
			parts := strings.Split(line, " ")
			var id, msg string
			for _, part := range parts {
				if strings.HasPrefix(part, "id=") {
					id = strings.TrimPrefix(part, "id=")
				}
				if strings.HasPrefix(part, "msg=") {
					msg = strings.TrimPrefix(part, "msg=")
				}
			}

			if id != "0" {
				return "", fmt.Errorf("TS3 error %s: %s", id, Unescape(msg))
			}

			// Success
			return responseBuilder.String(), nil
		}

		if responseBuilder.Len() > 0 {
			responseBuilder.WriteString("\n")
		}
		responseBuilder.WriteString(line)
	}
}
