package user

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/health-tracker/server/common"
	"github.com/smalex-als/health-tracker/server/dao"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
	"google.golang.org/appengine/search"
)

type SignInReq struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
}

type SignInResp struct {
	Token string       `json:"token,omitempty"`
	User  *common.User `json:"user,omitempty"`
	common.BaseResp
}

type SignUpReq struct {
	Username string `form:"username" json:"username"`
	Email    string `form:"email" json:"email"`
	Password string `form:"password" json:"password"`
}

type SignUpResp struct {
	Token string       `json:"token,omitempty"`
	User  *common.User `json:"user,omitempty"`
	common.BaseResp
}

type GetSettingsResp struct {
	NumberCalories int `json:"numberCalories"`
	common.BaseResp
}

type PutSettingsReq struct {
	NumberCalories int `json:"numberCalories"`
}

type AuthRemoteService struct {
	common.ServiceRegister
	UserDao *UserDao
}

type UserConfirmKey struct {
	StringID  string    `json:"stringID" datastore:"-"`
	UserId    int64     `json:"userId,string" datastore:"userId"`
	ExpiresAt time.Time `json:"expiresAt" datastore:"expiresAt"`
}

const SEND_CONFIRMATION = "/v1/users-send-confirmation/"

func (service *AuthRemoteService) Register(r *common.HandlerRegister) {
	service.UserDao = NewUserDao()
	r.AddHandler("GET", "/v1/users-install/", false, false, service.handleInstall)
	r.AddHandler("POST", "/v1/users/signin/", false, false, service.handleSignIn)
	r.AddHandler("POST", "/v1/users/signout/", true, false, service.handleUserSignOut)
	r.AddHandler("POST", "/v1/users/signup/", false, false, service.handleSignUp)
	r.AddHandler("GET", "/v1/users-confirm/", true, false, service.handleConfirm)
	r.AddHandler("GET", SEND_CONFIRMATION, false, false, service.handleSendConfirmation)
	r.AddHandler("POST", SEND_CONFIRMATION, false, false, service.handleSendConfirmation)
	r.AddHandler("GET", "/v1/users-settings/", true, true, service.handleGetSettings)
	r.AddHandler("POST", "/v1/users-settings/", true, true, service.handlePutSettings)
}

func (service *AuthRemoteService) handleGetSettings(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	user := common.UserFromContext(ctx)
	c.JSON(http.StatusOK, &GetSettingsResp{NumberCalories: user.NumberCalories})
	return nil
}

func (service *AuthRemoteService) handlePutSettings(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	user := common.UserFromContext(ctx)
	var json PutSettingsReq
	if err := c.BindJSON(&json); err != nil {
		return common.ErrBadRequest
	}
	user.NumberCalories = json.NumberCalories
	if err := service.UserDao.Put(ctx, user); err != nil {
		return common.AppErrorf(err, "put user failed")
	}
	c.JSON(http.StatusOK, &common.DummyResp{})
	return nil
}

func (service *AuthRemoteService) handleSignUp(c *gin.Context) common.AppError {
	var json SignUpReq
	if err := c.BindJSON(&json); err != nil {
		return common.ErrBadRequest
	}
	user := &common.User{
		Username:       strings.TrimSpace(strings.ToLower(json.Username)),
		Email:          strings.TrimSpace(strings.ToLower(json.Email)),
		Enabled:        true,
		NewPassword:    strings.TrimSpace(json.Password),
		Created:        time.Now(),
		RoleId:         common.ROLE_USER,
		NumberCalories: 1600,
	}
	var err error
	var token string
	ctx := common.GetAppEngineContext(c)
	if err = service.UserDao.Put(ctx, user); err != nil {
		return common.AppErrorf(err, "put user failed")
	} else if token, err = common.AuthCreateToken(c, user); err != nil {
		return common.AppErrorf(err, "creating token failed")
	}
	log.Infof(ctx, "successful registraion %+v", user)
	c.SetCookie("sess", token, 24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, &SignUpResp{User: user, Token: token})
	return nil
}

func (service *AuthRemoteService) handleSignIn(c *gin.Context) common.AppError {
	var err error
	var json SignInReq
	if err = c.BindJSON(&json); err != nil {
		return common.ErrBadRequest
	}
	username := strings.TrimSpace(strings.ToLower(json.Username))
	password := strings.TrimSpace(json.Password)
	var user *common.User
	var token string
	ctx := common.GetAppEngineContext(c)
	user, err = service.UserDao.UserFindByUsernamePassword(ctx, username, password)
	if err != nil {
		return common.AppErrorf(err, "User not found")
	} else if token, err = common.AuthCreateToken(c, user); err != nil {
		return common.AppErrorf(err, "creating token failed")
	}
	c.SetCookie("sess", token, 24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, &SignInResp{User: user, Token: token})
	return nil
}

func (service *AuthRemoteService) handleUserSignOut(c *gin.Context) common.AppError {
	c.SetCookie("sess", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func (service *AuthRemoteService) handleConfirm(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	curuser := common.UserFromContext(ctx)
	var id int64
	var err error
	if curuser != nil && curuser.RoleId == common.ROLE_ADMIN {
		if idStr, ok := c.GetQuery("id"); ok {
			id, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return common.AppErrorf(common.ErrBadRequest, "parsing id failed %+v", err)
			}
		} else {
			return common.AppErrorf(common.ErrBadRequest, "id is not specified")
		}
	} else {
		code, ok := c.GetQuery("code")
		if !ok {
			return common.AppErrorf(common.ErrUserCodeEmpty, "Code is not specified")
		}
		if len(code) != 8 {
			return common.AppErrorf(common.ErrUserCodeNotValid, "Bad code format")
		}
		var userConfirmKey UserConfirmKey
		key := datastore.NewKey(ctx, "UserConfirmKey", code, 0, nil)
		err := datastore.Get(ctx, key, &userConfirmKey)
		if err == datastore.ErrNoSuchEntity {
			return common.ErrUserCodeNotValid
		} else if err != nil {
			return common.AppErrorf(err, "Load from db failed")
		}
		if err := datastore.Delete(ctx, key); err != nil {
			log.Warningf(ctx, "deleting failed %+v %+v", err, key)
		}
		if time.Now().After(userConfirmKey.ExpiresAt) {
			return common.ErrUserCodeExpired
		}
		id = userConfirmKey.UserId
	}
	var user common.User
	err = datastore.Get(ctx, datastore.NewKey(ctx, "User", "", id, nil), &user)
	if err == datastore.ErrNoSuchEntity {
		return common.AppErrorf(common.ErrBadRequest, "User not found")
	} else if err != nil {
		return common.AppErrorf(err, "Load from db failed")
	}
	user.Id = id
	if !user.EmailConfirmed {
		user.EmailConfirmed = true
		if err := service.UserDao.Put(service.adminAuthContext(ctx), &user); err != nil {
			return common.AppErrorf(err, "cannot put user to db")
		}
	}
	// delete old confirmation keys
	q := datastore.NewQuery("UserConfirmKey").Filter("userId=", id)
	a := make([]UserConfirmKey, 0)
	if keys, err := q.GetAll(ctx, &a); err != nil {
		log.Warningf(ctx, "failed query UserConfirmKeys %+v", err)
	} else {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			log.Warningf(ctx, "failed delete UserConfirmKeys %+v", err)
		}
	}
	q = datastore.NewQuery("UserConfirmKey").Filter("expiresAt<", time.Now().Add(-24*time.Hour)).Limit(10)
	if keys, err := q.GetAll(ctx, &a); err != nil {
		log.Warningf(ctx, "failed query UserConfirmKeys %+v", err)
	} else {
		if err := datastore.DeleteMulti(ctx, keys); err != nil {
			log.Warningf(ctx, "failed delete UserConfirmKeys %+v", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"Status": "OK"})
	return nil
}

func (service *AuthRemoteService) adminAuthContext(ctx context.Context) context.Context {
	admin := &common.User{Id: 1, Username: "Admin", RoleId: common.ROLE_ADMIN}
	return common.ContextWithUser(ctx, admin)
}

func (service *AuthRemoteService) handleSendConfirmation(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	curuser := common.UserFromContext(ctx)
	var userId int64
	var username string
	var password string
	var userEmail string
	log.Infof(ctx, "curuser = %+v", curuser)
	if curuser != nil {
		// user request
		userEmail = curuser.Email
		userId = curuser.Id
	} else {
		// admin request
		h := c.Request.Header
		for k, v := range h {
			log.Infof(ctx, "header[%+v] = %+v", k, v)
		}
		username = c.Request.FormValue("username")
		password = c.Request.FormValue("password")
		idStr := c.Request.FormValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return common.AppErrorf(common.ErrBadRequest, "parsing id failed %+v", err)
		}
		var user common.User
		err = datastore.Get(ctx, datastore.NewKey(ctx, "User", "", id, nil), &user)
		if err != nil {
			return common.AppErrorf(err, "loading user from db failed")
		}
		user.Id = id
		userId = id
		userEmail = user.Email
	}

	var userConfirmKey UserConfirmKey
	var name string
	var err error
	for {
		name = common.RandStringBytesMaskImprSrc(8)
		key := datastore.NewKey(ctx, "UserConfirmKey", name, 0, nil)
		err := datastore.Get(ctx, key, &userConfirmKey)
		if err == datastore.ErrNoSuchEntity {
			break
		}
	}
	userConfirmKey.StringID = name
	userConfirmKey.UserId = userId
	userConfirmKey.ExpiresAt = time.Now().Add(15 * time.Minute)
	log.Infof(ctx, "confirmation code %s", name)
	_, err = datastore.Put(ctx, datastore.NewKey(ctx, "UserConfirmKey", name, 0, nil), &userConfirmKey)
	if err != nil {
		return common.AppErrorf(err, "insert to db failed")
	}
	var body string
	if len(username) > 0 && len(password) > 0 {
		body = fmt.Sprintf(confirmRegistrationMessage, username, password, name)
	} else {
		body = fmt.Sprintf(confirmMessage, name)
	}
	msg := &mail.Message{
		Sender:  "Support <support@health-tracker-1366.appspotmail.com>",
		To:      []string{userEmail},
		Subject: "Confirm your registration",
		Body:    body,
	}
	if err := mail.Send(ctx, msg); err != nil {
		return common.AppErrorf(err, "Couldn't send email")
	}
	c.JSON(http.StatusOK, gin.H{"Status": "OK"})
	return nil
}

const confirmRegistrationMessage = `
Thank you for creating an account!

username: %s
password: %s

Your confirmation is code:

%s

`
const confirmMessage = `
Your confirmation is code:

%s

`

func (service *AuthRemoteService) handleInstall(c *gin.Context) common.AppError {
	curUser := &common.User{Id: 1, Username: "Admin", RoleId: common.ROLE_ADMIN}
	ctx := common.ContextWithUser(appengine.WithContext(c, c.Request), curUser)
	if cnt, _ := datastore.NewQuery("Role").Count(ctx); cnt == 0 {
		roles := []common.Role{
			{1, "User"},
			{2, "Manager"},
			{3, "Admin"},
		}
		keys := make([]*datastore.Key, len(roles))
		for i, role := range roles {
			keys[i] = datastore.NewKey(ctx, "Role", "", role.Id, nil)
		}
		if _, err := datastore.PutMulti(ctx, keys, roles); err != nil {
			panic(err)
		}
	}
	if cnt, _ := datastore.NewQuery("User").Count(ctx); cnt == 0 {
		deleteIndex(ctx, "users")
		userDao := NewUserDao()
		users := []*common.User{
			{Email: "smalex69@gmail.com", Username: "smalex", RoleId: 3, NewPassword: "golang", EmailConfirmed: true, NumberCalories: 1600},
			{Email: "manager@gmail.com", Username: "manager", RoleId: 2, NewPassword: "golang", EmailConfirmed: true, NumberCalories: 1600},
		}
		for _, v := range users {
			err := userDao.Put(ctx, v)
			if err != nil {
				panic(err)
			}
		}
	}
	c.String(200, "OK")
	return nil
}

type UserRemoteService struct {
	dao.RemoteService
	UserDao *UserDao
}

func (service *UserRemoteService) Register(r *common.HandlerRegister) {
	service.UserDao = NewUserDao()
	service.RemoteService = dao.RemoteService{
		TypeDesc: dao.NewTypeDesc("types/User.json"),
		Entity:   &common.User{},
		Dao:      service.UserDao,
	}
	r.AddHandler("GET", "/v1/users/:id", true, true, service.HandleGet)
	r.AddHandler("GET", "/v1/users/", true, true, atLeastManager(service.HandleList))
	r.AddHandler("GET", "/v1/users-form/", true, true, atLeastManager(service.HandleForm))
	r.AddHandler("POST", "/v1/users/", true, true, atLeastManager(service.HandlePut))
	r.AddHandler("DELETE", "/v1/users/:id", true, true, atLeastManager(service.HandleDelete))
}

type handler func(*gin.Context) common.AppError

func atLeastManager(fn handler) handler {
	return func(c *gin.Context) common.AppError {
		ctx := common.GetAppEngineContext(c)
		user := common.UserFromContext(ctx)
		if user.RoleId >= common.ROLE_MANAGER {
			return fn(c)
		}
		return common.ErrPermissionDenied
	}
}

func deleteIndex(ctx context.Context, indexName string) {
	index, err := search.Open(indexName)
	if err != nil {
		panic(err)
	}
	for t := index.Search(ctx, "", nil); ; {
		id, err := t.Next(nil)
		if err == search.Done {
			break
		}
		if err != nil {
			panic(err)
			break
		}
		index.Delete(ctx, id)
	}
}
