package blog

import (
	"fmt"
	"strings"
	"testing"
	"testing/iotest"
)

var test_markdown = `
	# The mock markdown post #
	> this is the example blog

	How to write and test blog
`

func TestNew(t *testing.T) {
	reader := strings.NewReader(test_markdown)
	if _, err := New(reader); err != nil {
		// cannot create reader
		t.Fatalf("cannot create blog: %v", err)
	}

	io_err := fmt.Errorf("mock io.Error")
	err_reader := iotest.ErrReader(io_err)
	if _, err := New(err_reader); err == nil {
		// cannot create reader
		t.Fatalf("cannot create blog: %v", err)
	} else if err != io_err {
		// incorrect error response
		t.Errorf("get unexpect error %v: %v", io_err, err)
	}
}
