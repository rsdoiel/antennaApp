package antennaApp

import (
	//"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

// TextFragment holds the elements of a text fragment link along with useful
// metadata like time the link was created.
type TextFragment struct {
	Text    string    `json:"text,omitempty" yaml:"text,omitempty" xml:"text,omitempty"`
	Link    string    `json:"link,omitempty" yaml:"link,omitempty" yaml:"text,omitempty"`
	Host string `json:"host,omitempty" yaml:"host,omitempty" xml:"host,omitempty"`
	Created time.Time `json:"created,omitemtpy" yaml:"created,omitempty" xml:"created,omitempt"`
}

// String renders the text fragment as a
func (tf *TextFragment) String() string {
	return tf.Link
}

// ParseTextFragmentURL takes a text fragment expressed as a URL and returns a TextFragment struct and error
func ParseTextFragmentURL(s string) (*TextFragment, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	//src, _ := json.MarshalIndent(u, "", "  ") // DEBUG
	//fmt.Printf("DEBUG src: %s\n", src) //DEBUG
	tf := new(TextFragment)
	tf.Link = s
	tf.Text, err = url.QueryUnescape(u.Fragment)
	tf.Host = u.Hostname()
	if strings.Contains(tf.Text, ":~:") {
		parts := strings.SplitN(tf.Text, ":~:", 2)
		fragment := parts[1]
		if strings.Contains(fragment, "=") {
			parts = strings.SplitN(fragment, "=", 2)
			fragment = parts[1]
		}
		tf.Text = fragment
	}
	tf.Created = time.Now()
	return tf, err
}

// Reply takes a text fragment URL and writes a Markdown text from it to
// standard  output.
func (app *AntennaApp) Reply(out io.Writer, cfgName string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected a text fragment URL to parse")
	}
	tf, err := ParseTextFragmentURL(args[0])
	if err != nil {
		return fmt.Errorf("failed to parse, %q, %s", args[0], err)
	}
	fmt.Fprintf(out, "\n\n> %s\n\n([%s](%s), accessed %s)\n\n",
		strings.TrimSpace(tf.Text),
		tf.Host,
		strings.TrimSpace(tf.Link),
		tf.Created.Format("2006-01-02"))
	return nil
}