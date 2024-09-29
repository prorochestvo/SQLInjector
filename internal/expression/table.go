package expression

import (
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strings"
	"unicode"
)

func extractTableNameAndColumn(column string) (string, string) {
	t := ""
	c := column
	if strings.Contains(c, "/") {
		s := strings.SplitN(c, "/", 2)
		t = s[0]
		c = s[1]
	}
	return t, c
}

func joinTableNameAndColumn(table, column string, mods *[]qm.QueryMod) string {
	c := "\"" + column + "\""

	// TODO: rethink regarding this block, it is not cover all cases!
	if table != "" {
		t := toSnakeCase(table)
		pluralTable := "\"" + toPluralize(t) + "\""
		c = pluralTable + "." + c
		if mods != nil {
			*mods = append(*mods, qm.InnerJoin(pluralTable+" ON "+pluralTable+".\"id\" = \""+t+"_id\""))
		}
	}

	return c
}

func toPluralize(word string) string {
	if plural, found := irregularNouns[word]; found {
		return plural
	}

	// Regular nouns
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") || strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "sh") {
		return word + "es"
	} else if strings.HasSuffix(word, "y") && !strings.ContainsAny(string(word[len(word)-2]), "aeiou") {
		return word[:len(word)-1] + "ies"
	} else if strings.HasSuffix(word, "f") {
		return word[:len(word)-1] + "ves"
	} else if strings.HasSuffix(word, "fe") {
		return word[:len(word)-2] + "ves"
	} else {
		return word + "s"
	}
}

func toSnakeCase(str string) string {
	result := make([]rune, 0, len(str))
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i != 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// irregularNouns is a map of irregular nouns and their plural forms
// IMPORTANT: This map is not exhaustive and only contains the most common irregular nouns.
var irregularNouns = map[string]string{
	"man":    "men",
	"woman":  "women",
	"child":  "children",
	"tooth":  "teeth",
	"foot":   "feet",
	"person": "people",
	"mouse":  "mice",
	"goose":  "geese",
}
