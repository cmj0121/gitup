package blog

import (
	"html/template"
	"os"
	"time"

	"github.com/cmj0121/gitup/config"

	log "github.com/sirupsen/logrus"
)

type Category struct {
	Key string
	Blogs
}

type Summary []*Category

func (summary Summary) Write(conf *config.Config, filepath string) (err error) {
	var tmpl *template.Template

	if tmpl, err = conf.ListTemplate(); err != nil {
		// cannot get the template of List page
		return
	}

	var file *os.File
	file, err = os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  filepath,
			"error": err,
		}).Warn("cannot write HTML")
		return
	}
	defer file.Close()

	err = tmpl.Execute(file, struct {
		*config.Config
		Summary
		Style template.CSS

		UTCNow time.Time
	}{
		Config:  conf,
		Summary: summary,
		Style:   conf.CSS(),

		UTCNow: time.Now().UTC(),
	})
	return
}
