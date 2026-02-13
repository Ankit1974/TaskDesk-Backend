// Package middleware provides Gin middleware functions for authentication
// and authorization. It validates Supabase JWTs and enforces role-based
// access control (RBAC) on protected routes.
//
// Middleware chain example:
//
//	auth := api.Group("")
//	auth.Use(middleware.AuthMiddleware())   // verifies JWT, loads user
//	pm := auth.Group("")
//	pm.Use(middleware.RequireRole("PM"))    // restricts to PM role only
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Ankit1974/TaskDeskBackend/internal/config"
	"github.com/Ankit1974/TaskDeskBackend/internal/db"
	"github.com/Ankit1974/TaskDeskBackend/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// UserContext holds the authenticated user's information extracted from the
// JWT token and the registrations database table. It is stored in the Gin
// context by AuthMiddleware and retrieved by handlers via GetUser().
type UserContext struct {
	SupabaseUserID string // Supabase auth.users UUID (from JWT "sub" claim)
	Email          string // User's email address (from JWT "email" claim)
	RegistrationID string // Primary key from the registrations table
	Role           string // User's role from registrations (e.g., "PM", "Developer")
}

// UserContextKey is the key used to store/retrieve UserContext in the Gin context.
const UserContextKey = "user"

// AuthMiddleware validates the Supabase JWT from the Authorization header
// and loads the user's registration data from the database.
//
// Flow:
//  1. Extract "Bearer <token>" from the Authorization header
//  2. Parse and validate the JWT using the Supabase JWT secret (HMAC-SHA256)
//  3. Extract the "email" claim from the token
//  4. Query the registrations table to get the user's ID and role
//  5. Store the UserContext in Gin's context for downstream handlers
//
// Returns 401 Unauthorized if any step fails.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Extract the Bearer token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Step 2: Parse and validate the JWT using Supabase's signing secret
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token uses HMAC signing (Supabase uses HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.Cfg.SupabaseJWTSecret), nil
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

		supabaseUserID, _ := claims["sub"].(string)   // Supabase user UUID
		email, _ := claims["email"].(string)           // User email from Supabase auth

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
