package database

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsUniqueViolation(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		constraintName string
		expected      bool
	}{
		{
			name: "correct error code and constraint name",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "users_email_key",
			},
			constraintName: "users_email_key",
			expected:       true,
		},
		{
			name: "correct error code but wrong constraint name",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "users_email_key",
			},
			constraintName: "users_username_key",
			expected:       false,
		},
		{
			name: "wrong error code",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: "users_email_key",
			},
			constraintName: "users_email_key",
			expected:       false,
		},
		{
			name:           "nil error",
			err:            nil,
			constraintName: "users_email_key",
			expected:       false,
		},
		{
			name:           "non-postgres error",
			err:            errors.New("some error"),
			constraintName: "users_email_key",
			expected:       false,
		},
		{
			name: "empty constraint name in error",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "",
			},
			constraintName: "users_email_key",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUniqueViolation(tt.err, tt.constraintName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		constraintName string
		expected       bool
	}{
		{
			name: "correct error code and constraint name",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: "comments_post_id_fkey",
			},
			constraintName: "comments_post_id_fkey",
			expected:       true,
		},
		{
			name: "correct error code but wrong constraint name",
			err: &pgconn.PgError{
				Code:           "23503",
				ConstraintName: "comments_post_id_fkey",
			},
			constraintName: "comments_user_id_fkey",
			expected:       false,
		},
		{
			name: "wrong error code (unique violation instead)",
			err: &pgconn.PgError{
				Code:           "23505",
				ConstraintName: "comments_post_id_fkey",
			},
			constraintName: "comments_post_id_fkey",
			expected:       false,
		},
		{
			name:           "nil error",
			err:            nil,
			constraintName: "comments_post_id_fkey",
			expected:       false,
		},
		{
			name:           "non-postgres error",
			err:            errors.New("foreign key error"),
			constraintName: "comments_post_id_fkey",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsForeignKeyViolation(tt.err, tt.constraintName)
			assert.Equal(t, tt.expected, result)
		})
	}
}