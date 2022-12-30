package connect

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	cryptoSsh "golang.org/x/crypto/ssh"

	"github.com/gerrittrigger/trigger/config"
)

const (
	counter = 2
	prefix  = "gerrit "
)

type Ssh interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Reconnect(context.Context) error
	Run(context.Context, string) (string, error)
	Start(context.Context, string, chan string) error
}

type SshConfig struct {
	Config config.Config
	Logger hclog.Logger
}

type ssh struct {
	cfg          *SshConfig
	client       *cryptoSsh.Client
	clientConfig *cryptoSsh.ClientConfig
	sessions     []*cryptoSsh.Session
}

func SshNew(_ context.Context, cfg *SshConfig) Ssh {
	return &ssh{
		cfg:          cfg,
		client:       nil,
		clientConfig: nil,
		sessions:     []*cryptoSsh.Session{},
	}
}

func DefaultSshConfig() *SshConfig {
	return &SshConfig{}
}

func (s *ssh) Init(_ context.Context) error {
	s.cfg.Logger.Debug("ssh: Init")

	var err error

	key, err := os.ReadFile(s.cfg.Config.Spec.Connect.Ssh.Keyfile)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	signer, err := cryptoSsh.ParsePrivateKey(key)
	if err != nil {
		return errors.Wrap(err, "failed to parse key")
	}

	hostKeyCallback := func(name string, addr net.Addr, key cryptoSsh.PublicKey) error {
		return nil
	}

	s.clientConfig = &cryptoSsh.ClientConfig{
		User: s.cfg.Config.Spec.Connect.Ssh.Username,
		Auth: []cryptoSsh.AuthMethod{
			cryptoSsh.PublicKeys(signer),
		},
		HostKeyAlgorithms: []string{
			cryptoSsh.KeyAlgoDSA,
			cryptoSsh.KeyAlgoECDSA256,
			cryptoSsh.KeyAlgoECDSA384,
			cryptoSsh.KeyAlgoECDSA521,
			cryptoSsh.KeyAlgoED25519,
			cryptoSsh.KeyAlgoRSA,
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: hostKeyCallback,
	}

	host := s.cfg.Config.Spec.Connect.Hostname
	port := s.cfg.Config.Spec.Connect.Ssh.Port

	s.client, err = cryptoSsh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), s.clientConfig)
	if err != nil {
		s.client = nil
		return errors.Wrap(err, "failed to connect server")
	}

	return nil
}

func (s *ssh) Deinit(_ context.Context) error {
	s.cfg.Logger.Debug("ssh: Deinit")

	for i := range s.sessions {
		if s.sessions[i] != nil {
			_ = s.sessions[i].Close()
			s.sessions[i] = nil
		}
	}

	if s.client != nil {
		_ = s.client.Close()
		s.client = nil
	}

	return nil
}

func (s *ssh) Reconnect(ctx context.Context) error {
	s.cfg.Logger.Debug("ssh: Reconnect")

	var err error

	_ = s.Deinit(ctx)

	host := s.cfg.Config.Spec.Connect.Hostname
	port := s.cfg.Config.Spec.Connect.Ssh.Port

	s.client, err = cryptoSsh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), s.clientConfig)
	if err != nil {
		s.client = nil
		return errors.Wrap(err, "failed to connect server")
	}

	return nil
}

func (s *ssh) Run(_ context.Context, cmd string) (string, error) {
	s.cfg.Logger.Debug("ssh: Run")

	if s.client == nil {
		return "", errors.New("invalid client")
	}

	session, err := s.client.NewSession()

	defer func(session *cryptoSsh.Session) {
		if session != nil {
			_ = session.Close()
		}
	}(session)

	if err != nil {
		return "", errors.Wrap(err, "failed to create session")
	}

	out, err := session.CombinedOutput(prefix + cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to run session")
	}

	return string(out), nil
}

func (s *ssh) Start(_ context.Context, cmd string, out chan string) error {
	s.cfg.Logger.Debug("ssh: Start")

	var wg sync.WaitGroup

	helper := func(r io.Reader) {
		defer wg.Done()
		scan := bufio.NewScanner(r)
		scan.Split(bufio.ScanLines)
		for scan.Scan() {
			out <- scan.Text()
		}
		_ = scan.Err()
	}

	if s.client == nil {
		return errors.New("invalid client")
	}

	session, err := s.client.NewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create session")
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		if session != nil {
			_ = session.Close()
		}
		return errors.Wrap(err, "failed to pipe stderr")
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		if session != nil {
			_ = session.Close()
		}
		return errors.Wrap(err, "failed to pipe stdout")
	}

	wg.Add(counter)

	go helper(stderr)
	go helper(stdout)

	if err := session.Start(prefix + cmd); err != nil {
		if session != nil {
			_ = session.Close()
		}
		return errors.Wrap(err, "failed to start session")
	}

	s.sessions = append(s.sessions, session)

	wg.Add(1)

	go func() {
		defer wg.Done()
		_ = session.Wait()
	}()

	go func() {
		wg.Wait()
	}()

	return nil
}
