package gormutil

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

func TestIsUniqueConstraintError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "gorm duplicated key sentinel",
			err:  gorm.ErrDuplicatedKey,
			want: true,
		},
		{
			name: "postgres duplicate key value violates unique constraint",
			err:  errors.New(`ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`),
			want: true,
		},
		{
			name: "postgres unique constraint substring",
			err:  errors.New("some unique constraint violation"),
			want: true,
		},
		{
			name: "sqlite unique constraint failed",
			err:  errors.New("UNIQUE constraint failed: users.email"),
			want: true,
		},
		{
			name: "sqlite foreign key constraint failed",
			err:  errors.New("FOREIGN KEY constraint failed"),
			want: false,
		},
		{
			name: "sqlite check constraint failed",
			err:  errors.New("CHECK constraint failed"),
			want: false,
		},
		{
			name: "sqlite unique failed lowercase",
			err:  errors.New("unique failed"),
			want: true,
		},
		{
			name: "generic sql error",
			err:  errors.New("ERROR: syntax error at or near SELECT"),
			want: false,
		},
		{
			name: "connection error",
			err:  errors.New("connection refused"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUniqueConstraintError(tt.err)
			if got != tt.want {
				t.Fatalf("IsUniqueConstraintError(%q) = %v; want %v", tt.err, got, tt.want)
			}
		})
	}
}
