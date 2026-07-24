/**
 * Facet feature flag evaluation module (Node.js pre-v18 compatible)
 *
 * <p>An independent JS module that simplifies interaction with the Facet backend service.
 * It implements lazy auto-initialization by reading environment variables, requiring no manual setup.
 *
 * <p><b>Usage:</b>
 * 1. Set environment variables: {@code FACET_BASE_URL} and {@code FACET_SDK_KEY}.
 * 2. Call {@code evaluate(...)} directly in your business code.
 *
 * @version 3.0
 * @since Node.js 10+
 */

const http = require('http');
const https = require('https');
const { URL } = require('url');

let _initialized = false;
let _baseUrl;
let _sdkKey;

const _matchedPattern = new RegExp('"matched"\\s*:\\s*(true|false)');

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
function evaluate(featureName, accountId, userId) {
    return new Promise((resolve) => {
        try {
            _ensureInitialized();
        } catch (error) {
            console.error(error.message);
            return resolve(false);
        }

        if (!featureName || featureName.trim() === "") {
            console.error("Error: Feature flag name (featureName) cannot be empty.");
            return resolve(false);
        }

        try {
            const endpoint = new URL(_baseUrl + "/api/v1/evaluate");
            endpoint.searchParams.set("sdk_key", _sdkKey);
            endpoint.searchParams.set("feature_name", featureName);
            if (userId) {
                endpoint.searchParams.set("user_id", userId);
            }
            if (accountId) {
                endpoint.searchParams.set("account_id", accountId);
            }

            const protocol = endpoint.protocol === 'https:' ? https : http;

            const options = {
                method: 'GET',
                timeout: 3000,
                headers: {
                    'Content-Type': 'application/json',
                    'User-Agent': 'Facet-NodeJS-SDK/3.0 (native)'
                }
            };

            const req = protocol.request(endpoint, options, (res) => {
                if (res.statusCode !== 200) {
                    console.error(`Evaluation failed: API returned non-200 status code. Status: ${res.statusCode}, URL: ${endpoint}`);
                    res.resume();
                    return resolve(false);
                }

                let rawData = '';
                res.setEncoding('utf8');
                res.on('data', (chunk) => { rawData += chunk; });
                res.on('end', () => {
                    try {
                        const match = rawData.match(_matchedPattern);
                        if (match && match[1]) {
                            resolve(match[1] === 'true');
                        } else {
                            console.error("Evaluation failed: Response JSON format is incorrect or missing 'data.matched' field.", rawData);
                            resolve(false);
                        }
                    } catch (e) {
                        console.error(`Evaluation failed: Exception while parsing response: ${e.message}`);
                        resolve(false);
                    }
                });
            });

            req.on('error', (e) => {
                console.error(`Network request error during evaluation: ${e.message}`);
                resolve(false);
            });

            req.on('timeout', () => {
                console.error("Evaluation failed: Request timed out.");
                req.destroy();
                resolve(false);
            });

            req.end();

        } catch (error) {
            console.error(`Unknown error during evaluation: ${error.message}`);
            resolve(false);
        }
    });
}

module.exports = { evaluate };


// --- Usage example ---
// (async () => {
//     // Important: before running this example, ensure the following environment variables are set:
//     // export FACET_BASE_URL="http://your-facet-instance.com"
//     // export FACET_SDK_KEY="your-environment-sdk-key"
//
//     const { evaluate } = require('./before_node_18.js');
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
