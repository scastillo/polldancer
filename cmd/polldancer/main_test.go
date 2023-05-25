package main_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/slack-go/slack"
	// Import the package you want to test
)

// MockHTTPClient is a mock implementation of the HTTPClient interface.
type MockHTTPClient struct{}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	// Create a mock HTTP response with the desired status code, headers, and body
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       ioutil.NopCloser(strings.NewReader(`{"mock": "response"}`)),
	}, nil
}

func (m *MockHTTPClient) Post(url, contentType string, body []byte) (*http.Response, error) {
	// Create a mock HTTP response with the desired status code
	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

// MockSlackClient is a mock implementation of the SlackClient interface.
type MockSlackClient struct{}

func (m *MockSlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	// Return mock values or perform desired assertions
	return "", "", nil
}

func TestMain(t *testing.T) {
	// Save the original HTTPClient and SlackClient instances
	origHTTPClient := main.HTTPClient
	origSlackClient := main.SlackClient

	// Replace the original instances with the mock ones
	main.HTTPClient = &MockHTTPClient{}
	main.SlackClient = &MockSlackClient{}

	// Capture the output of the main goroutine
	output := captureOutput(func() {
		main.Main()
	})

	// Assert the expected output
	expectedOutput := "Polling cancelled\n"
	if output != expectedOutput {
		t.Errorf("unexpected output, expected %q, got %q", expectedOutput, output)
	}

	// Restore the original instances
	main.HTTPClient = origHTTPClient
	main.SlackClient = origSlackClient
}

// Helper function to capture the output of a function
func captureOutput(f func()) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return string(out)
}
