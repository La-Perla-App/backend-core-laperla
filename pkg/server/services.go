package server

import (
	"connectrpc.com/vanguard"
	"github.com/La-Perla-App/backend-core-laperla/pkg/server/middlewares"
	"github.com/go-chi/chi/v5"
)

type RegisterServiceFn = func() *vanguard.Service

func RegisterServicesHandlers(r chi.Router, registerFns []RegisterServiceFn, gatewayPrefixes []string) error {
	var services []*vanguard.Service
	for _, fn := range registerFns {
		services = append(services, fn())
	}

	m, err := middlewares.ApplyTranscoderMiddlewares(gatewayPrefixes)
	if err != nil {
		return err
	}

	transcoder, err := vanguard.NewTranscoder(
		services,
		vanguard.WithCodec(func(res vanguard.TypeResolver) vanguard.Codec {
			jsonCodec := vanguard.NewJSONCodec(res)
			jsonCodec.MarshalOptions.EmitDefaultValues = true
			jsonCodec.MarshalOptions.EmitUnpopulated = false
			return jsonCodec
		}),
		vanguard.WithDefaultServiceOptions(vanguard.WithRESTUnmarshalOptions(vanguard.RESTUnmarshalOptions{
			DiscardUnknownQueryParams: true,
		})),
	)
	if err != nil {
		return err
	}

	r.Handle("/*", m(transcoder.ServeHTTP))

	return nil
}
