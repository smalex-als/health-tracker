package common

import (
	"crypto/rsa"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/appengine/log"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

const userKey = "userKey"

const (
	// For simplicity these files are in the same folder as the
	// app binary.
	// You shouldn't do this in production.
	// privKeyPath = "keys/jwtRS256.key"
	// pubKeyPath  = "keys/jwtRS256.key.pub"
	privKeyPath = "keys/app.rsa"
	pubKeyPath  = "keys/app.rsa.pub"
)

var tokenParser TokenParser

func init() {
	if _, err := os.Stat(privKeyPath); os.IsNotExist(err) {
		tokenParser = NewTokenParserString([]byte("helloworld"))
	} else {
		tokenParser = NewTokenParserKeys()
	}
}

const tokenParserKey = "tokenParser"

func ContextWithTokenParser(ctx context.Context, parser TokenParser) context.Context {
	return context.WithValue(ctx, tokenParserKey, parser)
}

func TokenParserFromContext(ctx context.Context) TokenParser {
	parser, ok := ctx.Value(tokenParserKey).(TokenParser)
	if ok {
		return parser
	}
	return nil
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context) *User {
	user, ok := ctx.Value(userKey).(*User)
	if ok {
		return user
	}
	return nil
}

type MyCustomClaims struct {
	Name   string `json:"name"`
	UserId int64  `json:"userId,string"`
	RoleId int64  `json:"roleId,string"`
	jwt.StandardClaims
}

type TokenParser interface {
	CreateToken(name string, userId int64, roleId int64, hours int) (string, error)
	ParseToken(token string) (*MyCustomClaims, error)
}

type TokenParserKeys struct {
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
}

func NewTokenParserKeys() TokenParser {
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		panic(err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		panic(err)
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

	return &TokenParserKeys{verifyKey: verifyKey, signKey: signKey}
}

func (parser *TokenParserKeys) CreateToken(name string, userId int64, roleId int64, hours int) (string, error) {
	// token := jwt.New(jwt.SigningMethodHS256) for keystring
	claims := MyCustomClaims{Name: name, UserId: userId, RoleId: roleId}
	claims.ExpiresAt = time.Now().Add(time.Hour * time.Duration(hours)).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(parser.signKey)
}

func (parser *TokenParserKeys) ParseToken(myToken string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(myToken, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return parser.verifyKey, nil
	})

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

var _ TokenParser = &TokenParserKeys{}

type TokenParserString struct {
	secureWord []byte
}

var _ TokenParser = &TokenParserString{}

func NewTokenParserString(keys []byte) TokenParser {
	return &TokenParserString{keys}
}

func (parser *TokenParserString) CreateToken(name string, userId int64, roleId int64, hours int) (string, error) {
	claims := MyCustomClaims{Name: name, UserId: userId, RoleId: roleId}
	claims.ExpiresAt = time.Now().Add(time.Hour * time.Duration(hours)).Unix()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(parser.secureWord)
}

func (parser *TokenParserString) ParseToken(myToken string) (*MyCustomClaims, error) {
	token, err := jwt.ParseWithClaims(myToken, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return parser.secureWord, nil
	})

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

func CreateToken(name string, userId int64, roleId int64, hours int) (string, error) {
	return tokenParser.CreateToken(name, userId, roleId, hours)
}

type HandlerRegister struct {
	R *gin.Engine
}

func (a *HandlerRegister) AddHandler(method, url string, auth bool, checkEmailConfirmed bool, fn func(*gin.Context) AppError) {
	switch method {
	case "POST":
		a.R.POST(url, newAppHandler(auth, checkEmailConfirmed, fn))
	case "GET":
		a.R.GET(url, newAppHandler(auth, checkEmailConfirmed, fn))
	case "DELETE":
		a.R.DELETE(url, newAppHandler(auth, checkEmailConfirmed, fn))
	case "PUT":
		a.R.PUT(url, newAppHandler(auth, checkEmailConfirmed, fn))
	}
}

func newAppHandler(checkAuth, checkEmailConfirmed bool, fn func(*gin.Context) AppError) gin.HandlerFunc {
	return func(g *gin.Context) {
		InitContext(g)
		ctx := GetAppEngineContext(g)
		if checkAuth {
			u := UserFromContext(ctx)
			if u == nil || (checkEmailConfirmed && !u.EmailConfirmed) {
				g.AbortWithStatus(401)
				return
			}
		}
		if e := fn(g); e != nil { // e is *appError, not os.Error.
			resp := &DummyResp{}
			ctx := GetAppEngineContext(g)
			log.Errorf(ctx, "handler return error = %+v", e)
			resp.Errors = PrintClientErrors(ctx, e)
			g.JSON(e.Code(), resp)
		}
	}
}

type ServiceRegister interface {
	Register(r *HandlerRegister)
}
