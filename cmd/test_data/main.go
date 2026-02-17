package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const baseURL = "http://localhost:8080/api/v1"

func main() {
	// 1. Load Configuration
	config.LoadConfig()

	cwd, _ := os.Getwd()
	log.Printf("Current Working Directory: %s", cwd)

	if config.Cfg.SupabaseJWTSecret == "" {
		log.Printf("Error: SUPABASE_JWT_SECRET is empty. Config loaded: %+v", config.Cfg)
		log.Printf("Checking for .env file...")
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			log.Println(".env file does NOT exist in CWD")
		} else {
			log.Println(".env file exists in CWD")
		}
		log.Fatal("SUPABASE_JWT_SECRET is not set in .env")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	email := fmt.Sprintf("testuser_%d@example.com", time.Now().Unix())

	// 2. Register User
	log.Println("--- Step 1: Registering User ---")
	regBody := map[string]string{
		"full_name":         "Test User",
		"email":             email,
		"organisation_name": "Test Org",
		"role":              "PM",
	}
	resp, err := makeRequest(client, "POST", "/register", regBody, "")
	if err != nil {
		log.Fatalf("Registration failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Registration failed with status %d: %s", resp.StatusCode, string(body))
	}
	log.Printf("User registered: %s\n", email)

	// 3. Generate JWT
	log.Println("--- Step 2: Generating JWT ---")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   uuid.NewString(), // Random Supabase user ID
		"email": email,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.Cfg.SupabaseJWTSecret))
	if err != nil {
		log.Fatalf("Failed to sign token: %v", err)
	}
	log.Println("JWT generated successfully")

	// 4. Create Project
	log.Println("--- Step 3: Creating Project ---")
	projBody := map[string]interface{}{
		"project_name": "Automated Test Project",
		"description":  "Created by the test script",
		"teams":        []string{"backend", "frontend"},
		"start_date":   time.Now().Format("2006-01-02"),
	}
	resp, err = makeRequest(client, "POST", "/projects", projBody, tokenString)
	if err != nil {
		log.Fatalf("Create Project failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Create Project failed with status %d: %s", resp.StatusCode, string(body))
	}

	var projectResp map[string]interface{}
	if err := json.Unmarshal(body, &projectResp); err != nil {
		log.Fatalf("Failed to parse project response: %v", err)
	}
	projectID := projectResp["id"].(string)
	log.Printf("Project created with ID: %s\n", projectID)

	// 5. Create Bugs
	log.Println("--- Step 4: Creating Bugs ---")
	bugsBody := map[string]interface{}{
		"bugs": []map[string]interface{}{
			{
				"title":       "Login page crash",
				"priority":    "critical",
				"description": "App crashes when clicking login",
				"steps":       []string{"Open app", "Click login"},
			},
			{
				"title":    "Typo in header",
				"priority": "low",
			},
		},
	}
	resp, err = makeRequest(client, "POST", fmt.Sprintf("/projects/%s/bugs", projectID), bugsBody, tokenString)
	if err != nil {
		log.Fatalf("Create Bugs failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Create Bugs failed with status %d: %s", resp.StatusCode, string(body))
	}
	log.Println("Bugs created successfully")

	// 6. Verify Data
	log.Println("--- Step 5: Verifying Data ---")
	resp, err = makeRequest(client, "GET", "/projects", nil, tokenString)
	if err != nil {
		log.Fatalf("Get Projects failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Get Projects failed with status %d: %s", resp.StatusCode, string(body))
	}
	log.Println("Projects retrieved successfully")
	log.Println("Test script completed successfully!")
}

func makeRequest(client *http.Client, method, path string, body interface{}, token string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return client.Do(req)
}
