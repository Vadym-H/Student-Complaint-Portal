package swagger

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SetupSwagger configures and registers Swagger endpoints
func SetupSwagger(r chi.Router) {
	// Swagger JSON endpoint
	r.Get("/swagger/doc.json", swaggerDocHandler)

	// Swagger UI endpoint - serve static HTML
	r.Get("/swagger/index.html", swaggerUIHandler)
	r.Get("/swagger/*", swaggerUIHandler)
}

// swaggerDocHandler serves the Swagger specification as JSON
func swaggerDocHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(getSwaggerSpec())
	if err != nil {
		return
	}
}

// swaggerUIHandler serves the Swagger UI
func swaggerUIHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(swaggerUI))
	if err != nil {
		return
	}
}

// getSwaggerSpec returns the Swagger specification
func getSwaggerSpec() map[string]interface{} {
	return map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       "Student Complaint Portal API",
			"description": "A RESTful API for managing student complaints",
			"version":     "1.0.0",
			"contact": map[string]interface{}{
				"name": "API Support",
			},
		},
		"basePath": "/",
		"schemes":  []string{"http", "https"},
		"securityDefinitions": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type": "apiKey",
				"name": "Authorization",
				"in":   "header",
			},
		},
		"paths": getPaths(),
		"definitions": map[string]interface{}{
			"Complaint": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "string",
					},
					"userId": map[string]interface{}{
						"type": "string",
					},
					"description": map[string]interface{}{
						"type": "string",
					},
					"status": map[string]interface{}{
						"type": "string",
					},
					"createdAt": map[string]interface{}{
						"type":   "string",
						"format": "date-time",
					},
				},
			},
		},
	}
}

// getPaths returns all API paths for the Swagger spec
func getPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/auth/register": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Register a new user",
				"description": "Create a new user account with email and password",
				"tags":        []string{"auth"},
				"parameters": []map[string]interface{}{
					{
						"name":     "body",
						"in":       "body",
						"required": true,
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"email": map[string]interface{}{
									"type": "string",
								},
								"password": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "User registered successfully",
					},
					"400": map[string]interface{}{
						"description": "Bad request",
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
					},
				},
			},
		},
		"/api/auth/login": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Login user",
				"description": "Authenticate user and get JWT token",
				"tags":        []string{"auth"},
				"parameters": []map[string]interface{}{
					{
						"name":     "body",
						"in":       "body",
						"required": true,
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"email": map[string]interface{}{
									"type": "string",
								},
								"password": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Login successful",
					},
					"400": map[string]interface{}{
						"description": "Bad request",
					},
					"401": map[string]interface{}{
						"description": "Unauthorized",
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
					},
				},
			},
		},
		"/api/auth/logout": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Logout user",
				"description": "Logout the current user",
				"tags":        []string{"auth"},
				"security": []map[string]interface{}{
					{
						"BearerAuth": []interface{}{},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Logout successful",
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
					},
				},
			},
		},
		"/health": map[string]interface{}{
			"get": map[string]interface{}{
				"summary":     "Health check",
				"description": "Check if the service is running",
				"tags":        []string{"health"},
				"security": []map[string]interface{}{
					{
						"BearerAuth": []interface{}{},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Service is running",
					},
				},
			},
		},
		"/api/complaints": map[string]interface{}{
			"post": map[string]interface{}{
				"summary":     "Create a new complaint",
				"description": "Submit a new complaint",
				"tags":        []string{"complaints"},
				"security": []map[string]interface{}{
					{
						"BearerAuth": []interface{}{},
					},
				},
				"parameters": []map[string]interface{}{
					{
						"name":     "body",
						"in":       "body",
						"required": true,
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"description": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Complaint created successfully",
					},
					"400": map[string]interface{}{
						"description": "Bad request",
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
					},
				},
			},
			"get": map[string]interface{}{
				"summary":     "Get complaints",
				"description": "Retrieve complaints (students get their own, admins can filter by status or ID)",
				"tags":        []string{"complaints"},
				"security": []map[string]interface{}{
					{
						"BearerAuth": []interface{}{},
					},
				},
				"parameters": []map[string]interface{}{
					{
						"name":        "status",
						"in":          "query",
						"type":        "string",
						"description": "Filter by status (pending, approved, rejected)",
					},
					{
						"name":        "id",
						"in":          "query",
						"type":        "string",
						"description": "Get specific complaint by ID (admin only)",
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of complaints",
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
					},
				},
			},
		},
		"/api/complaints/{id}": map[string]interface{}{
			"put": map[string]interface{}{
				"summary":     "Update complaint status",
				"description": "Update the status of a complaint (admin only)",
				"tags":        []string{"complaints"},
				"security": []map[string]interface{}{
					{
						"BearerAuth": []interface{}{},
					},
				},
				"parameters": []map[string]interface{}{
					{
						"name":        "id",
						"in":          "path",
						"required":    true,
						"type":        "string",
						"description": "Complaint ID",
					},
					{
						"name":     "body",
						"in":       "body",
						"required": true,
						"schema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"status": map[string]interface{}{
									"type": "string",
									"enum": []string{"pending", "approved", "rejected"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Complaint updated successfully",
					},
					"400": map[string]interface{}{
						"description": "Bad request",
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
					},
				},
			},
		},
	}
}

// Health check handles health check requests
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Success 200 {string} string "OK"
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

const swaggerUI = `<!DOCTYPE html>
<html>
  <head>
    <title>Student Complaint Portal API</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
      body{
        margin:0;
        padding:0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='/swagger/doc.json'></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@2.0.0/bundles/redoc.standalone.js"></script>
  </body>
</html>
`
