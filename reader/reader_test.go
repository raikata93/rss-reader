package reader

import (
	"bytes"
	"github.com/go-test/deep"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse_WithEmptyUrls(t *testing.T) {
	var exp []RssItem
	actual, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error occurred: %s", err)
	}

	if diff := deep.Equal(exp, actual); diff != nil {
		t.Error(diff)
	}
}

func TestHandler(t *testing.T) {
	expected := []byte(testData)

	req, err := http.NewRequest("GET", buildUrl("/"), nil)
	if err != nil {
		t.Fatal(err)
	}

	res := httptest.NewRecorder()

	handler(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Response code was %v; want 200", res.Code)
	}

	if bytes.Compare(expected, res.Body.Bytes()) != 0 {
		t.Errorf("Response body was '%v'; want '%v'", expected, res.Body)
	}
}
