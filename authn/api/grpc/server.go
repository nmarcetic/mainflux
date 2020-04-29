// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	kitot "github.com/go-kit/kit/tracing/opentracing"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	mainflux "github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/authn"
	"github.com/mainflux/mainflux/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ mainflux.AuthNServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	issue    kitgrpc.Handler
	identify kitgrpc.Handler
}

// NewServer returns new AuthnServiceServer instance.
func NewServer(tracer opentracing.Tracer, svc authn.Service) mainflux.AuthNServiceServer {
	return &grpcServer{
		issue: kitgrpc.NewServer(
			kitot.TraceServer(tracer, "issue")(issueEndpoint(svc)),
			decodeIssueRequest,
			encodeIssueResponse,
		),
		identify: kitgrpc.NewServer(
			kitot.TraceServer(tracer, "identify")(identifyEndpoint(svc)),
			decodeIdentifyRequest,
			encodeIdentifyResponse,
		),
	}
}

func (s *grpcServer) Issue(ctx context.Context, req *mainflux.IssueReq) (*mainflux.Token, error) {
	_, res, err := s.issue.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.Token), nil
}

func (s *grpcServer) Identify(ctx context.Context, token *mainflux.Token) (*mainflux.UserID, error) {
	_, res, err := s.identify.ServeGRPC(ctx, token)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*mainflux.UserID), nil
}

func decodeIssueRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.IssueReq)
	return issueReq{issuer: req.GetIssuer(), keyType: req.GetType()}, nil
}

func encodeIssueResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &mainflux.Token{Value: res.id}, encodeError(res.err)
}

func decodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*mainflux.Token)
	return identityReq{token: req.GetValue()}, nil
}

func encodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &mainflux.UserID{Value: res.id}, encodeError(res.err)
}

func encodeError(err error) error {
	switch {
	case errors.Contains(err, nil):
		return nil
	case errors.Contains(err, authn.ErrMalformedEntity):
		return status.Error(codes.InvalidArgument, "received invalid token request")
	case errors.Contains(err, authn.ErrUnauthorizedAccess):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Contains(err, authn.ErrKeyExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
