package antennaApp

import (
	"fmt"
	"testing"
)

// TestParse tests the Parse method that returns a TextFragment struct and error
func TestParse(t *testing.T) {
	testLinks := map[string]string{
		// a text fragment using "link-to-text.github.io"
		"link-to-text.github.io": `https://link-to-text.github.io/#:~:text=A%20Text%20Fragment%20(scroll-to-text)%20link%20lets%20you%20create%20URLs%20that%20automatically%20scroll%20to%20and%20highlight%20specific%20text%20on%20a%20web%20page%20using%20the%20%23%3A~%3Atext%3D%20syntax.%20Supported%20by%20Chrome%20and%20other%20Chromium-based%20browsers%2C%20this%20feature%20is%20perfect%20for%20sharing%20quotes%2C%20emphasizing%20content%2C%20or%20linking%20directly%20to%20important%20passages.`,
		// a text fragment created with Firefox's "copy link highlight"
		"firefox": `https://link-to-text.github.io/#:~:text=A%20Text%20Fragment%20%28scroll%2Dto%2Dtext%29%20link%20lets%20you%20create%20URLs%20that%20automatically%20scroll%20to%20and%20highlight%20specific%20text%20on%20a%20web%20page%20using%20the`,
		// a text fragment created with Safari's "copy link to highlight"
		"safari": `https://link-to-text.github.io/#:~:text=link?-,A%20Text%20Fragment%20(scroll,directly%20to%20important%20passages.,-Use`,
	}
	// NOTE: With Firefox I could not copy the whole paragraph. Not sure if this this to limit the URL length or
	// was that content would have cause issues in interpretation.
	// I was trying to highlight this: `A Text Fragment (scroll-to-text) link lets you create URLs that automatically scroll to and highlight specific text on a web page using the #:~:text= syntax. Supported by Chrome and other Chromium-based browsers, this feature is perfect for sharing quotes, emphasizing content, or linking directly to important passages. `
	for _, link := range testLinks {
		tf, err := ParseTextFragmentURL(link)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("DEBUG tf.Text -> %s\ntf.Link -> %s\n", tf.Text, tf.Link)
	}
}
