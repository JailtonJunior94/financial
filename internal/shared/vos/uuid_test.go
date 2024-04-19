package vos

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestNewUUIDFromString(t *testing.T) {
	type args struct {
		input string
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(vo UUID, err error)
	}{
		{
			name: "must return valid uuid given a valid string",
			args: args{input: "0f4fb1cb-6b5b-4779-9170-9da9fa4ed134"},
			expected: func(vo UUID, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "0f4fb1cb-6b5b-4779-9170-9da9fa4ed134", vo.String())
			},
		},
		{
			name: "should return error when the uuid string is invalid",
			args: args{input: "00000008780-0000-0000-87812331231"},
			expected: func(vo UUID, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "should return an error when the uuid string is null or invalid",
			args: args{input: "00000000-0000-0000-0000-000000000000"},
			expected: func(vo UUID, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidUUID, err)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			vo, err := NewUUIDFromString(scenario.args.input)
			scenario.expected(vo, err)
		})
	}
}

func TestNewFromUUID(t *testing.T) {
	type args struct {
		input uuid.UUID
	}

	input := uuid.New()

	scenarios := []struct {
		name     string
		args     args
		expected func(vo UUID, err error)
	}{
		{
			name: "must return value object valid given a uuid",
			args: args{input: input},
			expected: func(vo UUID, err error) {
				assert.NoError(t, err)
				assert.Equal(t, input.String(), vo.String())
			},
		},
		{
			name: "should return error when uuid is null or empty",
			args: args{input: uuid.Nil},
			expected: func(vo UUID, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidUUID, err)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			vo, err := NewFromUUID(scenario.args.input)
			scenario.expected(vo, err)
		})
	}
}
