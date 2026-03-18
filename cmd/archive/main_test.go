//nolint:testpackage
package archive

import (
	"bytes"
	"encoding/json"
	"html"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/utiltype"
)

func TestBuildArchiveRecords_UnescapesHTMLStrings(t *testing.T) {
	t.Parallel()

	wikiValue := `[nickname|"First Hassan"] Tom & Jerry <tag> 'quote'`
	nameValue := `Tom & Jerry <The "Best">`
	nameCNValue := `6' 1"`

	subject := buildSubjectArchiveRecord(&dao.Subject{
		Name:    mustScanHTMLString(t, html.EscapeString(nameValue)),
		NameCN:  mustScanHTMLString(t, html.EscapeString(nameCNValue)),
		Infobox: mustScanHTMLString(t, html.EscapeString(wikiValue)),
	}, nil, nil, "", 0)
	person := buildPersonArchiveRecord(&dao.Person{
		Name:    mustScanHTMLString(t, html.EscapeString(nameValue)),
		Infobox: mustScanHTMLString(t, html.EscapeString(wikiValue)),
	})
	character := buildCharacterArchiveRecord(&dao.Character{
		Name:    mustScanHTMLString(t, html.EscapeString(nameValue)),
		Infobox: mustScanHTMLString(t, html.EscapeString(wikiValue)),
	})

	require.Equal(t, nameValue, subject.Name)
	require.Equal(t, nameCNValue, subject.NameCN)
	require.Equal(t, wikiValue, subject.Infobox)
	require.Equal(t, nameValue, person.Name)
	require.Equal(t, wikiValue, person.Infobox)
	require.Equal(t, nameValue, character.Name)
	require.Equal(t, wikiValue, character.Infobox)

	assertJSONRoundTrip(t, subject, func(decoded Subject) {
		require.Equal(t, nameValue, decoded.Name)
		require.Equal(t, nameCNValue, decoded.NameCN)
		require.Equal(t, wikiValue, decoded.Infobox)
	})
	assertJSONRoundTrip(t, person, func(decoded Person) {
		require.Equal(t, nameValue, decoded.Name)
		require.Equal(t, wikiValue, decoded.Infobox)
	})
	assertJSONRoundTrip(t, character, func(decoded Character) {
		require.Equal(t, nameValue, decoded.Name)
		require.Equal(t, wikiValue, decoded.Infobox)
	})
}

func assertJSONRoundTrip[T any](t *testing.T, value T, assertFn func(decoded T)) {
	t.Helper()

	var buf bytes.Buffer
	encode(&buf, value)

	var decoded T
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assertFn(decoded)
}

func mustScanHTMLString(t *testing.T, value string) utiltype.HTMLEscapedString {
	t.Helper()

	var result utiltype.HTMLEscapedString
	require.NoError(t, (&result).Scan(value))

	return result
}
