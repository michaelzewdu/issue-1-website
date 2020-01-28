package web

import (
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"html/template"
	"log"
)

// ParseTemplates is used to refresh the templates from disk.
func (s *Setup) ParseTemplates() error {
	funcMap := template.FuncMap{
		"postStarCount":    postStarCount(),
		"postCommentCount": postCommentCount(),
	}
	temp := template.New("issue1")
	temp.Funcs(funcMap)
	temp, err := temp.ParseGlob(s.TemplatesStoragePath + "/*")
	if err != nil {
		return err
	}
	log.Printf("%s\n", temp.DefinedTemplates())
	s.templates = temp
	return nil
}

func postStarCount() func(*issue1.Post) uint {
	return func(p *issue1.Post) (count uint) {
		if p.Stars == nil {
			return 0
		}
		for _, stars := range p.Stars {
			count += stars
		}
		return count
	}
}

func postCommentCount() func(*issue1.Post) int {
	return func(p *issue1.Post) (count int) {
		if p.CommentsID == nil {
			return 0
		}
		return len(p.CommentsID)
	}
}
