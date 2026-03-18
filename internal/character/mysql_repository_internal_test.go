package character

import (
	"html"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/utiltype"
)

func TestConvertDao_UnescapesHTMLStrings(t *testing.T) {
	wikiValue := `[nickname|"First Hassan"] Tom & Jerry <tag> 'quote'`
	nameValue := `Tom & Jerry <The "Best">`

	character := convertDao(&dao.Character{
		Name:    mustScanHTMLString(t, html.EscapeString(nameValue)),
		Infobox: mustScanHTMLString(t, html.EscapeString(wikiValue)),
	})

	require.Equal(t, nameValue, character.Name)
	require.Equal(t, wikiValue, character.Infobox)
}

func mustScanHTMLString(t *testing.T, value string) utiltype.HTMLEscapedString {
	t.Helper()

	var result utiltype.HTMLEscapedString
	require.NoError(t, (&result).Scan(value))

	return result
}
