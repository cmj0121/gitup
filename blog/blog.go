package blog

import (
	"bytes"
	"io"

	log "github.com/sirupsen/logrus"
)

// the blog/post instance
type Blog struct {
	// the raw markdown context
	md []byte
}

// create the blog from the open file
func New(reader io.Reader) (blog *Blog, err error) {
	var buff bytes.Buffer

	if _, err = io.Copy(&buff, reader); err != nil {
		// cannot read and save to buffer
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("cannot read blog")
		return
	}

	blog = &Blog{
		md: buff.Bytes(),
	}

	return
}
