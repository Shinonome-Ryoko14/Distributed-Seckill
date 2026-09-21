package errs

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestNewCarriesBusinessContract(t *testing.T) {
	err := New(CodeInvalidParam, "bad param", http.StatusBadRequest)

	if err == nil {
		t.Fatal("New returned nil")
	}
	if err.Code != CodeInvalidParam {
		t.Errorf("Code = %d, want %d", err.Code, CodeInvalidParam)
	}
	if err.Message != "bad param" {
		t.Errorf("Message = %q, want %q", err.Message, "bad param")
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d, want %d", err.HTTPStatus, http.StatusBadRequest)
	}
	if err.Unwrap() != nil {
		t.Errorf("New should have no cause, got %v", err.Unwrap())
	}
}

func TestErrorIncludesCauseForLogs(t *testing.T) {
	base := errors.New("dial tcp 127.0.0.1:3306: connection refused")
	err := Wrap(base, CodeInternal, "query failed", http.StatusInternalServerError)

	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("Error() = %q, want it to contain the message", err.Error())
	}
	if !strings.Contains(err.Error(), "dial tcp 127.0.0.1:3306") {
		t.Errorf("Error() = %q, want it to contain the cause for logs", err.Error())
	}
}

func TestWrapKeepsCauseChain(t *testing.T) {
	base := errors.New("boom")
	err := Wrap(base, CodeInternal, "wrap", http.StatusInternalServerError)

	if err.Unwrap() != base {
		t.Errorf("Unwrap() = %v, want the original cause", err.Unwrap())
	}
	if !errors.Is(err, base) {
		t.Error("errors.Is(wrapped, base) = false, want true")
	}
}

func TestFromPassesThroughAppError(t *testing.T) {
	orig := New(CodeUserNotFound, "no such user", http.StatusNotFound)

	got := From(orig)
	if got == nil {
		t.Fatal("From(AppError) returned nil")
	}
	if got.Code != orig.Code || got.Message != orig.Message || got.HTTPStatus != orig.HTTPStatus {
		t.Errorf("From(AppError) = %+v, want to preserve %+v", got, orig)
	}
}

func TestFromWrapsUnknownAsInternal(t *testing.T) {
	base := errors.New("sql: SELECT password FROM users")
	got := From(base)

	if got == nil {
		t.Fatal("From(plain error) returned nil")
	}
	if got.Code != CodeInternal {
		t.Errorf("Code = %d, want %d", got.Code, CodeInternal)
	}
	if got.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("HTTPStatus = %d, want %d", got.HTTPStatus, http.StatusInternalServerError)
	}
	if got.Message != "internal server error" {
		t.Errorf("Message = %q, want %q", got.Message, "internal server error")
	}
	if !errors.Is(got, base) {
		t.Error("From should keep the original cause reachable via errors.Is")
	}
}

func TestFromNilIsNil(t *testing.T) {
	if got := From(nil); got != nil {
		t.Errorf("From(nil) = %v, want nil", got)
	}
}
