/*
This is a textcasting extension for gofeed. It adds the source namespace defined at https://source.scripting.com
*/
package antennaApp

// TextcastingExtension represents a feed extension
// for the Textcasting, see https://textcasting.org based on Dave Winer's source namespace
type TextCastingExtension struct {
	Source       []string `json:"source,omitempty"`
}

// NewExtension creates a new TextcastingExtension
// given the generic extension map for the "dc" prefix.
func NewExtension(extensions map[string][]Extension) *TextcastingExtension {
	tc := &Extension{}
	tc.Source = parseTextArrayExtension("source", extensions)
	return tc
}
