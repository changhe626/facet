/**
 * Facet feature flag evaluation module (Browser environment final)
 *
 * <p>An independent JS script for browser environments. It implements lazy auto-initialization
 * by reading the `data-*` attributes from its own `<script>` tag, requiring no manual setup.
 *
 * <p><b>Usage:</b>
 * 1. In your HTML, include this script with the required data attributes:
 *    <pre>{@code <script src="web.js" id="facet-sdk-script" data-base-url="http://api.your-facet.com" data-sdk-key="your-sdk-key"></script>}</pre>
 * 2. In your business code, call the global function {@code facetEvaluate(...)} directly.
 *
 * @version 2.0
 * @since ECMAScript 2017 (ES8) for async/await
 */

(function(window) {
    let _initialized = false;
    let _baseUrl;
    let _sdkKey;

    const SCRIPT_ID = "facet-sdk-script";

    function _ensureInitialized() {
        if (!_initialized) {
            const scriptTag = document.getElementById(SCRIPT_ID);
            if (!scriptTag) {
                throw new Error(`Error: Cannot find script tag with ID "${SCRIPT_ID}". Ensure your script tag includes this ID.`);
            }

            _baseUrl = scriptTag.dataset.baseUrl;
            _sdkKey = scriptTag.dataset.sdkKey;

            if (!_baseUrl || _baseUrl.trim() === "") {
                throw new Error(`Error: Attribute "data-base-url" not found or empty on the <script> tag.`);
            }
            if (!_sdkKey || _sdkKey.trim() === "") {
                throw new Error(`Error: Attribute "data-sdk-key" not found or empty on the <script> tag.`);
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

    window.facetEvaluate = evaluate;

})(window);


// --- Usage example ---

/*
// 1. In your HTML file, include this script:
// <script 
//   src="./web.js" 
//   id="facet-sdk-script" 
//   data-base-url="http://your-facet-instance.com" 
//   data-sdk-key="your-environment-sdk-key"
// ></script>

// 2. In your business logic script, call the global function facetEvaluate.
(async () => {
    const featureName = 'new-dashboard';
    const currentUserId = 'user-123';
    const currentAccountId = 'corp-a';

    console.log(`Evaluating feature '${featureName}'...`);
    const showNewDashboard = await facetEvaluate(featureName, currentAccountId, currentUserId);

    if (showNewDashboard) {
        console.log(`Result: true. Rendering new dashboard for user ${currentUserId}.`);
        // renderNewDashboard();
    } else {
        console.log(`Result: false. Rendering old dashboard for user ${currentUserId}.`);
        // renderOldDashboard();
    }
})();
*/
