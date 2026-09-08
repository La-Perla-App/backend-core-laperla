package interceptors

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api/auth"
	"google.golang.org/protobuf/types/known/structpb"
)

var DefaultInterceptors = connect.WithInterceptors(NewInterceptor())

func NewInterceptor() connect.Interceptor {
	return interceptor{}
}

type interceptor struct{}

type nextWrapperFn func(ctx context.Context) (connect.AnyResponse, connect.StreamingHandlerConn, connect.Spec, error)

func (interceptor) intercept(ctx context.Context, header http.Header, nextWrapper nextWrapperFn) (connect.AnyResponse, connect.StreamingHandlerConn, error) {
	generalParams, _ := api.GeneralParamsFromHeaders(header)
	validateSession(header, generalParams)

	res, conn, spec, err := nextWrapper(ctx)

	appInfo, _ := api.ResponseInfoFromHeaders(header)
	appInfo.ClientId = generalParams.ClientId

	if err != nil || ctx.Err() != nil {
		if err != nil {
			log.Println("handler error: ", err)
		} else if ctx.Err() != nil {
			log.Println("handler context error: ", ctx.Err())
		}
		header.Set("Has-Error", "true")
		appInfo.FillErrorInfo()
	} else {
		appInfo.FillSuccessInfo()
	}
	if err == nil {
		api.SetResponseInfoHeader(appInfo, header)
	} else {
		connectErr, ok := err.(*connect.Error)
		if !ok {
			connectErr = connect.NewError(connect.CodeInternal, err)
		}
		encoded, _ := appInfo.Encode()
		detail, _ := connect.NewErrorDetail(structpb.NewStringValue(spec.Procedure + ":responseInfo:" + encoded))
		connectErr.AddDetail(detail)
		err = connectErr
	}
	api.SetResponseInfoHeader(appInfo, header)

	return res, conn, err
}

func (i interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req)
		}
		res, _, err := i.intercept(ctx, req.Header(), func(ctx context.Context) (connect.AnyResponse, connect.StreamingHandlerConn, connect.Spec, error) {
			spec := req.Spec()
			res, err := next(ctx, req)
			return res, nil, spec, err
		})
		return res, err
	})
}

func (interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	})
}

func (i interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		_, _, err := i.intercept(ctx, conn.RequestHeader(), func(ctx context.Context) (connect.AnyResponse, connect.StreamingHandlerConn, connect.Spec, error) {
			spec := conn.Spec()
			err := next(ctx, conn)
			return nil, conn, spec, err
		})
		return err
	})
}

func validateSession(header http.Header, generalParams api.GeneralParams) {
	if generalParams.SessionToken == "" {
		generalParams.SessionToken, _ = auth.GetTokenFromHeader(header)
	}
	if generalParams.SessionToken != "" {
		sessionData, err := auth.ValidateSessionToken(generalParams.SessionToken)
		if err != nil {
			log.Println("SESSION WARNING: ", err)
		} else {
			generalParams.Session = sessionData
		}
	}
	api.SetGeneralParamsHeader(generalParams, header)
}
