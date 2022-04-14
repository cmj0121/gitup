package blog

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"testing/iotest"
	"time"
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

func ExampleRender() {
	reader := strings.NewReader(test_markdown)

	blog, _ := New(reader)
	text, _ := blog.RenderHTML()
	os.Stdout.Write(text)
	// Output:
	// <nav>
	//
	// <ul>
	// <li><a href="#the-mock-markdown-post">The mock markdown post</a></li>
	// </ul>
	//
	// </nav>
	//
	// <h1 id="the-mock-markdown-post">The mock markdown post</h1>
	//
	// <blockquote>
	// <p>this is the example blog</p>
	// </blockquote>
	//
	// <p>How to write and test blog</p>
}

func TestBlogsSort(t *testing.T) {
	x := Blog{
		Path:       "x",
		created_at: time.Now(),
	}
	y := Blog{
		Path:       "y",
		created_at: x.created_at.Add(time.Second),
	}

	blogs := Blogs{x, y}
	sort.Sort(blogs)

	if !(blogs[0].Path == "y" && blogs[1].Path == "x") {
		t.Errorf("expect sort to %v: %v", Blogs{y, x}, blogs)
	}
}
