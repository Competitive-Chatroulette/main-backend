package app

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	gcontext "mmr/context"
	"mmr/models"
	"mmr/shared"
	"net/http"
	"os"
	"strings"
)

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	usr := gcontext.GetUser(r.Context())
	at, rt, cerr := a.authSvc.Login(usr)
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token":  at,
		"refresh_token": rt,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	usr := gcontext.GetUser(r.Context())
	at, rt, cerr := a.authSvc.Register(usr)
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token":  at,
		"refresh_token": rt,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	uuid := gcontext.GetUUID(r.Context())
	cerr := a.authSvc.Logout(uuid)
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) refresh(w http.ResponseWriter, r *http.Request) {
	uuid := gcontext.GetUUID(r.Context())
	userID := gcontext.GetUserID(r.Context())
	at, rt, cerr := a.authSvc.Refresh(uuid, userID)
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"access_token":  at,
		"refresh_token": rt,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (a *App) withValidatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//unmarshal user
		var usr models.User
		if err := json.NewDecoder(r.Body).Decode(&usr); err != nil {
			fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		//validate user
		if err := shared.Validate.Struct(usr); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err.(validator.ValidationErrors))
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}
		//put user in context
		ctx := gcontext.WithUser(r.Context(), &usr)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

//withClaims is a middleware that parses and validates jwt, inserts token uuid and userID into request context
func (a *App) withClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//extract jwt from header
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			fmt.Fprintf(os.Stderr, "No token")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		splitToken := strings.Split(tokenString, "Bearer ")
		if len(splitToken) < 2 {
			fmt.Fprintf(os.Stderr, "Invalid bearer token")
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		tokenString = splitToken[1]

		//parse & validate jwt
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid token: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		//extract claims from jwt
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			fmt.Fprintf(os.Stderr, "Invalid token: %v", token)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		//get uuid from claims
		uuid, ok := claims["uuid"].(string)
		if !ok || uuid == "" {
			fmt.Fprintf(os.Stderr, "Couldn't extract uuid from token: %v", token)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		//put uuid in context
		ctx := gcontext.WithUUID(r.Context(), uuid)

		//get userID from tokenRepo storage. this is mainly for checking if access token is valid, i.e. still in storage
		userID, cerr := a.authSvc.GetUserID(uuid)
		if cerr != nil {
			http.Error(w, cerr.Error(), cerr.GetStatusCode())
			return
		}
		//put userID in context
		ctx = gcontext.WithUserID(ctx, userID)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
