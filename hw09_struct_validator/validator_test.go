package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Numbers struct {
		Values []int `validate:"min:0|max:10"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name               string
		in                 interface{}
		expectedProgramErr error
		expectedFieldErrs  map[string]error
	}{
		{
			name: "valid user",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Name:   "John",
				Age:    25,
				Email:  "user@mail.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
		},
		{
			name: "valid app version",
			in:   App{Version: "1.0.0"},
		},
		{
			name: "invalid app version length",
			in:   App{Version: "1.0"},
			expectedFieldErrs: map[string]error{
				"Version": ErrInvalidLength,
			},
		},
		{
			name: "token without validate tags",
			in: Token{
				Header:    []byte("h"),
				Payload:   []byte("p"),
				Signature: []byte("s"),
			},
		},
		{
			name: "valid response code",
			in:   Response{Code: 200, Body: "ok"},
		},
		{
			name: "invalid response code",
			in:   Response{Code: 418, Body: "teapot"},
			expectedFieldErrs: map[string]error{
				"Code": ErrNotInList,
			},
		},
		{
			name: "user with several validation errors",
			in: User{
				ID:     "short",
				Name:   "John",
				Age:    10,
				Email:  "bad-email",
				Role:   "guest",
				Phones: []string{"123", "79001234567"},
			},
			expectedFieldErrs: map[string]error{
				"ID":     ErrInvalidLength,
				"Age":    ErrLessThanMin,
				"Email":  ErrRegexpNotMatch,
				"Role":   ErrNotInList,
				"Phones": ErrInvalidLength,
			},
		},
		{
			name: "user age greater than max",
			in: User{
				ID:     "123e4567-e89b-12d3-a456-426614174000",
				Age:    60,
				Email:  "user@mail.com",
				Role:   "stuff",
				Phones: []string{"79001234567"},
			},
			expectedFieldErrs: map[string]error{
				"Age": ErrGreaterThanMax,
			},
		},
		{
			name: "valid int slice",
			in:   Numbers{Values: []int{0, 5, 10}},
		},
		{
			name: "invalid int slice",
			in:   Numbers{Values: []int{1, 15}},
			expectedFieldErrs: map[string]error{
				"Values": ErrGreaterThanMax,
			},
		},
		{
			name:               "not a struct",
			in:                 "hello",
			expectedProgramErr: ErrNotStruct,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d %s", i, tt.name), func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedProgramErr != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.expectedProgramErr), "actual err: %v", err)
				return
			}

			if len(tt.expectedFieldErrs) == 0 {
				require.NoError(t, err)
				return
			}

			var gotErrors ValidationErrors
			require.True(t, errors.As(err, &gotErrors), "actual err: %v", err)
			require.NotEmpty(t, gotErrors)

			for fieldName, expectedErr := range tt.expectedFieldErrs {
				found := false
				for _, item := range gotErrors {
					if item.Field == fieldName && errors.Is(item.Err, expectedErr) {
						found = true
						break
					}
				}
				require.Truef(t, found, "field %q: want %v, got %v", fieldName, expectedErr, gotErrors)
			}
		})
	}
}
