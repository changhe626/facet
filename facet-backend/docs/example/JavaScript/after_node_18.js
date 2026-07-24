/**
 * Facet feature flag evaluation module (Node.js v18+ final)
 *
 * <p>An independent JS module that simplifies interaction with the Facet backend service.
 * It implements lazy auto-initialization by reading environment variables, requiring no manual setup.
 *
 * <p><b>Usage:</b>
 * 1. Set environment variables: {@code FACET_BASE_URL} and {@code FACET_SDK_KEY}.
 * 2. Call {@code evaluate(...)} directly in your business code.
 *
 * @version 3.0
 * @since Node.js v18.0.0
 */

let _initialized = false;
let _baseUrl;
let _sdkKey;

function _ensureInitialized() {
    if (!_initialized) {
        _baseUrl = process.env.FACET_BASE_URL;
        _sdkKey = process.env.FACET_SDK_KEY;

        if (!_baseUrl || _baseUrl.trim() === "") {
            throw new Error("Error: Environment variable FACET_BASE_URL is not set or is empty");
        }
        if (!_sdkKey || _sdkKey.trim() === "") {
            throw new Error("Error: Environment variable FACET_SDK_KEY is not set or is empty");
        }
        
        _initialized = true;
    }
}

/**
 * Core evaluation method: determines whether a feature flag matches for a given user/account.
 * In any exceptional case, this method safely returns false.
 *
 * @param {string} featureName - The unique name of the feature flag.
 * @param {string|null} accountId - The unique identifier of the account the user belongs to.
 * @param {string|null} userId - The unique identifier of the current user.
 * @returns {Promise<boolean>} - A Promise that resolves to the evaluation result (true or false).
 */
async function evaluate(featureName, accountId, userId) {
    try {
        _ensureInitialized();
    } catch (error) {
        console.error(error.message);
        return Promise.resolve(false);
    }

    if (!featureName || featureName.trim() === "") {
        console.error("Error: Feature flag name (featureName) cannot be empty.");
        return Promise.resolve(false);
    }

    const params = new URLSearchParams({
        sdk_key: _sdkKey,
        feature_name: featureName,
    });

    if (userId) {
        params.append('user_id', userId);
    }
    if (accountId) {
        params.append('account_id', accountId);
    }

    const url = `${_baseUrl}/api/v1/evaluate?${params.toString()}`;

    try {
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'User-Agent': 'Facet-NodeJS-SDK/3.0'
            },
            signal: AbortSignal.timeout(3000)
        });

        if (!response.ok) {
            console.error(`Evaluation failed: API returned non-success status code. Status: ${response.status}, URL: ${url}`);
            return false;
        }

        const apiResponse = await response.json();

        if (apiResponse && typeof apiResponse.data === 'object' && typeof apiResponse.data.matched === 'boolean') {
            return apiResponse.data.matched;
        } else {
            console.error("Evaluation failed: Response JSON format is incorrect or missing 'data.matched' field.", apiResponse);
            return false;
        }

    } catch (error) {
        console.error(`Network or unknown error during evaluation:`, error);
        return false;
    }
}

module.exports = { evaluate };

// --- Usage example ---
// (async () => {
//     // Important: before running this example, ensure the following environment variables are set:
//     // export FACET_BASE_URL="http://your-facet-instance.com"
//     // export FACET_SDK_KEY="your-environment-sdk-key"
//
//     // Now you can call the evaluate method directly without any initialization.
//     const featureName = 'new-dashboard';
//     const currentUserId = 'user-123';
//     const currentAccountId = 'corp-a';
//
//     console.log(`Evaluating feature '${featureName}'...`);
//     const showNewDashboard = await evaluate(featureName, currentAccountId, currentUserId);
//
//     if (showNewDashboard) {
//         console.log(`Result: true. Serving new dashboard API data to user ${currentUserId}.`);
//     } else {
//         console.log(`Result: false. Serving old dashboard API data to user ${currentUserId}.`);
//     }
// })();
