// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
)

const (
	// CTJSON represents JSON content type.
	CTJSON ContentType = "application/json"

	// CTJSONSenML represents JSON SenML content type.
	CTJSONSenML ContentType = "application/senml+json"

	// CTBinary represents binary content type.
	CTBinary ContentType = "application/octet-stream"
)

const minPassLen = 8

var (
	// ErrUnauthorized indicates that entity creation failed.
	ErrUnauthorized = errors.New("unauthorized, missing credentials")

	// ErrFailedCreation indicates that entity creation failed.
	ErrFailedCreation = errors.New("failed to create entity")

	// ErrFailedUpdate indicates that entity update failed.
	ErrFailedUpdate = errors.New("failed to update entity")

	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.New("failed to fetch entity")

	// ErrFailedRemoval indicates that entity removal failed.
	ErrFailedRemoval = errors.New("failed to remove entity")

	// ErrFailedConnect indicates that connecting thing to channel failed.
	ErrFailedConnect = errors.New("failed to connect thing to channel")

	// ErrFailedDisconnect indicates that disconnecting thing from a channel failed.
	ErrFailedDisconnect = errors.New("failed to disconnect thing from channel")

	// ErrFailedPublish indicates that publishing message failed.
	ErrFailedPublish = errors.New("failed to publish message")

	// ErrFailedRead indicates that read messages failed.
	ErrFailedRead = errors.New("failed to read messages")

	// ErrInvalidContentType indicates that non-existent message content type
	// was passed.
	ErrInvalidContentType = errors.New("Unknown Content Type")

	// ErrFetchVersion indicates that fetching of version failed.
	ErrFetchVersion = errors.New("failed to fetch version")

	// ErrFailedWhitelist failed to whitelist configs
	ErrFailedWhitelist = errors.New("failed to whitelist")

	// ErrCerts indicates error fetching certificates.
	ErrCerts = errors.New("failed to fetch certs data")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.New("failed to remove certificate")
)

// ContentType represents all possible content types.
type ContentType string

var _ SDK = (*mfSDK)(nil)

// User represents mainflux user its credentials.
type User struct {
	Email    string                 `json:"email,omitempty"`
	Password string                 `json:"password,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Thing represents mainflux thing.
type Thing struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Channel represents mainflux channel.
type Channel struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SDK contains Mainflux API.
type SDK interface {
	// CreateUser registers mainflux user.
	CreateUser(user User) error

	// User returns user object.
	User(token string) (User, error)

	// CreateToken receives credentials and returns user token.
	CreateToken(user User) (string, error)

	// UpdateUser updates existing user.
	UpdateUser(user User, token string) error

	// UpdatePassword updates user password.
	UpdatePassword(oldPass, newPass, token string) error

	// CreateThing registers new thing and returns its id.
	CreateThing(thing Thing, token string) (string, error)

	// CreateThings registers new things and returns their ids.
	CreateThings(things []Thing, token string) ([]Thing, error)

	// Things returns page of things.
	Things(token string, offset, limit uint64, name string) (ThingsPage, error)

	// ThingsByChannel returns page of things that are connected to specified
	// channel.
	ThingsByChannel(token, chanID string, offset, limit uint64) (ThingsPage, error)

	// Thing returns thing object by id.
	Thing(id, token string) (Thing, error)

	// UpdateThing updates existing thing.
	UpdateThing(thing Thing, token string) error

	// DeleteThing removes existing thing.
	DeleteThing(id, token string) error

	// Connect bulk connects things to channels specified by id.
	Connect(conns ConnectionIDs, token string) error

	// DisconnectThing disconnect thing from specified channel by id.
	DisconnectThing(thingID, chanID, token string) error

	// CreateChannel creates new channel and returns its id.
	CreateChannel(channel Channel, token string) (string, error)

	// CreateChannels registers new channels and returns their ids.
	CreateChannels(channels []Channel, token string) ([]Channel, error)

	// Channels returns page of channels.
	Channels(token string, offset, limit uint64, name string) (ChannelsPage, error)

	// ChannelsByThing returns page of channels that are connected to specified
	// thing.
	ChannelsByThing(token, thingID string, offset, limit uint64) (ChannelsPage, error)

	// Channel returns channel data by id.
	Channel(id, token string) (Channel, error)

	// UpdateChannel updates existing channel.
	UpdateChannel(channel Channel, token string) error

	// DeleteChannel removes existing channel.
	DeleteChannel(id, token string) error

	// SendMessage send message to specified channel.
	SendMessage(chanID, msg, token string) error

	// ReadMessages read messages of specified channel.
	ReadMessages(chanID, token string) (MessagesPage, error)

	// SetContentType sets message content type.
	SetContentType(ct ContentType) error

	// Version returns used mainflux version.
	Version() (string, error)

	// AddBootstrap add bootstrap configuration
	AddBootstrap(token string, cfg BootstrapConfig) (string, error)

	// View returns Thing Config with given ID belonging to the user identified by the given token.
	ViewBootstrap(token, id string) (BootstrapConfig, error)

	// Update updates editable fields of the provided Config.
	UpdateBootstrap(token string, cfg BootstrapConfig) error

	// Remove removes Config with specified token that belongs to the user identified by the given token.
	RemoveBootstrap(token, id string) error

	// Bootstrap returns Config to the Thing with provided external ID using external key.
	Bootstrap(externalKey, externalID string) (BootstrapConfig, error)

	// Whitelist updates Thing state Config with given ID belonging to the user identified by the given token.
	Whitelist(token string, cfg BootstrapConfig) error

	// Cert issues a certificate for a thing required for mtls.
	Cert(thingID, thingKey, token string) (Cert, error)

	// RemoveCert remove a certificate
	RemoveCert(id, token string) error
}

type mfSDK struct {
	baseURL           string
	readerURL         string
	bootstrapURL      string
	certsURL          string
	readerPrefix      string
	usersPrefix       string
	thingsPrefix      string
	channelsPrefix    string
	httpAdapterPrefix string
	bootstrapPrefix   string
	msgContentType    ContentType
	client            *http.Client
}

// Config contains sdk configuration parameters.
type Config struct {
	BaseURL           string
	ReaderURL         string
	BootstrapURL      string
	CertsURL          string
	ReaderPrefix      string
	UsersPrefix       string
	ThingsPrefix      string
	HTTPAdapterPrefix string
	BootstrapPrefix   string
	MsgContentType    ContentType
	TLSVerification   bool
}

// NewSDK returns new mainflux SDK instance.
func NewSDK(conf Config) SDK {
	return &mfSDK{
		baseURL:           conf.BaseURL,
		readerURL:         conf.ReaderURL,
		bootstrapURL:      conf.BootstrapURL,
		certsURL:          conf.CertsURL,
		readerPrefix:      conf.ReaderPrefix,
		usersPrefix:       conf.UsersPrefix,
		thingsPrefix:      conf.ThingsPrefix,
		httpAdapterPrefix: conf.HTTPAdapterPrefix,
		bootstrapPrefix:   conf.BootstrapPrefix,
		msgContentType:    conf.MsgContentType,
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !conf.TLSVerification,
				},
			},
		},
	}
}

func (sdk mfSDK) sendRequest(req *http.Request, token, contentType string) (*http.Response, error) {
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	return sdk.client.Do(req)
}

func createURL(baseURL, prefix, endpoint string) string {
	if prefix == "" {
		return fmt.Sprintf("%s/%s", baseURL, endpoint)
	}

	return fmt.Sprintf("%s/%s/%s", baseURL, prefix, endpoint)
}
