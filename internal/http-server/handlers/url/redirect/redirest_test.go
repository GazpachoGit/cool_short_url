package redirect_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"short-url/internal/http-server/handlers/url/redirect"
	"short-url/internal/http-server/handlers/url/save"
	"short-url/internal/lib/api"
	"short-url/internal/lib/logger/handlers/silentlog"
	"short-url/internal/storage"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "123",
			url:   "http:google.com",
		},
		{
			name:      "some db error",
			alias:     "123",
			mockError: errors.New("some error"),
			respError: "failed to get url",
		},
		{
			name:      "no url",
			alias:     "123",
			mockError: storage.ErrURLNotFound,
			respError: "url not found",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			urlGetterMock := redirect.NewMockURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).Return(tc.url, tc.mockError).Once()
			}
			//here using chi becouse there is URL param {alias}
			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(silentlog.NewSilentLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			if tc.url != "" {
				redirectedURL, err := api.GetRedirectURL(ts.URL + "/" + tc.alias)

				require.NoError(t, err)

				require.Equal(t, tc.url, redirectedURL)
			} else {
				body, err := api.GetRedirectResponse(ts.URL + "/" + tc.alias)
				require.NoError(t, err)

				var resp save.Response

				require.NoError(t, json.Unmarshal(body, &resp))

				require.Equal(t, tc.respError, resp.Error)
			}

		})
	}
}
