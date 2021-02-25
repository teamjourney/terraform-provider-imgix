package imgix

import (
	"errors"
	"fmt"
	"strings"
)

type ApiError struct {
	Errors []struct {
		Detail string `json:"detail"`
		Status string `json:"status"`
		Title  string `json:"title"`
	} `json:"errors"`
}

func (er ApiError) Error() string {
	return er.String()
}

func (er ApiError) String() string {
	msg := ""
	for _, e := range er.Errors {
		msg += fmt.Sprintf("status: %s, details: %s\n", e.Status, e.Detail)
	}
	return strings.TrimRight(msg, "\n")
}

func isImgixApiErr(err error, title string) bool {
	var imgixErr ApiError
	if errors.As(err, &imgixErr) {
		for _, k := range imgixErr.Errors {
			if k.Title == title {
				return true
			}
		}
	}

	return false
}
