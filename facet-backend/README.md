# facet-backend

facet-backend is a feature flag management platform.

## Quick Start

This section provides a quick overview of how to use the Facet SDK in different languages to evaluate feature flags.

### Go

#### Go 1.18+

This example uses modern features available in Go 1.18 and later.

**`after_1_18.go`**
```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// FacetClient is a client for interacting with the Facet backend service.
type FacetClient struct {
	baseURL    string
	sdkKey     string
	httpClient *http.Client
}

// ... (struct definitions) ...

func NewFacetClient(baseUrl, sdkKey string, timeout time.Duration) (*FacetClient, error) {
	// ...
	return &FacetClient{
		baseURL: baseUrl,
		sdkKey:  sdkKey,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func (c *FacetClient) Evaluate(ctx context.Context, featureName, userID, accountId string, defaultValue bool) bool {
	// ... (implementation) ...
}

func main() {
	client, err := NewFacetClient("http://localhost:8080", "your-environment-sdk-key", 10*time.Second)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Facet client: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	showNewDashboard := client.Evaluate(ctx, "new-dashboard", "user-123", "corp-a", false)

	if showNewDashboard {
		fmt.Println("Showing new dashboard.")
	} else {
		fmt.Println("Showing old dashboard.")
	}
}
```

#### Go (before 1.18)

This example uses `ioutil` for compatibility with older Go versions.

**`before_1_18.go`**
```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// ... (struct and NewFacetClient definitions are the same) ...

func (c *FacetClient) Evaluate(ctx context.Context, featureName, userID, accountId string, defaultValue bool) bool {
    // ...
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    // ...
}

func main() {
	client, err := NewFacetClient("http://localhost:8080", "your-environment-sdk-key", 10*time.Second)
    // ...
}
```

### Java

#### Java 17+

This example uses modern features like `var` and text blocks available in Java 17.

**`Java17Demo.java`**
```java
import java.net.URI;
import java.net.http.HttpClient;
// ...

public class FacetClient {
    // ...
    public static void main(String[] args) {
        var client = new FacetClient("http://localhost:8080", "your-environment-sdk-key");
        boolean showNewDashboard = client.evaluate("new-dashboard", "user-123", "corp-a", false);
        System.out.println("Decision: " + (showNewDashboard ? "Show new dashboard" : "Show old dashboard"));
    }
}
```

#### Java 11+

This example uses the `HttpClient` introduced in Java 11.

**`Java11Demo.java`**
```java
import java.net.URI;
import java.net.http.HttpClient;
// ...

public class FacetClient {
    // ...
    public static void main(String[] args) {
        FacetClient client = new FacetClient("http://localhost:8080", "your-environment-sdk-key");
        boolean showNewDashboard = client.evaluate("new-dashboard", "user-123", "corp-a", false);
        // ...
    }
}
```

#### Java 8

This example uses the classic `HttpURLConnection` for compatibility with Java 8.

**`Java8Demo.java`**
```java
import java.net.HttpURLConnection;
import java.net.URL;
// ...

public class FacetClient {
    // ...
    public static void main(String[] args) {
        FacetClient client = new FacetClient("http://localhost:8080", "your-environment-sdk-key");
        boolean showNewDashboard = client.evaluate("new-dashboard", "user-123", "corp-a", false);
        // ...
    }
}
```

### JavaScript

#### Node.js (v18+)

This example uses the built-in `fetch` API available in Node.js v18 and later.

**`after_node_18.js`**
```javascript
class FacetClient {
    // ...
}

(async () => {
    const facetClient = new FacetClient('http://localhost:8080', 'your-environment-sdk-key');
    const showNewDashboard = await facetClient.evaluate('new-dashboard', 'user-123', 'corp-a', false);
    if (showNewDashboard) {
        console.log("Serving new dashboard API data.");
    } else {
        console.log("Serving old dashboard API data.");
    }
})();
```

#### Node.js (before v18)

This example uses the built-in `http` and `https` modules for compatibility with older Node.js versions.

**`before_node_18.js`**
```javascript
const http = require('http');
const https = require('https');
const { URL } = require('url');

class FacetClient {
    // ...
}

(async () => {
    const facetClient = new FacetClient('http://localhost:8080', 'your-environment-sdk-key');
    // ...
})();
```

#### Browser (Web)

This example uses the standard `fetch` API available in all modern browsers.

**`web.js`**
```javascript
class FacetClient {
    // ...
}

const facetClient = new FacetClient('http://localhost:8080', 'your-environment-sdk-key');

async function checkNewDashboard() {
    const showNewDashboard = await facetClient.evaluate('new-dashboard', 'user-123', 'corp-a', false);
    if (showNewDashboard) {
        console.log("Rendering new dashboard.");
    } else {
        console.log("Rendering old dashboard.");
    }
}

checkNewDashboard();
```
