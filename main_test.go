package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestRemoveScrapedHrefs(t *testing.T) {
	tests := []struct {
		Name   string
		Url    string
		Input  []string
		Expect []string
	}{
		{
			Name:   "Manges to remove all hrefs when leaf is reached",
			Url:    "http://10.0.0.83/.hidden/amcbevgondgcrloowluziypjdh/acbnunauucfplzmaglkvqgswwn/ayuprpftypqspruffmkuucjccv/",
			Input:  []string{"../", "README"},
			Expect: []string{},
		},
		{
			Name:  "Manages to scrape href mid tree",
			Url:   "http://10.0.0.83/.hidden/amcbevgondgcrloowluziypjdh/acbnunauucfplzmaglkvqgswwn/",
			Input: []string{"../", "becskiwlclcuqxshqmxhicouoj/", "ayuprpftypqspruffmkuucjccv/", "README"},
			Expect: []string{
				"http://10.0.0.83/.hidden/amcbevgondgcrloowluziypjdh/acbnunauucfplzmaglkvqgswwn/" + "becskiwlclcuqxshqmxhicouoj/",
				"http://10.0.0.83/.hidden/amcbevgondgcrloowluziypjdh/acbnunauucfplzmaglkvqgswwn/" + "ayuprpftypqspruffmkuucjccv/"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			hrefs := removeUnwantedHrefs(tc.Url, tc.Input)

			if len(tc.Expect) != len(hrefs) {
				t.Errorf("expected len is %d, actual is %d: expected values %s, got %s", len(tc.Expect), len(hrefs), tc.Expect, hrefs)
			}

			for i, href := range hrefs {
				if tc.Expect[i] != href {
					t.Errorf("expected %s, got %s", tc.Expect[i], href)
				}
			}

		})
	}

}

func TestTraversForAnchorTags(t *testing.T) {
	s := `<html>
<head><title>Index of /.hidden/amcbevgondgcrloowluziypjdh/</title></head>
<body bgcolor="white">
<h1>Index of /.hidden/amcbevgondgcrloowluziypjdh/</h1><hr><pre><a href="../">../</a>
<a href="acbnunauucfplzmaglkvqgswwn/">acbnunauucfplzmaglkvqgswwn/</a>                        11-Sep-2001 21:21                   -
<a href="bvwrujeymrvzurvywnjxzlfkwa/">bvwrujeymrvzurvywnjxzlfkwa/</a>                        11-Sep-2001 21:21                   -
<a href="ccevyakvydrjhsvbnwvestcfeb/">ccevyakvydrjhsvbnwvestcfeb/</a>                        11-Sep-2001 21:21                   -
<a href="dephqnhvretuprssiccazdpwyt/">dephqnhvretuprssiccazdpwyt/</a>                        11-Sep-2001 21:21                   -
<a href="eotxvxzbogrepmvuiplzkfjohm/">eotxvxzbogrepmvuiplzkfjohm/</a>                        11-Sep-2001 21:21                   -
<a href="README">README</a>                                             11-Sep-2001 21:21                  34
</pre><hr></body>
</html>
`
	expectHrefs := []string{
		"../",
		"acbnunauucfplzmaglkvqgswwn/",
		"bvwrujeymrvzurvywnjxzlfkwa/",
		"ccevyakvydrjhsvbnwvestcfeb/",
		"dephqnhvretuprssiccazdpwyt/",
		"eotxvxzbogrepmvuiplzkfjohm/",
		"README",
	}

	reader := strings.NewReader(s)
	io.NopCloser(reader)
	anchors, err := traversForHrefs(io.NopCloser(reader))
	if err != nil {
		t.Fatal(err)
	}

	for i, a := range anchors {
		if a != expectHrefs[i] {
			t.Errorf("expected %s, got %s", expectHrefs[i], a)
		}
	}
}

func TestContainsPossibleMd5String(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Expect bool
	}{
		{
			Name:   "Happy day",
			Input:  toMd5("this is a md5 string"),
			Expect: true,
		},
		{
			Name:   "Not md5 string",
			Input:  "this is not a md5 string",
			Expect: false,
		},
		{
			Name:  "Both md5 and normal string WITH whitespace",
			Input: fmt.Sprintf("%s and some other text %s", toMd5("prefix"), toMd5("postfix")),
			Expect: true,
		},
		{
			Name: "Both md5 and normal string WITHOUT whitespace",
			Input: fmt.Sprintf("%sand some other text", toMd5("prefix")),
			Expect: true,
		},
		{
			Name: "A 32 len string with char outside of md5 chars",
			Input: strings.Repeat("z", 32),
			Expect: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			if match := containsPossibleMd5String(tc.Input); tc.Expect != match {
				t.Errorf("expected %t, got %t: input md5 string: %s", tc.Expect, match, tc.Input)
			}
		})
	}

}

func toMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}
