package connect

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	cryptoSsh "golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"

	"github.com/gerrittrigger/trigger/config"
	"github.com/gerrittrigger/trigger/queue"
)

const (
	num    = -1
	prefix = "gerrit "
)

type Ssh interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Run(context.Context, string) (string, error)
	Start(context.Context, string, queue.Queue) error
	Reconnect(context.Context) error
}

type SshConfig struct {
	Config config.Config
	Logger hclog.Logger
}

type ssh struct {
	cfg          *SshConfig
	client       *cryptoSsh.Client
	clientConfig *cryptoSsh.ClientConfig
	session      *cryptoSsh.Session
}

func SshNew(_ context.Context, cfg *SshConfig) Ssh {
	return &ssh{
		cfg:          cfg,
		client:       nil,
		clientConfig: nil,
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
		Timeout:         time.Duration(s.cfg.Config.Spec.Watchdog.TimeoutSeconds) * time.Second,
		HostKeyCallback: hostKeyCallback,
	}

	host := s.cfg.Config.Spec.Connect.Hostname
	port := s.cfg.Config.Spec.Connect.Ssh.Port

	s.client, err = cryptoSsh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), s.clientConfig)
	if err != nil {
		s.client = nil
		return errors.Wrap(err, "failed to connect server")
	}

	s.session, err = s.client.NewSession()
	if err != nil {
		_ = s.client.Close()
		s.client = nil
		return errors.Wrap(err, "failed to create session")
	}

	return nil
}

func (s *ssh) Deinit(_ context.Context) error {
	s.cfg.Logger.Debug("ssh: Deinit")

	if s.session != nil {
		_ = s.session.Close()
		s.session = nil
	}

	if s.client != nil {
		_ = s.client.Close()
		s.client = nil
	}

	return nil
}

func (s *ssh) Run(_ context.Context, cmd string) (string, error) {
	if s.client == nil {
		return "", errors.New("invalid client")
	}

	out, err := s.session.CombinedOutput(prefix + cmd)
	if err != nil {
		return "", errors.Wrap(err, "failed to run session")
	}

	return string(out), nil
}

func (s *ssh) Start(ctx context.Context, cmd string, _queue queue.Queue) error {
	helper := func(r io.Reader) error {
		scan := bufio.NewScanner(r)
		scan.Split(bufio.ScanLines)
		for scan.Scan() {
			_ = _queue.Put(ctx, scan.Text())
		}
		return scan.Err()
	}

	if s.client == nil {
		return errors.New("invalid client")
	}

	stderr, err := s.session.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "failed to pipe stderr")
	}

	stdout, err := s.session.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to pipe stdout")
	}

	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(num)

	g.Go(func() error {
		return helper(stderr)
	})

	g.Go(func() error {
		return helper(stdout)
	})

	if err := s.session.Start(prefix + cmd); err != nil {
		return errors.Wrap(err, "failed to start session")
	}

	g.Go(func() error {
		_ = s.session.Wait()
		return nil
	})

	return nil
}

func (s *ssh) Reconnect(ctx context.Context) error {
	var err error

	_ = s.Deinit(ctx)

	host := s.cfg.Config.Spec.Connect.Hostname
	port := s.cfg.Config.Spec.Connect.Ssh.Port

	s.client, err = cryptoSsh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), s.clientConfig)
	if err != nil {
		s.client = nil
		return errors.Wrap(err, "failed to connect server")
	}

	s.session, err = s.client.NewSession()
	if err != nil {
		_ = s.client.Close()
		s.client = nil
		return errors.Wrap(err, "failed to create session")
	}

	return nil
}
