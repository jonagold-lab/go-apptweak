package apptweak

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordSearch(t *testing.T) {
	tests := []struct {
		description   string
		token         string
		urlPath       string
		options       Options
		expectedError error
		responseCode  int
		fixture       string
		firstResult   string
	}{
		{
			description:   "happypath",
			token:         "12345x",
			urlPath:       "/ios/searches.json",
			options:       Options{Term: "micro-learning", Num: 100},
			expectedError: nil,
			responseCode:  200,
			fixture:       "./fixtures/keyword_search.json",
			firstResult:   "Micro-Learning",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			responseFile, err := ioutil.ReadFile(tc.fixture)
			if err != nil {
				t.Fatal(err)
			}

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// Testing if the requested Path matches the expected Path
				assert.Equal(t, r.URL.Path, tc.urlPath, "Url Path should match")

				w.WriteHeader(tc.responseCode)
				w.Write(responseFile)
				w.Header().Set("Content-Type", setContentType(tc.fixture))
			}))

			defer s.Close()
			u, err := url.Parse(s.URL)
			if err != nil {
				log.Fatalln("failed to parse httptest.Server URL:", err)
			}
			hc := &http.Client{}
			hc.Transport = RewriteTransport{URL: u}

			client := NewAuthClient(tc.token, hc)
			resp, err := client.KeywordSearch(tc.options)

			// In case of an expected Error, testing if the returned Error matches the expected Error
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err, tc.description)
				return
			}

			//Handling any unexpected Errors and let test fail
			if err != nil && tc.expectedError == nil {
				t.Errorf("Unexpected Error in %v", tc.description)
			}

			// Testing if Unmarshaling of returned JSON works as expected
			assert.Equal(t, tc.firstResult, resp.AppList[0].Title, tc.description)

			// If case of given options, testing if Options are correctly represented in request params
			if tc.description == "with options" {
				assert.Equal(t, tc.options.Country, resp.MD.Req.Params.Country, tc.description)
				assert.Equal(t, tc.options.Language, resp.MD.Req.Params.Language, tc.description)
				assert.Equal(t, tc.options.Device, resp.MD.Req.Params.Device, tc.description)
			}

		})

	}

}
