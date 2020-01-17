// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authn "github.com/mainflux/mainflux/authn"
	httpapi "github.com/mainflux/mainflux/authn/api/http"
	"github.com/mainflux/mainflux/authn/jwt"
	"github.com/mainflux/mainflux/authn/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

const (
	secret       = "secret"
	contentType  = "application/json"
	invalidEmail = "userexample.com"
	wrongID      = "123e4567-e89b-12d3-a456-000000000042"
	id           = "123e4567-e89b-12d3-a456-000000000001"
	email        = "user@example.com"
)

type issueRequest struct {
	Duration time.Duration `json:"duration,omitempty"`
	Type     uint32        `json:"type,omitempty"`
}

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}

	req.Header.Set("Referer", "http://localhost")
	return tr.client.Do(req)
}

func newService() authn.Service {
	repo := mocks.NewKeyRepository()
	idp := mocks.NewIdentityProvider()
	t := jwt.New(secret)
	return authn.New(repo, idp, t)
}

func newServer(svc authn.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc, mocktracer.New())
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestIssue(t *testing.T) {
	svc := newService()
	userKey, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	uk := issueRequest{Type: authn.UserKey}
	ak := issueRequest{Type: authn.APIKey, Duration: time.Hour}
	rk := issueRequest{Type: authn.RecoveryKey}

	cases := []struct {
		desc   string
		req    string
		ct     string
		token  string
		status int
	}{
		{
			desc:   "issue user key",
			req:    toJSON(uk),
			ct:     contentType,
			token:  "",
			status: http.StatusCreated,
		},
		{
			desc:   "issue API key",
			req:    toJSON(ak),
			ct:     contentType,
			token:  userKey.Secret,
			status: http.StatusCreated,
		},
		{
			desc:   "issue recovery key",
			req:    toJSON(rk),
			ct:     contentType,
			token:  userKey.Secret,
			status: http.StatusBadRequest,
		},
		{
			desc: "issue user key wrong content type",
			req:  toJSON(uk),
			ct:   "", token: userKey.Secret,
			status: http.StatusUnsupportedMediaType,
		},
		{
			desc:   "issue key wrong content type",
			req:    toJSON(rk),
			ct:     "",
			token:  userKey.Secret,
			status: http.StatusUnsupportedMediaType,
		},
		{
			desc:   "issue key unauthorized",
			req:    toJSON(ak),
			ct:     contentType,
			token:  "wrong",
			status: http.StatusForbidden,
		},
		{
			desc:   "issue recovery key with empty token",
			req:    toJSON(rk),
			ct:     contentType,
			token:  "",
			status: http.StatusBadRequest,
		},
		{
			desc:   "issue key with invalid request",
			req:    "{",
			ct:     contentType,
			token:  "",
			status: http.StatusBadRequest,
		},
		{
			desc:   "issue key with invalid JSON",
			req:    "{invalid}",
			ct:     contentType,
			token:  "",
			status: http.StatusBadRequest,
		},
		{
			desc:   "issue key with invalid JSON content",
			req:    `{"Type":{"key":"value"}}`,
			ct:     contentType,
			token:  "",
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/keys", ts.URL),
			contentType: tc.ct,
			token:       tc.token,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRetrieve(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := authn.Key{Type: authn.APIKey, IssuedAt: time.Now()}

	k, err := svc.Issue(context.Background(), loginKey.Secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		desc   string
		id     string
		token  string
		status int
	}{
		{
			desc:   "retrieve an existing key",
			id:     k.ID,
			token:  loginKey.Secret,
			status: http.StatusOK,
		},
		{
			desc:   "retrieve a non-existing key",
			id:     "non-existing",
			token:  loginKey.Secret,
			status: http.StatusNotFound,
		},
		{
			desc:   "retrieve a key unauthorized",
			id:     k.ID,
			token:  "wrong",
			status: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: client,
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/keys/%s", ts.URL, tc.id),
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRevoke(t *testing.T) {
	svc := newService()
	userKey, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))
	key := authn.Key{Type: authn.APIKey, IssuedAt: time.Now()}

	k, err := svc.Issue(context.Background(), userKey.Secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		desc   string
		id     string
		token  string
		status int
	}{
		{
			desc:   "revoke an existing key",
			id:     k.ID,
			token:  userKey.Secret,
			status: http.StatusNoContent,
		},
		{
			desc:   "revoke a non-existing key",
			id:     "non-existing",
			token:  userKey.Secret,
			status: http.StatusNoContent,
		},
		{
			desc:   "revoke a key unauthorized",
			id:     k.ID,
			token:  "wrong",
			status: http.StatusForbidden},
	}

	for _, tc := range cases {
		req := testRequest{
			client: client,
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/keys/%s", ts.URL, tc.id),
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}
