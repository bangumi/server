package person

import (
	"html"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/utiltype"
)

func TestConvertDao_UnescapesHTMLStrings(t *testing.T) {
	t.Parallel()

	wikiValue := `[nickname|"First Hassan"] Tom & Jerry <tag> 'quote'`
	nameValue := `Tom & Jerry <The "Best">`

	person := convertDao(&dao.Person{
		Name:    mustScanHTMLString(t, html.EscapeString(nameValue)),
		Infobox: mustScanHTMLString(t, html.EscapeString(wikiValue)),
	})

	require.Equal(t, nameValue, person.Name)
	require.Equal(t, wikiValue, person.Infobox)
}

func mustScanHTMLString(t *testing.T, value string) utiltype.HTMLEscapedString {
	t.Helper()

	var result utiltype.HTMLEscapedString
	require.NoError(t, (&result).Scan(value))

	return result
}
