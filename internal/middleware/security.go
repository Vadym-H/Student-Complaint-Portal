package middleware

import "net/http"

// SecurityHeaders adds security-related HTTP headers to responses.
// These headers are essential for production deployments and are NOT added by Azure App Service.
// They protect against common web vulnerabilities like MIME-type sniffing, clickjacking, and XSS attacks.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict-Transport-Security: Enforce HTTPS for all future requests to this domain
		// max-age=31536000 (1 year in seconds), includeSubDomains applies to all subdomains,
		// preload allows browsers to preload the domain as HTTPS-only
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Content-Type-Options: Prevent browsers from MIME-type sniffing (interpreting files as different types)
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: Prevent clickjacking by denying the page from being framed
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection: Enable browser XSS protection (for older browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Control how much referrer information is shared
		// strict-origin-when-cross-origin: Send referrer only when navigating to same origin HTTPS
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Restrict access to sensitive browser APIs
		// Disable features not used by this application (example - adjust based on your needs)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
