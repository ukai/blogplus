package blogplus

import (
	"bytes"
	"log"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var (
	anchorTagRe = regexp.MustCompile("<a .*?</a>")
	htmlTagRe   = regexp.MustCompile("<.*?>")
)

func isMeaningfulContent(content string) bool {
	return len(content) > 200
}

func IsMeaningfulPost(post Activity) bool {
	// nothing for reshares
	if post.Verb == "share" {
		return false
	}
	return isMeaningfulContent(htmlTagRe.ReplaceAllString(post.Object.Content, ""))
}

type TextAttachmentContext struct {
	VisualAttachments []Attachment
	TextAttachments   []Attachment
}

func extractSubject(post *Activity) {
	lines := strings.Split(post.Object.Content, "<br />")
	for _, line := range lines {
		line = anchorTagRe.ReplaceAllString(line, "")
		if len(line) == 0 {
			continue
		}
		idx := strings.IndexAny(line, "\u3002.")
		if idx > 0 {
			if strings.HasPrefix(line[idx:], "\u3002") {
				idx += len("\u3002")
			} else if line[idx] == '.' {
				idx += len(".")
			} else {
				panic("unexpected character at " + line[idx:])
			}
			post.Object.Subject = line[:idx]
		} else {
			post.Object.Subject = line
		}
		return
	}
	// if it fails to extract the subject, use "title" instead
	post.Object.Subject = post.Title
}

func formAttachments(post *Activity) {
	attachments := post.Object.Attachments
	if len(attachments) == 0 {
		post.FormedAttachment = ""
	} else if len(attachments) == 1 {
		attachment := attachments[0]
		buf := bytes.NewBuffer([]byte{})
		var err error
		switch attachment.ObjectType {
		case "video", "photo":
			err = ImageAttachmentTempl.Execute(buf, attachment)
		case "article":
			err = TextAttachmentTempl.Execute(buf,
				TextAttachmentContext{
					TextAttachments: []Attachment{attachment}})
		default:
			log.Println("unknown attachment type:", attachment.ObjectType)
		}
		if err != nil {
			log.Println("template error:", err)
		}
		post.FormedAttachment = buf.String()
	} else {
		var tc TextAttachmentContext
		for _, attachment := range attachments {
			if attachment.ObjectType == "article" {
				tc.TextAttachments = append(tc.TextAttachments, attachment)
			} else {
				tc.VisualAttachments = append(tc.VisualAttachments, attachment)
			}
		}
		buf := bytes.NewBuffer([]byte{})
		err := TextAttachmentTempl.Execute(buf, tc)
		if err != nil {
			log.Println("template error:", err)
		}
		post.FormedAttachment = buf.String()
	}
}

func processPost(post *Activity, postUrl *url.URL) {
	u := *postUrl
	u.Path = path.Join(u.Path, post.Id)
	post.Permalink = u.String()
	formAttachments(post)
	extractSubject(post)
}
