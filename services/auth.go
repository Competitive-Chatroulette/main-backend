package services

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	Cerr "mmr/errors"
	"mmr/models"
	"os"
	"time"
)

type tokenPair struct {
	at *tokenDetails
	rt *tokenDetails
}

type tokenDetails struct {
	token  string
	uuid   string
	userID int32
	exp    int64
}

type TokenRepository interface {
	Get(key string) (int32, Cerr.CError)
	Set(key string, val int32, exp time.Duration) Cerr.CError
	Del(key string) Cerr.CError
}

type Auth struct {
	usrRepo   UserRepository
	tokenRepo TokenRepository
}

func NewAuth(usrRepo UserRepository, tokenRepo TokenRepository) *Auth {
	return &Auth{
		usrRepo:   usrRepo,
		tokenRepo: tokenRepo,
	}
}

func (auth *Auth) Register(usr *models.User) (string, string, Cerr.CError) {
	if err := usr.HashPass(usr.Pass); err != nil {
		fmt.Fprintf(os.Stderr, "Can't hash the password: %v\n", err)
		return "", "", Cerr.NewInternal()
	}

	userID, cerr := auth.usrRepo.Create(usr)
	if cerr != nil {
		return "", "", cerr
	}

	tp, cerr := genTP()
	if cerr != nil {
		return "", "", cerr
	}

	if err := storeTP(auth.tokenRepo, userID, tp); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't store token: %v", err)
		return "", "", Cerr.NewInternal()
	}

	return tp.at.token, tp.rt.token, nil
}

func (auth *Auth) Login(usr *models.User) (string, string, Cerr.CError) {
	dbUsr, cerr := auth.usrRepo.FindByEmail(usr.Email)
	if cerr != nil {
		return "", "", cerr
	}

	if err := dbUsr.ValidatePass(usr.Pass); err != nil {
		fmt.Fprintf(os.Stderr, "incorrect password : %v\n", err)
		return "", "", Cerr.NewUnauthorized("password")
	}

	tp, cerr := genTP()
	if cerr != nil {
		return "", "", cerr
	}

	if cerr = storeTP(auth.tokenRepo, dbUsr.Id, tp); cerr != nil {
		return "", "", cerr
	}

	return tp.at.token, tp.rt.token, nil
}

func (auth *Auth) Logout(uuid string) Cerr.CError {
	cerr := auth.tokenRepo.Del(uuid)
	if cerr != nil {
		return cerr
	}

	return nil
}

func (auth *Auth) Refresh(uuid string, userID int32) (string, string, Cerr.CError) {
	//remove refresh token from token repo
	if cerr := auth.tokenRepo.Del(uuid); cerr != nil {
		fmt.Fprintf(os.Stderr, "Couldn't delete token from redis: %v", cerr)
		return "", "", cerr
	}

	//generate new token pair
	tp, cerr := genTP()
	if cerr != nil {
		return "", "", cerr
	}

	//store new token pair in token repo
	if cerr = storeTP(auth.tokenRepo, userID, tp); cerr != nil {
		return "", "", cerr
	}

	return tp.at.token, tp.rt.token, nil
}

func (auth *Auth) GetUserID(uuid string) (int32, Cerr.CError) {
	userID, cerr := auth.tokenRepo.Get(uuid)
	if cerr != nil {
		return -1, cerr
	}

	return userID, nil
}

func genTP() (*tokenPair, Cerr.CError) {
	tp := &tokenPair{}
	at, cerr := genToken(time.Now().Add(time.Minute * 15))
	if cerr != nil {
		return nil, cerr
	}
	tp.at = at

	rt, cerr := genToken(time.Now().Add(time.Hour * 24 * 7))
	if cerr != nil {
		return nil, cerr
	}
	tp.rt = rt

	return tp, nil
}

func genToken(exp time.Time) (*tokenDetails, Cerr.CError) {
	td := &tokenDetails{}
	td.exp = exp.Unix()
	td.uuid = uuid.NewString()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uuid": td.uuid,
		"exp":  td.exp,
	})
	var err error
	td.token, err = token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get signed jwt: %v", err)
		return nil, Cerr.NewInternal()
	}

	return td, nil
}

func storeTP(tokenRepo TokenRepository, userID int32, tp *tokenPair) Cerr.CError {
	at := time.Unix(tp.at.exp, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(tp.rt.exp, 0)
	now := time.Now()

	err := tokenRepo.Set(tp.at.uuid, userID, at.Sub(now))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't store access token in tokenRepo: %v", err)
		return Cerr.NewInternal()
	}
	err = tokenRepo.Set(tp.rt.uuid, userID, rt.Sub(now))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't store refresh token in tokenRepo: %v", err)
		return Cerr.NewInternal()
	}
	return nil
}
