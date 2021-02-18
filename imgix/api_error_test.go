package imgix

import (
	"errors"
	"testing"
)

func TestErrorImplementsInterface(t *testing.T) {
	var _ error = ApiError{}
}

func TestErrorStringSerializing(t *testing.T) {
	e := ApiError{
		Errors: []struct {
			Detail string `json:"detail"`
			Status string `json:"status"`
			Title  string `json:"title"`
		}{
			{
				Status: "error_1",
				Detail: "error 1",
			},
			{
				Status: "error_2",
				Detail: "error 2",
			},
		},
	}

	res := e.String()
	expected := "status: error_1, details: error 1\nstatus: error_2, details: error 2"

	if res != expected {
		t.Error("invalid error string")
	}
}

func TestIsImgixApiErrorInvalid(t *testing.T) {
	e := errors.New("not_imgix_error")
	is := isImgixApiErr(e, "not_imgix_error")

	if is == true {
		t.Error("not_imgix_error is not imgix error")
	}
}

func TestIsImgixApiErrorValidTitle(t *testing.T) {
	e := ApiError{
		Errors: []struct {
			Detail string `json:"detail"`
			Status string `json:"status"`
			Title  string `json:"title"`
		}{
			{
				Title: "example_imgix_api_err",
			},
		},
	}

	is := isImgixApiErr(e, "example_imgix_api_err")
	if is == false {
		t.Error("example_imgix_api_err is an api error")
	}
}

func TestIsImgixApiErrorInvalidTitle(t *testing.T) {
	e := ApiError{
		Errors: []struct {
			Detail string `json:"detail"`
			Status string `json:"status"`
			Title  string `json:"title"`
		}{
			{
				Title: "example_imgix_api_err",
			},
		},
	}

	is := isImgixApiErr(e, "invalid_error")
	if is == true {
		t.Error("invalid_error is not an api error")
	}
}
