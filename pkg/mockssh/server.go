package mockssh

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

type Server struct {
	// t and hostKey are set in NewServer.
	t       *testing.T
	hostKey ssh.Signer

	// If CertChecker is nil then a default checker will be used, which checks that the
	// certificate's public key is in CertAuthorityKeys.
	CertAuthorityKeys []ssh.PublicKey
	CertChecker       ssh.CertChecker

	// An optional CommandHandler, which responds to commands sent over SSH.
	// NewServer will give this a default using ExecHandler, which can also
	// be reused from custom handlers.
	CommandHandler CommandHandler

	// listener and port are set after Start.
	listener net.Listener
	port     int
}

type CommandIO struct {
	StdIn  io.Reader
	StdOut io.Writer
	StdErr io.Writer
}

type CommandHandler func(conn ssh.ConnMetadata, command string, commandIO CommandIO) int

// NewServer creates and starts a local SSH server for a test.
// It must be stopped with the Server.Stop method.
// The authorityEndpoint returns SSH public keys in JSON under the key "authorities".
func NewServer(t *testing.T, authorityEndpoint string) (*Server, error) {
	hk, err := generateHostKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate host key: %v", err)
	}

	// TODO use test context
	keys, err := fetchAuthorityKeys(context.Background(), authorityEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch authority keys: %v", err)
	}

	s := &Server{t: t, hostKey: hk}
	s.CommandHandler = ExecHandler("", nil)
	s.CertChecker = s.defaultCertChecker()
	s.CertAuthorityKeys = keys

	if err := s.start(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) HostKeyConfig() string {
	return fmt.Sprintf("[127.0.0.1]:%d %s %s",
		s.port,
		s.hostKey.PublicKey().Type(),
		base64.StdEncoding.EncodeToString(s.hostKey.PublicKey().Marshal()),
	)
}

func (s *Server) HostKey() ssh.PublicKey {
	return s.hostKey.PublicKey()
}

func (s *Server) start() error {
	t := s.t

	config := s.serverConfig()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	s.listener = listener
	if addr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.port = addr.Port
	}
	t.Logf("Test SSH server listening at %s", listener.Addr())

	go func(l net.Listener) {
		for {
			conn, err := l.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					break
				}
				t.Errorf("Failed to accept connection: %v", err)
				continue
			}

			sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
			if err != nil {
				t.Logf("Handshake failed: %v", err)
				return
			}

			t.Logf("Handling SSH connection from %s", sshConn.RemoteAddr())

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				ssh.DiscardRequests(reqs)
				wg.Done()
			}()
			wg.Add(1)
			go func() {
				s.handleChannels(sshConn, chans)
				wg.Done()
			}()
			wg.Wait()
		}
	}(s.listener)

	return nil
}

func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// ExecHandler returns a CommandHandler to execute a command in the given environment.
func ExecHandler(workingDir string, env []string) CommandHandler {
	return func(_ ssh.ConnMetadata, command string, commandIO CommandIO) int {
		c := exec.Command("bash", "-c", command)
		c.Stdout = commandIO.StdOut
		c.Stderr = commandIO.StdErr
		c.Stdin = commandIO.StdIn
		c.Dir = workingDir
		c.Env = env
		if err := c.Run(); err != nil {
			exitErr := &exec.ExitError{}
			if errors.As(err, &exitErr) {
				return exitErr.ExitCode()
			}
			_, _ = fmt.Fprintf(commandIO.StdErr, "Failed to execute command: %v", err)
			return 1
		}
		return 0
	}
}

func (s *Server) defaultCertChecker() ssh.CertChecker {
	return ssh.CertChecker{IsUserAuthority: func(auth ssh.PublicKey) bool {
		m := auth.Marshal()
		for _, ak := range s.CertAuthorityKeys {
			if bytes.Equal(ak.Marshal(), m) {
				return true
			}
		}
		return false
	}}
}

func (s *Server) serverConfig() *ssh.ServerConfig {
	t := s.t
	conf := &ssh.ServerConfig{}
	conf.AddHostKey(s.hostKey)
	conf.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if cert, ok := key.(*ssh.Certificate); ok {
			t.Logf("SSH certificate received from %s with key ID %s", conn.RemoteAddr(), cert.KeyId)
			return s.CertChecker.Authenticate(conn, cert)
		}
		return nil, fmt.Errorf("not accepting public key type: %s", key.Type())
	}
	conf.AuthLogCallback = func(conn ssh.ConnMetadata, method string, err error) {
		if err != nil {
			t.Logf("SSH auth log: client %s (%s), server %s, user %s, method %s, error: %s",
				conn.RemoteAddr(), conn.ClientVersion(), conn.LocalAddr(), conn.User(), method, err)
			return
		}

		t.Logf("SSH auth log: client %s, user %s, method %s", conn.RemoteAddr(), conn.User(), method)
	}

	return conf
}

func (s *Server) handleChannels(conn ssh.ConnMetadata, chans <-chan ssh.NewChannel) {
	t := s.t

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			err := newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			if err != nil {
				t.Errorf("Failed to reject channel: %v", err)
			}
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			t.Errorf("Failed to accept channel: %v", err)
			return
		}

		timer := time.NewTimer(time.Second * 10)

		var exitWithStatus = make(chan int, 1)
		go func(in <-chan *ssh.Request) {
			for req := range in {
				if !req.WantReply {
					continue
				}
				switch req.Type {
				case "exec":
					err := req.Reply(true, nil)
					if err != nil {
						t.Errorf("Failed to reply to command: %v", err)
					}
					// Strip the first four bytes of the payload, the uint32 representing the string length.
					// See https://datatracker.ietf.org/doc/html/rfc4251#section-5
					cmd := req.Payload[4:]
					t.Logf("Handling command: %s", cmd)
					exitWithStatus <- s.CommandHandler(conn, string(cmd), CommandIO{
						StdIn:  channel,
						StdOut: channel,
						StdErr: channel.Stderr(),
					})
					return
				default:
					_ = req.Reply(false, nil)
				}
			}
		}(requests)

		for {
			select {
			case s := <-exitWithStatus:
				_, err = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{uint32(s)})) //nolint: gosec
				if err != nil {
					t.Fatalf("Failed to send exit status: %v", err)
				}
				goto closeChannel
			case <-timer.C:
				t.Error("Timed out")
				goto closeChannel
			}
		}

	closeChannel:
		_ = channel.Close()
	}
}

var hostKey ssh.Signer

func generateHostKey() (ssh.Signer, error) {
	if hostKey == nil {
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		s, err := ssh.NewSignerFromKey(privateKey)
		if err != nil {
			return nil, err
		}
		hostKey = s
	}
	return hostKey, nil
}

func fetchAuthorityKeys(ctx context.Context, url string) ([]ssh.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch authority endpoint: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad authority endpoint status: %s", resp.Status)
	}
	var data struct {
		Authorities []string `json:"authorities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode authority response: %v", err)
	}
	sshPubKeys := make([]ssh.PublicKey, len(data.Authorities))
	for i, a := range data.Authorities {
		pk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(a))
		if err != nil {
			return nil, fmt.Errorf("failed to parse authority key: %v", err)
		}
		sshPubKeys[i] = pk
	}
	return sshPubKeys, nil
}
