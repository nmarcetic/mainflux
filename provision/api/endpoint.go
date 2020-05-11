package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/provision"
)

func doProvision(svc provision.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {

		req := request.(addThingReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		token := req.token

		res, err := svc.Provision(req.Name, token, req.ExternalID, req.ExternalKey)

		if err != nil {
			return provisionRes{Error: err.Error()}, nil
		}

		provisionResponse := provisionRes{
			Things:      res.Things,
			Channels:    res.Channels,
			ClientCert:  res.ClientCert,
			ClientKey:   res.ClientKey,
			CACert:      res.CACert,
			Whitelisted: res.Whitelisted,
		}

		return provisionResponse, nil

	}
}
