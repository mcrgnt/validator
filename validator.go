package validator

import (
	"context"
	"net/http"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

type (
	validator interface {
		Validate() error
	}

	strictMiddlewareFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error)
)

type Config struct{}

func (t *Config) Build() (any, error) {
	pbValidator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &Validator{
		pbValidator: pbValidator,
	}, nil
}

type Validator struct {
	pbValidator protovalidate.Validator
}

func (v *Validator) StrictMiddlewareValidate(f strictMiddlewareFunc, operationID string) strictMiddlewareFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (response any, err error) {
		if err := v.validate(request); err != nil {
			return nil, err
		}

		if vmp, ok := request.(interface{ GetBody() any }); ok {
			if err := v.validate(vmp.GetBody()); err != nil {
				return nil, err
			}
		}

		return f(ctx, w, r, request)
	}
}

func (v *Validator) validate(obj any) error {
	if obj == nil {
		return nil
	}

	if vmp, ok := obj.(validator); ok {
		return vmp.Validate()
	}

	if m, ok := obj.(proto.Message); ok {
		if err := v.pbValidator.Validate(m); err != nil {
			return err
		}
	}

	return nil
}
