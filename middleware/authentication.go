package middleware

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	gcontext "mmr/context"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//TODO: access, refresh tokens
func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//extract jwt token from header
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			fmt.Fprintf(os.Stderr, "No token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		splitToken := strings.Split(tokenString, "Bearer ")
		if len(splitToken) < 2 {
			fmt.Fprintf(os.Stderr, "Invalid bearer token")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tokenString = splitToken[1]

		//parse & validate jwt token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid token: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		//extract id and put in context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := gcontext.WithUserID(r.Context(), strconv.FormatInt(int64(claims["user_id"].(float64)), 10)) //TODO: type assertion and double conversion in the same line is ridiculous
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		} else {
			fmt.Fprintf(os.Stderr, "Invalid token: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}
