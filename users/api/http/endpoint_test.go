// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users"
	httpapi "github.com/mainflux/mainflux/users/api/http"
	"github.com/mainflux/mainflux/users/mocks"
	"github.com/mainflux/mainflux/users/token"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

const (
	contentType  = "application/json"
	invalidEmail = "userexample.com"
	wrongID      = "123e4567-e89b-12d3-a456-000000000042"
	id           = "123e4567-e89b-12d3-a456-000000000001"
)

var user = users.User{Email: "user@example.com", Password: "password"}

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
	return tr.client.Do(req)
}

func newService() users.Service {
	repo := mocks.NewUserRepository()
	hasher := mocks.NewHasher()
	idp := mocks.NewIdentityProvider()

	return users.New(repo, hasher, idp)
}

func newServer(svc users.Service) *httptest.Server {
	logger, _ := log.New(os.Stdout, log.Info.String())
	mux := httpapi.MakeHandler(svc, mocktracer.New(), logger)
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestRegister(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	data := toJSON(user)
	invalidData := toJSON(users.User{Email: invalidEmail, Password: "password"})
	invalidFieldData := fmt.Sprintf(`{"email": "%s", "pass": "%s"}`, user.Email, user.Password)

	cases := []struct {
		desc        string
		req         string
		contentType string
		status      int
	}{
		{"register new user", data, contentType, http.StatusCreated},
		{"register existing user", data, contentType, http.StatusConflict},
		{"register user with invalid email address", invalidData, contentType, http.StatusBadRequest},
		{"register user with invalid request format", "{", contentType, http.StatusBadRequest},
		{"register user with empty JSON request", "{}", contentType, http.StatusBadRequest},
		{"register user with empty request", "", contentType, http.StatusBadRequest},
		{"register user with invalid field name", invalidFieldData, contentType, http.StatusBadRequest},
		{"register user with missing content type", data, "", http.StatusUnsupportedMediaType},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/users", ts.URL),
			contentType: tc.contentType,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestLogin(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	tokenData := toJSON(map[string]string{"token": user.Email})
	data := toJSON(user)
	invalidEmailData := toJSON(users.User{
		Email:    invalidEmail,
		Password: "password",
	})
	invalidData := toJSON(users.User{
		Email:    "user@example.com",
		Password: "invalid_password",
	})
	nonexistentData := toJSON(users.User{
		Email:    "non-existentuser@example.com",
		Password: "pass",
	})
	svc.Register(context.Background(), user)

	cases := []struct {
		desc        string
		req         string
		contentType string
		status      int
		res         string
	}{
		{"login with valid credentials", data, contentType, http.StatusCreated, tokenData},
		{"login with invalid credentials", invalidData, contentType, http.StatusForbidden, ""},
		{"login with invalid email address", invalidEmailData, contentType, http.StatusBadRequest, ""},
		{"login non-existent user", nonexistentData, contentType, http.StatusForbidden, ""},
		{"login with invalid request format", "{", contentType, http.StatusBadRequest, ""},
		{"login with empty JSON request", "{}", contentType, http.StatusBadRequest, ""},
		{"login with empty request", "", contentType, http.StatusBadRequest, ""},
		{"login with missing content type", data, "", http.StatusUnsupportedMediaType, ""},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/tokens", ts.URL),
			contentType: tc.contentType,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		token := strings.Trim(string(body), "\n")

		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, token, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, token))
	}
}

func TestPasswordResetRequest(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	data := toJSON(user)

	nonexistentData := toJSON(users.User{
		Email:    "non-existentuser@example.com",
		Password: "pass",
	})

	expectedNonExistent := toJSON(struct {
		Msg string `json:"msg"`
	}{
		users.ErrUserNotFound.Error(),
	})

	expectedExisting := toJSON(struct {
		Msg string `json:"msg"`
	}{
		httpapi.MailSent,
	})

	svc.Register(context.Background(), user)

	cases := []struct {
		desc        string
		req         string
		contentType string
		status      int
		res         string
	}{
		{"password reset with valid email", data, contentType, http.StatusCreated, expectedExisting},
		{"password reset with invalid email", nonexistentData, contentType, http.StatusCreated, expectedNonExistent},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/password/request", ts.URL),
			contentType: tc.contentType,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()

		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		token := strings.Trim(string(body), "\n")

		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, token, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, token))
	}
}

func TestPasswordReset(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	reqData := struct {
		Token    string `json:"token,omitempty"`
		Password string `json:"password,omitempty"`
		ConfPass string `json:"confirmPassword,omitempty"`
	}{}

	expectedSuccess := toJSON(struct {
		Msg string `json:"msg"`
	}{
		"",
	})

	svc.Register(context.Background(), user)
	tok, _ := token.Generate(user.Email, 0)

	reqData.Password = user.Password
	reqData.ConfPass = user.Password
	reqData.Token = tok
	dataResExisting := toJSON(reqData)

	cases := []struct {
		desc        string
		req         string
		contentType string
		status      int
		res         string
		tok         string
	}{
		{"password reset with valid token and mail", dataResExisting, contentType, http.StatusCreated, expectedSuccess, tok},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/password/reset", ts.URL),
			contentType: tc.contentType,
			body:        strings.NewReader(tc.req),
		}

		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		token := strings.Trim(string(body), "\n")

		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, token, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, token))

	}
}

func TestPasswordReset(t *testing.T) {

	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	reqData := struct {
		Email    string
		Password string
		Token    string
	}{}
	// this probably needs to be moved somwhere else
	expected := struct {
		Msg   string `json:"msg"`
		Error string `json:"error"`
	}{
		"", // it would be good to use actuall error constant but it is private
		"Token expired",
	}

	expectedExpired := toJSON(expected)

	expected.Msg = ""
	expected.Error = users.ErrUserNotFound.Error()
	expectedNonExistent := toJSON(expected)

	expected.Error = ""
	expectedSucces := toJSON(expected)

	svc.Register(context.Background(), user)
	tok, _ := token.Generate(user.Email, 0)
	svc.SaveToken(context.Background(), user.Email, tok)
	tokExp, _ := token.Generate(user.Email, -5)

	reqData.Email = user.Email
	reqData.Password = user.Password
	reqData.Token = tok
	dataResExisting := toJSON(reqData)
	reqData.Token = tokExp
	dataResExpTok := toJSON(reqData)

	reqData.Email = "non-existentuser@example.com"
	dataResNonExisting := toJSON(reqData)

	cases := []struct {
		desc        string
		req         string
		contentType string
		status      int
		res         string
		tok         string
	}{
		{"password reset with valid token and mail", dataResExisting, contentType, http.StatusOK, expectedSucces, tok},
		{"password reset with invalid email", dataResNonExisting, contentType, http.StatusOK, expectedNonExistent, tok},
		{"password reset expired token", dataResExpTok, contentType, http.StatusOK, expectedExpired, tokExp},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/passwd/reset", ts.URL),
			contentType: tc.contentType,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		token := strings.Trim(string(body), "\n")

		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, token, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, token))
	}
}
