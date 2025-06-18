package tests

import (
	"net/http"
	"net/url"
	"short-url/internal/http-server/handlers/url/save"
	"short-url/internal/lib/api"
	"short-url/internal/lib/random"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:9000"
)

func TestURLShortner_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/url").WithJSON(save.Request{
		URL:   gofakeit.URL(),
		Alias: random.NewRandomString(10),
	}).WithBasicAuth("user", "password").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("alias")
}

func TestURLShortner_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alais string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alais: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "123456",
			alais: gofakeit.Word(),
			error: "invalid body,field URL is not in URL format",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			req := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alais,
				}).
				WithBasicAuth("user", "password").
				Expect().Status(http.StatusOK).
				JSON().Object()

			if tc.error != "" {
				req.NotContainsKey("alis")
				req.Value("error").String().IsEqual(tc.error)
				return
			}

			alias := tc.alais

			if tc.alais != "" {
				req.Value("alias").String().IsEqual(tc.alais)
			} else {
				req.Value("alias").String().NotEmpty()
				alias = req.Value("alias").String().Raw()
			}
			testRedirect(t, alias, tc.url)
		})
	}

}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}
	redirectToURL, err := api.GetRedirectURL(u.String())

	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectToURL)

}
