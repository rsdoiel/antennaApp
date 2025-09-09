package antennaApp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	//"os"
	"regexp"
	"strings"

	// 3rd party packages
	"gopkg.in/yaml.v3"
	"github.com/mmcdole/gofeed"
 	"github.com/yuin/goldmark"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/parser"
    "github.com/yuin/goldmark/renderer/html"
)

// CommonMark holds the structure of front matter and the CommonMark
// text.
type CommonMark struct {
	// FrontMatter holds the object of the parsed FrontMatter if available
	// in the document.
	FrontMatter map[string]interface{} `json:"frontMatter,omitempty" yaml:"frontMatter,omitempty"`
	// Text holds the CommonMark text that comes after any front matter
	Text string `json:"text,omitempty" yaml:"text,omitempty"`
}

// ParseMarkdownLinks parses a Markdown text for links in the format `- [LABEL](URL "DESCRIPTION")`
// and returns a slice of Link structures.
func ParseMarkdownLinks(markdownText string) ([]Link, error) {
	// Regular expression to match Markdown links with optional description
	re := regexp.MustCompile(`-\s+\[([^\]]+)\]\(([^)\s]+)(?:\s+"([^"]+)")?\)`)

	lines := strings.Split(markdownText, "\n")
	var links []Link

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue // Skip lines that don't match the pattern
		}

		label := matches[1]
		url := matches[2]
		description := ""
		if len(matches) > 3 {
			description = matches[3]
		}

		links = append(links, Link{
			Label:       label,
			URL:         url,
			Description: description,
		})
	}

	return links, nil
}

// SplitFrontMatter splits a CommonMark document into FrontMatter and the rest of the content.
// It uses bufio.ScanLines to find the "---" delimiters.
func SplitFrontMatter(src []byte) (map[string]interface{}, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(src))

	var frontMatterLines []string
	var inFrontMatter bool
	var rest []string
	var foundEnd bool

	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if !inFrontMatter {
				inFrontMatter = true
				continue
			} else {
				foundEnd = true
				inFrontMatter = false
				continue
			}
		}
		if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		} else {
			rest = append(rest, line)
		}
	}

	if !foundEnd && inFrontMatter {
		return nil, fmt.Sprintf("%s", src), errors.New("unclosed FrontMatter")
	}

	if len(frontMatterLines) == 0 {
		return nil, strings.Join(rest, "\n"), nil // No FrontMatter
	}

	// Parse FrontMatter as YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(strings.Join(frontMatterLines, "\n")), &result); err != nil {
		return nil, "", err
	}
	// The rest of the document starts after the second "---"
	return result, strings.Join(rest, "\n"), nil
}

// Parse will read a byte slice and populate any FrontMatter found
// and set the remaining text as the Text element of CommonMark structure.
func (doc *CommonMark) Parse(src []byte) error {
	frontMatter, text, err := SplitFrontMatter(src)
	if err != nil {
		return err
	}
	doc.FrontMatter = frontMatter
	doc.Text = text
	return nil
}

// GetLinks process the Text of a CommonMark struct and returns
// a list of Link objects if found.
func (doc *CommonMark) GetLinks() ([]Link, error) {
	return ParseMarkdownLinks(doc.Text)
}

// GetAttributeString returns a string attribute from 
// the front matter the document
func (doc *CommonMark) GetAttributeString(key string, defaultValue string) string {
	if val, ok := doc.FrontMatter[key].(string); ok {
		return val
	}
	return defaultValue
}

func mapToPerson(val map[string]interface{}) *gofeed.Person {
	person := &gofeed.Person{}
	if v, ok := val["name"].(string); ok {
		person.Name = v
	}
	if v, ok := val["email"].(string); ok {
		person.Email = v
	}
	if person.Name != "" || person.Email != "" {
		return person
	}
	return nil
}

// GetPersons returns a list of `*gofeed.Person{}`  
// from the front matter in the document document
func (doc *CommonMark) GetPersons(key string) ([]*gofeed.Person, error) {
	if val, ok := doc.FrontMatter[key].(interface{}); ok {
		postList := []*gofeed.Person{}
		switch val.(type) {
		case string:
			// FIXME: Need to handle a RSS style author string of `something@something (name)`
			person := &gofeed.Person{}
			person.Name = fmt.Sprintf("%s", val.(string))
			postList = append(postList, person)
		case []interface{}:
			for _, v := range val.([]interface{}) {
				person := &gofeed.Person{}
				switch v.(type) {
				case string:
					// FIXME: Need to handle a RSS style author string of `something@something (name)``
					person.Name = fmt.Sprintf("%s", v.(string))
					postList = append(postList, person)
				case map[string]interface{}:
					person = mapToPerson(val.(map[string]interface{}))
					if person != nil {
						postList = append(postList, person)
					}
				}
			}
		case map[string]interface{}:
			person := mapToPerson(val.(map[string]interface{}))
			if person != nil {
				postList = append(postList, person)
			}
		default:
			return nil, fmt.Errorf("unable to parse %q", key)
		}
		return postList, nil
	}
	return nil, fmt.Errorf("no persons found for %q", key)
}

// GetAttributeBool returns a boolean attribute from 
// the front matter the document
func (doc *CommonMark) GetAttributeBool(key string, defaultValue bool) bool{
	if val, ok := doc.FrontMatter[key].(bool); ok {
		return val
	}
	return defaultValue
}

func (doc *CommonMark) ToHTML() (string, error) {
	md := goldmark.New(
          goldmark.WithExtensions(extension.GFM),
          goldmark.WithParserOptions(
              parser.WithAutoHeadingID(),
          ),
          goldmark.WithRendererOptions(
              html.WithHardWraps(),
              html.WithXHTML(),
          ),
      )
	var buf bytes.Buffer
	if err := md.Convert([]byte(doc.Text), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
