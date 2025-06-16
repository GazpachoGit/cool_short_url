package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"short-url/internal/http-server/handlers/url/save"
	"short-url/internal/lib/logger/handlers/silentlog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "http:google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "http:google.com",
		},
		{
			name:      "Empty url",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field",
		},
		{
			name:      "Invalid url",
			url:       "invalid url text",
			alias:     "some_alias",
			respError: "field URL is a valid field",
		},
		{
			name:      "SaveURL Error",
			url:       "http://google.com",
			alias:     "some_alias",
			respError: "field to add url",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		//for the parallel execution
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			urlSaverMock := mocks.NewURLSaver(t)

			/*
				explanation of the condition:
				if tc.respError == "" || tc.mockError != nil:

				DON'T NEED MOCK:
					respError != "" AND mockError = nil -> some error appeared before the storage call
				NEED MOCK:
					respError == "" - when for sure we call the storage
					mockError = nil - when we want to return err from storage
			*/
			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mockAnythingOfType("string")).Return(int64(1), tc.mockError).Once()
			}

			handler := save.New(silentlog.NewSilentLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))

			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusOK)

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(rr.Body.String()), &resp))

			require.Equal(t, tc.respError, resp.Error)

		})
	}
}
