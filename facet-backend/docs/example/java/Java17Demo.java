import java.io.IOException;
import java.net.URI;
import java.net.URLEncoder;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.Objects;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Facet 功能开关评估工具类 (Java 17+ 最终版)
 *
 * <p>这是一个静态工具类，用于简化与 Facet 后端服务的交互。
 * 它实现了“懒汉式”自动初始化，通过环境变量获取配置，无需手动调用 init()。
 *
 * <p><b>使用方法:</b>
 * 1. 设置环境变量: {@code FACET_BASE_URL} 和 {@code FACET_SDK_KEY}。
 * 2. 在业务代码中，直接调用 {@code FacetUtil.evaluate(...)} 方法进行评估。
 *
 * @version 5.0
 * @since Java 17
 */
public final class FacetUtil {

    private static volatile boolean initialized = false;
    private static String baseUrl;
    private static String sdkKey;
    private static HttpClient httpClient;

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

                    httpClient = HttpClient.newBuilder()
                            .version(HttpClient.Version.HTTP_2)
                            .connectTimeout(Duration.ofSeconds(5))
                            .build();
                    
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

        try {
            URI uri = new URI(buildUrlString(featureName, userId, accountId));

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(uri)
                    .timeout(Duration.ofSeconds(3))
                    .GET()
                    .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() != 200) {
                System.err.printf(
                    """"\n评估失败: API 返回了非 200 的状态码。\n  - 状态码: %d\n  - URL: %s\n  - 响应体: %s\n"""",
                    response.statusCode(), uri, response.body()
                );
                return false;
            }

            Matcher matcher = MATCHED_PATTERN.matcher(response.body());
            if (matcher.find()) {
                return Boolean.parseBoolean(matcher.group(1));
            } else {
                System.err.printf("评估失败: 响应 JSON 中缺少 'matched' 字段或格式不正确。响应: %s%n", response.body());
                return false;
            }

        } catch (Exception e) {
            System.err.printf("评估过程中发生未知异常: %s%n", e.getMessage());
            return false;
        }
    }

    private static String buildUrlString(String featureName, String userId, String accountId) {
        var urlString = new StringBuilder(baseUrl + "/api/v1/evaluate?");
        urlString.append("sdk_key=").append(URLEncoder.encode(sdkKey, StandardCharsets.UTF_8));
        urlString.append("&feature_name=").append(URLEncoder.encode(featureName, StandardCharsets.UTF_8));

        if (userId != null && !userId.isEmpty()) {
            urlString.append("&user_id=").append(URLEncoder.encode(userId, StandardCharsets.UTF_8));
        }
        if (accountId != null && !accountId.isEmpty()) {
            urlString.append("&account_id=").append(URLEncoder.encode(accountId, StandardCharsets.UTF_8));
        }
        return urlString.toString();
    }

    public static void main(String[] args) {
        // 1. (预先设置环境变量)
        // export FACET_BASE_URL="http://localhost:8080"
        // export FACET_SDK_KEY="your-environment-sdk-key"

        // 2. 在业务代码中，直接调用静态的 evaluate 方法，无需初始化。
        var featureName = "new-dashboard";
        var currentUserId = "user-123";
        var currentAccountId = "corp-a";

        boolean showNewDashboard = FacetUtil.evaluate(featureName, currentAccountId, currentUserId);

        System.out.println(
            String.format(
                """"\n功能: '%s'\n上下文: user_id=%s, account_id=%s\n评估结果: %b\n业务决策: %s\n"""",
                featureName,
                currentUserId,
                currentAccountId,
                showNewDashboard,
                showNewDashboard ? "展示新版仪表盘" : "展示旧版仪表盘"
            )
        );
    }
}
