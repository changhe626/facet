//go:build !go1.18

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil" // <-- Keep using ioutil for compatibility with older versions
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	initOnce   sync.Once
	baseUrl    string
	sdkKey     string
	httpClient *http.Client
)

type apiResponse struct {
	Data struct {
		Matched bool `json:"matched"`
	} `json:"data"`
}

func ensureInitialized() {
	initOnce.Do(func() {
		baseUrl = os.Getenv("FACET_BASE_URL")
		sdkKey = os.Getenv("FACET_SDK_KEY")

		if baseUrl == "" {
			panic("error: environment variable FACET_BASE_URL is not set or is empty")
		}
		if sdkKey == "" {
			panic("error: environment variable FACET_SDK_KEY is not set or is empty")
		}

		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	})
}

func evaluate(ctx context.Context, featureName, accountId, userId string) bool {
	ensureInitialized()

	if featureName == "" {
		fmt.Println("error: feature flag name (featureName) cannot be empty.")
		return false
	}

	endpoint, err := url.Parse(baseUrl + "/api/v1/evaluate")
	if err != nil {
		fmt.Printf("error: failed to parse base URL: %v\n", err)
		return false
	}

	queryParams := url.Values{}
	queryParams.Set("sdk_key", sdkKey)
	queryParams.Set("feature_name", featureName)
	if userId != "" {
		queryParams.Set("user_id", userId)
	}
	if accountId != "" {
		queryParams.Set("account_id", accountId)
	}
	endpoint.RawQuery = queryParams.Encode()

	// Compatible with older version request creation method
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		fmt.Printf("error: failed to create HTTP request: %v\n", err)
		return false
	}
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("network error during evaluation: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body) // <-- Use ioutil
		fmt.Printf("evaluation failed: API returned non-200 status code. Status: %d, URL: %s, Body: %s\n",
			resp.StatusCode, endpoint.String(), string(bodyBytes))
		return false
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body) // <-- Use ioutil
	if err != nil {
		fmt.Printf("evaluation failed: failed to read response body: %v\n", err)
		return false
	}

	var res apiResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		fmt.Printf("evaluation failed: failed to parse JSON response: %v, Body: %s\n", err, string(bodyBytes))
		return false
	}

	return res.Data.Matched
}

func main() {
	// Important: Before running this example, make sure to set the following environment variables:
	// export FACET_BASE_URL="http://localhost:8080"
	// export FACET_SDK_KEY="your-environment-sdk-key"

	// Now you can directly call the evaluate method without any initialization steps.
	featureName := "new-dashboard"
	currentUserID := "user-123"
	currentAccountID := "corp-a"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	showNewDashboard := evaluate(ctx, featureName, currentAccountID, currentUserID)

	if showNewDashboard {
		fmt.Printf("evaluation result: true. Showing new dashboard for user %s.\n", currentUserID)
	} else {
		fmt.Printf("evaluation result: false. Showing old dashboard for user %s.\n", currentUserID)
	}
}
