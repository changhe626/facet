import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Objects;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Facet 功能开关评估工具类 (Java 8+ 最终版)
 *
 * <p>这是一个静态工具类，用于简化与 Facet 后端服务的交互。
 * 它实现了“懒汉式”自动初始化，通过环境变量获取配置，无需手动调用 init()。
 *
 * <p><b>使用方法:</b>
 * 1. 设置环境变量: {@code FACET_BASE_URL} 和 {@code FACET_SDK_KEY}。
 * 2. 在业务代码中，直接调用 {@code FacetUtil.evaluate(...)} 方法进行评估。
 *
 * @version 5.0
 * @since Java 8
 */
public final class FacetUtil {

    private static volatile boolean initialized = false;
    private static String baseUrl;
    private static String sdkKey;

    private static final Pattern MATCHED_PATTERN = Pattern.compile("\"matched\"\\s*:\\s*(true|false)");
    private static final Object LOCK = new Object();

    private FacetUtil() {}

    private static void ensureInitialized() {
        if (!initialized) {
            synchronized (LOCK) {
                if (!initialized) {
                    baseUrl = System.getenv("FACET_BASE_URL");
                    sdkKey = System.getenv("FACET_SDK_KEY");

                    if (baseUrl == null || baseUrl.trim().isEmpty()) {
                        throw new IllegalStateException("环境变量 FACET_BASE_URL 未设置或为空");
                    }
                    if (sdkKey == null || sdkKey.trim().isEmpty()) {
                        throw new IllegalStateException("环境变量 FACET_SDK_KEY 未设置或为空");
                    }
                    
                    initialized = true;
                }
            }
        }
    }

    public static boolean evaluate(String featureName, String accountId, String userId) {
        ensureInitialized();

        if (featureName == null || featureName.trim().isEmpty()) {
            System.err.println("错误: 功能开关名称 (featureName) 不能为空。");
            return false;
        }

        HttpURLConnection connection = null;
        try {
            URL url = new URL(buildUrlString(featureName, userId, accountId));

            connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("GET");
            connection.setConnectTimeout(5000);
            connection.setReadTimeout(3000);
            connection.setDoOutput(false);

            int responseCode = connection.getResponseCode();
            if (responseCode != HttpURLConnection.HTTP_OK) {
                System.err.printf("评估失败: API 返回了非 200 的状态码。状态码: %d, URL: %s%n", responseCode, url);
                return false;
            }

            StringBuilder responseBody = new StringBuilder();
            try (BufferedReader in = new BufferedReader(new InputStreamReader(connection.getInputStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = in.readLine()) != null) {
                    responseBody.append(line);
                }
            }

            Matcher matcher = MATCHED_PATTERN.matcher(responseBody.toString());
            if (matcher.find()) {
                return Boolean.parseBoolean(matcher.group(1));
            } else {
                System.err.printf("评估失败: 响应 JSON 中缺少 'matched' 字段或格式不正确。响应: %s%n", responseBody);
                return false;
            }

        } catch (Exception e) {
            System.err.printf("评估过程中发生未知异常: %s%n", e.getMessage());
            return false;
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    private static String buildUrlString(String featureName, String userId, String accountId) throws IOException {
        StringBuilder urlString = new StringBuilder(baseUrl + "/api/v1/evaluate?");
        urlString.append("sdk_key=").append(URLEncoder.encode(sdkKey, "UTF-8"));
        urlString.append("&feature_name=").append(URLEncoder.encode(featureName, "UTF-8"));

        if (userId != null && !userId.isEmpty()) {
            urlString.append("&user_id=").append(URLEncoder.encode(userId, "UTF-8"));
        }
        if (accountId != null && !accountId.isEmpty()) {
            urlString.append("&account_id=").append(URLEncoder.encode(accountId, "UTF-8"));
        }
        return urlString.toString();
    }

    public static void main(String[] args) {
        // 1. (预先设置环境变量)
        // export FACET_BASE_URL="http://localhost:8080"
        // export FACET_SDK_KEY="your-environment-sdk-key"

        // 2. 在业务代码中，直接调用静态的 evaluate 方法，无需初始化。
        String featureName = "new-dashboard";
        String currentUserId = "user-123";
        String currentAccountId = "corp-a";

        boolean showNewDashboard = FacetUtil.evaluate(featureName, currentAccountId, currentUserId);

        if (showNewDashboard) {
            System.out.println("评估结果: true. 正在为用户 user-123 展示新版仪表盘。");
        } else {
            System.out.println("评估结果: false. 正在为用户 user-123 展示旧版仪表盘。");
        }
    }
}
