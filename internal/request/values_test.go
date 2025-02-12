package request_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/request"
)

const (
	TestKey         = int64(1)
	TestID          = "1"
	TestUUID        = "11223344-5566-7788-9900-aabbccddeeff"
	TestName        = "test"
	TestInvalidID   = "ˆ˜√å¬ˆ∂"
	TestInvalidName = "ˆ˜√å¬ˆ∂"
)

func TestValidAccountID(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "valid",
		args: args{id: TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: TestInvalidID},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := request.ValidAccountID(tt.args.id); got != tt.want {
				t.Errorf("ValidAccountID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidAccountName(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "valid",
		args: args{name: TestName},
		want: true,
	}, {
		name: "invalid",
		args: args{name: TestInvalidName},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := request.ValidAccountName(tt.args.name); got != tt.want {
				t.Errorf("ValidAccountName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidUserID(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "valid",
		args: args{id: TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: TestInvalidID},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := request.ValidUserID(tt.args.id); got != tt.want {
				t.Errorf("ValidUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidResourceID(t *testing.T) {
	t.Parallel()

	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "valid",
		args: args{id: TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: TestInvalidID},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := request.ValidResourceID(tt.args.id); got != tt.want {
				t.Errorf("ValidResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args string
		want bool
	}{{
		name: "valid",
		args: "superuser",
		want: true,
	}, {
		name: "invalid",
		args: "invalid",
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := request.ValidScopes(tt.args); got != tt.want {
				t.Errorf("ValidScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}
