package validator

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Мок для интерфейса validator (Legacy)
type mockValidatable struct {
	err error
}

func (m *mockValidatable) Validate() error {
	return m.err
}

// Мок для Strict Server Request (имитируем oapi-codegen)
type mockRequest struct {
	Body any
}

func (m mockRequest) GetBody() any {
	return m.Body
}

func TestValidator(t *testing.T) {
	cfg := &Config{}
	res, err := cfg.Build()
	require.NoError(t, err)
	v := res.(*Validator)

	t.Run("Validate Legacy Interface", func(t *testing.T) {
		err := v.validate(&mockValidatable{err: errors.New("fail")})
		assert.EqualError(t, err, "fail")
	})

	t.Run("Validate Proto Message", func(t *testing.T) {
		msg := wrapperspb.String("test")
		err := v.validate(msg)
		assert.NoError(t, err)
	})

	t.Run("StrictMiddleware Success", func(t *testing.T) {
		next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			return "final_res", nil
		}

		handler := v.StrictMiddlewareValidate(next, "opId")

		req := mockRequest{Body: &mockValidatable{err: nil}}
		resp, err := handler(context.Background(), nil, nil, req)

		assert.NoError(t, err)
		assert.Equal(t, "final_res", resp)
	})

	t.Run("StrictMiddleware Body Fail", func(t *testing.T) {
		next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			return "should_not_reach", nil
		}

		handler := v.StrictMiddlewareValidate(next, "opId")

		req := mockRequest{Body: &mockValidatable{err: errors.New("validation_error")}}
		resp, err := handler(context.Background(), nil, nil, req)

		assert.Nil(t, resp)
		assert.EqualError(t, err, "validation_error")
	})
}
