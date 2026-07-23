package utils

import (
	"github.com/russross/blackfriday/v2"
)

func Markdown2HTML(md string) string {
	return string(blackfriday.Run([]byte(md), blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{Flags: blackfriday.UseXHTML | blackfriday.CompletePage}))))
}
