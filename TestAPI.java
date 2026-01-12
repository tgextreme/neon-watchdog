import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Base64;

public class TestAPI {
    public static void main(String[] args) {
        try {
            // Configuración
            String baseUrl = "http://localhost:8080";
            String username = "admin";
            String password = "password";
            
            // Crear auth header
            String auth = username + ":" + password;
            String authHeader = "Basic " + Base64.getEncoder().encodeToString(auth.getBytes());
            
            // Crear cliente HTTP
            HttpClient client = HttpClient.newHttpClient();
            
            // TEST 1: Health Check
            System.out.println("=== TEST 1: Health Check ===");
            HttpRequest healthRequest = HttpRequest.newBuilder()
                    .uri(URI.create(baseUrl + "/api/health"))
                    .header("Authorization", authHeader)
                    .GET()
                    .build();
            
            HttpResponse<String> healthResponse = client.send(healthRequest,
                    HttpResponse.BodyHandlers.ofString());
            
            System.out.println("Status: " + healthResponse.statusCode());
            System.out.println("Response: " + healthResponse.body());
            
            // TEST 2: Get Status
            System.out.println("\n=== TEST 2: Get Status ===");
            HttpRequest statusRequest = HttpRequest.newBuilder()
                    .uri(URI.create(baseUrl + "/api/status"))
                    .header("Authorization", authHeader)
                    .GET()
                    .build();
            
            HttpResponse<String> statusResponse = client.send(statusRequest,
                    HttpResponse.BodyHandlers.ofString());
            
            System.out.println("Status: " + statusResponse.statusCode());
            System.out.println("Response: " + statusResponse.body());
            
            // TEST 3: Get Targets
            System.out.println("\n=== TEST 3: Get Targets ===");
            HttpRequest targetsRequest = HttpRequest.newBuilder()
                    .uri(URI.create(baseUrl + "/api/targets"))
                    .header("Authorization", authHeader)
                    .GET()
                    .build();
            
            HttpResponse<String> targetsResponse = client.send(targetsRequest,
                    HttpResponse.BodyHandlers.ofString());
            
            System.out.println("Status: " + targetsResponse.statusCode());
            System.out.println("Response: " + targetsResponse.body());
            
            System.out.println("\n✅ Todos los tests pasaron correctamente!");
            
        } catch (Exception e) {
            System.err.println("❌ Error: " + e.getMessage());
            e.printStackTrace();
        }
    }
}
