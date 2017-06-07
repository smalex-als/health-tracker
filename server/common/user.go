package common

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const (
	ROLE_USER    = 1
	ROLE_MANAGER = 2
	ROLE_ADMIN   = 3
)

type UserPassword struct {
	Id       int64  `json:"id,string" datastore:"-"`
	Password string `datastore:"password"`
}

type Role struct {
	Id   int64  `json:"id,string" datastore:"-"`
	Name string `json:"name" datastore:"name"`
}

type User struct {
	Id             int64     `json:"id,string" datastore:"-"`
	Email          string    `json:"email" datastore:"email"`
	EmailConfirmed bool      `json:"emailConfirmed" datastore:"emailConfirmed"`
	Username       string    `json:"username" datastore:"username"`
	NewPassword    string    `json:"newPassword,omitempty" datastore:"-"`
	Enabled        bool      `json:"enabled" datastore:"enabled"`
	RoleId         int64     `json:"roleId,string" datastore:"roleId" meccano:"Role"`
	Created        time.Time `json:"created" datastore:"created"`
	LastVisit      time.Time `json:"lastVisit" datastore:"lastVisit"`
	NumberCalories int       `json:"numberCalories" datastore:"numberCalories"`
}

func init() {
	RegisterType("User", &User{})
	RegisterType("Role", &Role{})
}

func getTokenFromRequest(req *http.Request, name string) string {
	var token string

	// Get token from the Authorization header
	// format: Authorization: Bearer
	tokens, ok := req.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token != "" {
		return token
	}

	cookie, err := req.Cookie(name)
	if err != nil {
		return ""
	}
	token, _ = url.QueryUnescape(cookie.Value)
	return token
}

func AuthParseToken(ctx context.Context, req *http.Request) (*User, error) {
	token := getTokenFromRequest(req, "sess")
	if token == "" {
		return nil, nil
	}
	// token, err := c.Cookie("sess")
	// if err != nil && err != http.ErrNoCookie {
	// 	return nil, err
	// } else if err == http.ErrNoCookie || len(token) == 0 {
	// 	return nil, nil
	// }
	tokenParser := TokenParserFromContext(ctx)
	val, err := tokenParser.ParseToken(token)
	if err != nil {
		return nil, err
	}
	key := datastore.NewKey(ctx, "User", "", val.UserId, nil)
	var user User
	err = DbGetCached(ctx, key, &user)
	if err == datastore.ErrNoSuchEntity {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("user %s not found %s", key, err)
	}
	if !user.Enabled {
		return nil, ErrUserDeleted
	}
	// if !user.EmailConfirmed {
	// 	return nil, ErrUserEmailNotConfirmed
	// }
	return &user, nil
}

func AuthCreateToken(c *gin.Context, user *User) (string, error) {
	ctx := GetAppEngineContext(c)
	// update user last visit
	user.LastVisit = time.Now()
	key := datastore.NewKey(ctx, "User", "", user.Id, nil)
	if _, err := DbPutCached(ctx, key, user); err != nil {
		log.Warningf(ctx, "Error: %s", err)
		return "", err
	}
	tokenParser := TokenParserFromContext(ctx)
	token, err := tokenParser.CreateToken(user.Username, user.Id, user.RoleId, 24)
	if err != nil {
		return "", err
	}
	return token, nil
}

func RoleRequired(role int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := GetAppEngineContext(c)
		user := UserFromContext(ctx)
		if user != nil && user.RoleId >= role {
			return
		}
		c.AbortWithStatus(403)
	}
}

const AppEngineContextKey = "appengine"

func GetAppEngineContext(c *gin.Context) context.Context {
	return c.MustGet(AppEngineContextKey).(context.Context)
}

func InitContext(c *gin.Context) {
	ctx := appengine.WithContext(c, c.Request)
	ctx = ContextWithTokenParser(ctx, tokenParser)
	c.Set(AppEngineContextKey, ctx)
	if user, err := AuthParseToken(ctx, c.Request); err != nil {
		log.Warningf(ctx, "auth parse token failed %+v", err)
	} else {
		ctx = ContextWithUser(ctx, user)
		c.Set(AppEngineContextKey, ctx)
	}
}
