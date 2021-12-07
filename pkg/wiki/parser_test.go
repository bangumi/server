package wiki_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"app/pkg/wiki"
)

func TestParseFull(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|官网= https://www.mizukinana.jp
|FanClub= https://fanclub.mizukinana.jp
|Twitter= https://twitter.com/NM_NANAPARTY
}}`)

	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, wiki.Wiki{Type: "Crt", Fields: []wiki.Field{
		{Key: "简体中文名", Value: "水树奈奈"},
		{Key: "官网", Value: "https://www.mizukinana.jp"},
		{Key: "FanClub", Value: "https://fanclub.mizukinana.jp"},
		{Key: "Twitter", Value: "https://twitter.com/NM_NANAPARTY"},
	}}, value) {
		return
	}
}

func TestParseFullError(t *testing.T) {
	t.Parallel()
	_, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|别名={
[第二中文名]
[英文名]
[日文名|近藤奈々 (こんどう なな)]
[纯假名|みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]
}}`)

	assert.ErrorIs(t, err, wiki.ErrWikiSyntax)
}

var expected = wiki.Wiki{
	Type: "Crt",
	Fields: []wiki.Field{
		{Key: "简体中文名", Value: "水树奈奈"},
		{Key: "别名", Array: true, Values: []wiki.Item{
			{Key: "", Value: "第二中文名"},
			{Key: "", Value: "英文名"},
			{Key: "日文名", Value: "近藤奈々 (こんどう なな)"},
			{Key: "纯假名", Value: "みずき なな"},
			{Key: "罗马字", Value: "Mizuki Nana"},
			{Key: "昵称", Value: "奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド"},
			{Key: "其他名义", Value: ""},
		}},
	},
}

func TestParseFullArray(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|别名={
[第二中文名]
[英文名]
[日文名|近藤奈々 (こんどう なな)]
[纯假名|みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]
}
}}`)

	if !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, expected, value) {
		return
	}
}

func TestParseEmptyLine(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈
|别名={


[第二中文名]
[英文名]
[日文名|近藤奈々 (こんどう なな)]

[纯假名|みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]

}
}}`)

	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, expected, value) {
		return
	}
}

func TestParseExtraSpace(t *testing.T) {
	t.Parallel()
	value, err := wiki.Parse(`{{Infobox Crt
|简体中文名= 水树奈奈

| 别名 = {


[第二中文名]
[ 英文名]
[日文名|近藤奈々 (こんどう なな)]

[纯假名 |みずき なな]
[罗马字|Mizuki Nana]
[昵称|奈々ちゃん、奈々さん、奈々様、お奈々、ヘッド]
[其他名义|]

 }
}}`)

	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, expected, value) {
		return
	}
}
