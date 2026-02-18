// Package middleware provides Gin middleware functions for authentication
package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/config"
	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWKS types and cache for Supabase ES256 public key verification
type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

var (
	jwksCache     map[string]*ecdsa.PublicKey
	jwksCacheMu   sync.RWMutex
	jwksCacheTime time.Time
	jwksCacheTTL  = 1 * time.Hour
)

// fetchJWKS fetches and caches the Supabase JWKS public keys
func fetchJWKS() error {
	jwksURL := config.Cfg.SupabaseURL + "/auth/v1/.well-known/jwks.json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "EC" || key.Crv != "P-256" {
			continue
		}
		xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			continue
		}
		pubKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
		keys[key.Kid] = pubKey
	}

	jwksCacheMu.Lock()
	jwksCache = keys
	jwksCacheTime = time.Now()
	jwksCacheMu.Unlock()

	return nil
}

// getPublicKey returns the cached ECDSA public key for the given kid
func getPublicKey(kid string) (*ecdsa.PublicKey, error) {
	jwksCacheMu.RLock()
	needsRefresh := jwksCache == nil || time.Since(jwksCacheTime) > jwksCacheTTL
	if !needsRefresh {
		if key, ok := jwksCache[kid]; ok {
			jwksCacheMu.RUnlock()
			return key, nil
		}
	}
	jwksCacheMu.RUnlock()

	// Fetch fresh keys
	if err := fetchJWKS(); err != nil {
		return nil, err
	}

	jwksCacheMu.RLock()
	defer jwksCacheMu.RUnlock()
	if key, ok := jwksCache[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("key ID %q not found in JWKS", kid)
}

/*
	    UserContext holds the authenticated user's information extracted from the
		JWT token and the registrations database table. It is stored in the Gin
		context by AuthMiddleware and retrieved by handlers via GetUser().
*/
type UserContext struct {
	SupabaseUserID string // Supabase auth.users UUID (from JWT "sub" claim)
	Email          string // User's email address (from JWT "email" claim)
	RegistrationID string // Primary key from the registrations table
	Role           string // User's role from registrations (e.g., "PM", "Developer")
}

// UserContextKey is the key used to store/retrieve UserContext in the Gin context.
const UserContextKey = "user"

/*
	AuthMiddleware validates the Supabase JWT from the Authorization header
	and loads the user's registration data from the database.
*/
/*
	Flow:
	  1. Extract "Bearer <token>" from the Authorization header
	  2. Parse and validate the JWT using the Supabase JWT secret (HMAC-SHA256)
	  3. Extract the "email" claim from the token
	  4. Query the registrations table to get the user's ID and role
	  5. Store the UserContext in Gin's context for downstream handlers
*/

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Extract the Bearer token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Step 2: Parse and validate the JWT (supports both ES256 and HS256)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			switch token.Method.(type) {
			case *jwt.SigningMethodECDSA:
				// ES256: verify using Supabase JWKS public key
				kid, _ := token.Header["kid"].(string)
				if kid == "" {
					return nil, fmt.Errorf("missing kid in token header")
				}
				return getPublicKey(kid)
			case *jwt.SigningMethodHMAC:
				// HS256: verify using shared secret
				return []byte(config.Cfg.SupabaseJWTSecret), nil
			default:
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Step 3: Extract claims from the validated token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		supabaseUserID, _ := claims["sub"].(string) // Supabase user UUID
		email, _ := claims["email"].(string)        // User email from Supabase auth

		if email == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token missing email claim"})
			return
		}

		// Step 4: Look up the user in the registrations table to get their role
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var userCtx UserContext
		userCtx.SupabaseUserID = supabaseUserID
		userCtx.Email = email

		err = db.Pool.QueryRow(ctx,
			`SELECT id, role FROM registrations WHERE email = $1`,
			email,
		).Scan(&userCtx.RegistrationID, &userCtx.Role)

		if err != nil {
			logger.Log.Error("Auth middleware: user not found in registrations: " + err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not registered"})
			return
		}

		// Step 5: Store the authenticated user in Gin's context for handlers to access
		c.Set(UserContextKey, &userCtx)
		c.Next()
	}
}

// RequireRole returns middleware that restricts access to users with one of the specified roles.
// Must be used AFTER AuthMiddleware in the middleware chain.
//
// Example: middleware.RequireRole("PM") — only allows Project Managers.
// Example: middleware.RequireRole("PM", "Admin") — allows PM or Admin.
//
// Returns 403 Forbidden if the user's role does not match any allowed role.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		// Case-insensitive role comparison
		for _, role := range allowedRoles {
			if strings.EqualFold(user.Role, role) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions. Required role: " + strings.Join(allowedRoles, " or "),
		})
	}
}

// GetUser is a helper that extracts the authenticated UserContext from the Gin context.
// Returns nil if the user is not authenticated (AuthMiddleware was not applied or failed).
func GetUser(c *gin.Context) *UserContext {
	val, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	user, ok := val.(*UserContext)
	if !ok {
		return nil
	}
	return user
}
