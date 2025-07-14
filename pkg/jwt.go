package pkg

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mhdiiilham/gosm/entity"
	log "github.com/sirupsen/logrus"
)

// TokenClaims defines the structure for JWT payload claims.
// It includes standard claims such as expiration time, issuer,
// and custom claims like the user ID and email.
type TokenClaims struct {
	jwt.StandardClaims
	ID        int             `json:"user_id"`
	CompanyID int             `json:"company_id"`
	Email     string          `json:"email"`
	Role      entity.UserRole `json:"role"`
}

// JwtGenerator is responsible for generating and validating JWT tokens.
// It holds necessary configurations such as the application's name,
// token expiration duration, signing method, and signature key.
type JwtGenerator struct {
	applicationName string
	signingMethod   *jwt.SigningMethodHMAC
	signatureKey    string
}

// NewJwtGenerator creates and returns a new JwtGenerator instance.
// It initializes the token generator with the application name, token duration,
// and signature key for signing the tokens.
func NewJwtGenerator(
	applicationName string,
	signatureKey string,
) *JwtGenerator {
	return &JwtGenerator{
		applicationName: applicationName,
		signingMethod:   jwt.SigningMethodHS256,
		signatureKey:    signatureKey,
	}
}

// CreateAccessToken generates a JWT token containing the user's ID and email.
// The token is signed using the configured signing method and secret key.
func (g JwtGenerator) CreateAccessToken(userID int, CompanyID int, email string, userRole entity.UserRole) (response *entity.AuthResponse, err error) {
	claims := TokenClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    g.applicationName,
			NotBefore: time.Now().Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		ID:        userID,
		CompanyID: CompanyID,
		Email:     email,
		Role:      userRole,
	}

	token := jwt.NewWithClaims(g.signingMethod, claims)
	signedToken, err := token.SignedString([]byte(g.signatureKey))
	if err != nil {
		log.Warnf("[JwtGenerator.CreateAccessToken] Error signing token: %v", err)
		return nil, entity.ErrInvalidAccessToken
	}

	return &entity.AuthResponse{
		AccessToken: signedToken,
		Email:       email,
		Role:        userRole,
	}, nil
}

// ParseToken verifies and extracts claims from a signed JWT token.
// It validates the token's signature and extracts user-related claims.
func (g JwtGenerator) ParseToken(accessToken string) (*TokenClaims, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		if method, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || method != g.signingMethod {
			log.Warn("[JwtGenerator.ParseToken] Invalid signing method")
			return nil, entity.ErrInvalidAccessToken
		}
		return []byte(g.signatureKey), nil
	})

	if err != nil {
		log.Warnf("[JwtGenerator.ParseToken] Error parsing token: %v", err)
		return nil, entity.ErrAuthTokenIsExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Warn("[JwtGenerator.ParseToken] Failed to parse claims into TokenClaims")
		return nil, entity.ErrInvalidAccessToken
	}

	// fmt.Printf("claims: %+v user_id:%d\n", claims, claims["user_id"].(int))
	email, _ := claims["email"].(string)
	ID, _ := claims["user_id"].(float64)
	companyID, _ := claims["company_id"].(float64)
	userRole, _ := claims["role"].(string)

	return &TokenClaims{
		ID:        int(ID),
		CompanyID: int(companyID),
		Email:     email,
		Role:      entity.UserRole(userRole),
	}, nil
}
