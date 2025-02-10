package request_test

import (
	"testing"

	"github.com/dhaifley/apigo/internal/request"
	"github.com/dhaifley/apigo/tests/mocks"
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
		args: args{id: mocks.TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: mocks.TestInvalidID},
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
		args: args{name: mocks.TestName},
		want: true,
	}, {
		name: "invalid",
		args: args{name: mocks.TestInvalidName},
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
		args: args{id: mocks.TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: mocks.TestInvalidID},
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

func TestValidTokenID(t *testing.T) {
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
		args: args{id: mocks.TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: mocks.TestInvalidID},
		want: false,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := request.ValidTokenID(tt.args.id); got != tt.want {
				t.Errorf("ValidTokenID() = %v, want %v", got, tt.want)
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
		args: args{id: mocks.TestUUID},
		want: true,
	}, {
		name: "invalid",
		args: args{id: mocks.TestInvalidID},
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
