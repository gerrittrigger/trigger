package connect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/gerrittrigger/trigger/config"
)

const (
	CHANGES   = "/changes/"
	DETAIL    = "/detail"
	REVIEW    = "/review"
	REVISIONS = "/revisions/"
	VERSION   = "/config/server/version"

	PREFIX = "/a"
)

const (
	optionAccounts = "DETAILED_ACCOUNTS"
	optionCommit   = "CURRENT_COMMIT"
	optionRevision = "CURRENT_REVISION"
)

type Rest interface {
	Init(context.Context) error
	Deinit(context.Context) error
	Detail(context.Context, int) (map[string]any, error)
	Query(context.Context, string, int) (map[string]any, error)
	Version(context.Context) (string, error)
	Vote(context.Context, int, int, string, string, string) error
}

type RestConfig struct {
	Config config.Config
	Logger hclog.Logger
}

type rest struct {
	cfg  *RestConfig
	pass string
	url  string
	user string
}

func RestNew(_ context.Context, cfg *RestConfig) Rest {
	return &rest{
		cfg:  cfg,
		pass: cfg.Config.Spec.Connect.Http.Password,
		url:  cfg.Config.Spec.Connect.FrontendUrl,
		user: cfg.Config.Spec.Connect.Http.Username,
	}
}

func DefaultRestConfig() *RestConfig {
	return &RestConfig{}
}

func (r *rest) Init(_ context.Context) error {
	r.cfg.Logger.Debug("rest: Init")
	return nil
}

func (r *rest) Deinit(_ context.Context) error {
	r.cfg.Logger.Debug("rest: Deinit")
	return nil
}

func (r *rest) Detail(_ context.Context, change int) (map[string]any, error) {
	url := r.url + CHANGES + strconv.Itoa(change) + DETAIL
	if r.user != "" && r.pass != "" {
		url = r.url + PREFIX + CHANGES + strconv.Itoa(change) + DETAIL
	}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set request")
	}

	if r.user != "" && r.pass != "" {
		req.SetBasicAuth(r.user, r.pass)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read")
	}

	var buf map[string]any

	if err := json.Unmarshal(data[4:], &buf); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}

	return buf, nil
}

func (r *rest) Query(_ context.Context, search string, start int) (map[string]any, error) {
	url := r.url + CHANGES
	if r.user != "" && r.pass != "" {
		url = r.url + PREFIX + CHANGES
	}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set request")
	}

	if r.user != "" && r.pass != "" {
		req.SetBasicAuth(r.user, r.pass)
	}

	q := req.URL.Query()

	q.Add("o", optionAccounts)
	q.Add("o", optionCommit)
	q.Add("o", optionRevision)
	q.Add("q", search)
	q.Add("start", strconv.Itoa(start))

	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read")
	}

	var buf []map[string]any

	if err := json.Unmarshal(data[4:], &buf); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal")
	}

	if len(buf) == 0 {
		return nil, errors.New("not matched")
	}

	return buf[0], nil
}

func (r *rest) Version(_ context.Context) (string, error) {
	url := r.url + VERSION
	if r.user != "" && r.pass != "" {
		url = r.url + PREFIX + VERSION
	}

	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to set request")
	}

	if r.user != "" && r.pass != "" {
		req.SetBasicAuth(r.user, r.pass)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to send request")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid status")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read")
	}

	var buf string

	if err := json.Unmarshal(data[4:], &buf); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal")
	}

	return buf, nil
}

func (r *rest) Vote(_ context.Context, change, revision int, label, message, vote string) error {
	url := r.url + CHANGES + strconv.Itoa(change) + REVISIONS + strconv.Itoa(revision) + REVIEW
	if r.user != "" && r.pass != "" {
		url = r.url + PREFIX + CHANGES + strconv.Itoa(change) + REVISIONS + strconv.Itoa(revision) + REVIEW
	}

	buf := map[string]any{
		"comments": nil,
		"labels":   map[string]any{label: vote},
		"message":  message,
	}

	body, err := json.Marshal(buf)
	if err != nil {
		return errors.Wrap(err, "failed to marshal")
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return errors.Wrap(err, "failed to set request")
	}

	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	if r.user != "" && r.pass != "" {
		req.SetBasicAuth(r.user, r.pass)
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}

	defer func() {
		_ = rsp.Body.Close()
	}()

	if rsp.StatusCode != http.StatusOK {
		return errors.New("invalid status")
	}

	_, err = io.ReadAll(rsp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read")
	}

	return nil
}
