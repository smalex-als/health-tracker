package user

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/expense-tracker/server/apptest"
	"github.com/smalex-als/health-tracker/server/common"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestAll(t *testing.T) {
	r := gin.New()
	register := &common.HandlerRegister{r}
	(&AuthRemoteService{}).Register(register)
	(&UserRemoteService{}).Register(register)
	http.Handle("/", r)

	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()
	req, _ := inst.NewRequest("GET", "/", nil)
	ctx := appengine.NewContext(req)

	users := []*common.User{
		{Id: 10000, Username: "admin", Enabled: true, EmailConfirmed: true, RoleId: common.ROLE_ADMIN},
		{Id: 20000, Username: "manager", Enabled: true, EmailConfirmed: true, RoleId: common.ROLE_MANAGER},
		{Id: 30000, Username: "smalex", Enabled: true, EmailConfirmed: true, RoleId: common.ROLE_USER},
	}
	keys := make([]*datastore.Key, len(users))
	for i, u := range users {
		keys[i] = datastore.NewKey(ctx, "User", "", u.Id, nil)
	}
	if _, err := datastore.PutMulti(ctx, keys, users); err != nil {
		panic(err)
	}
	roles := []*common.Role{
		{1, "User"},
		{2, "Manager"},
		{3, "Admin"},
	}
	keys = make([]*datastore.Key, len(roles))
	for i, role := range roles {
		keys[i] = datastore.NewKey(ctx, "Role", "", role.Id, nil)
	}
	if _, err := datastore.PutMulti(ctx, keys, roles); err != nil {
		panic(err)
	}

	userAdmin, _ := common.CreateToken("admin", 10000, common.ROLE_ADMIN, 1)
	userManager, _ := common.CreateToken("manager", 20000, common.ROLE_MANAGER, 1)
	// userOrdinary, _ := common.CreateToken("smalex", 30000, common.ROLE_USER)
	// userAdmin2, _ := common.CreateToken("admin2", 40000, common.ROLE_ADMIN
	userNew, _ := common.CreateToken("newsmalex", 1, common.ROLE_USER, 1)

	const signinURL = "/v1/users/signin/"
	const signupURL = "/v1/users/signup/"
	const signoutURL = "/v1/users/signout/"
	const userURL = "/v1/users/"

	commands := []*apptest.TestCommand{
		{
			Name:         "signin without any params",
			Href:         signinURL,
			Code:         400,
			ErrorMessage: common.ErrBadRequest.Message(),
		},
		{
			Name:         "signup without any params",
			Href:         signupURL,
			Code:         400,
			ErrorMessage: common.ErrBadRequest.Message(),
		},
		{
			Name:         "signup without username",
			Href:         signupURL,
			Src:          &SignUpReq{"", "", ""},
			Code:         400,
			ErrorMessage: ErrUsernameIsEmpty.Message(),
		},
		{
			Name:         "signup without email",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "", ""},
			Code:         400,
			ErrorMessage: ErrEmailIsEmpty.Message(),
		},
		{
			Name:         "signup password is too short",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "newsmalex69@gmail.com", ""},
			Code:         400,
			ErrorMessage: ErrPasswordTooShort.Message(),
		},
		{
			Name:         "signup password is too short",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "newsmalex69@gmail.com", "123"},
			Code:         400,
			ErrorMessage: ErrPasswordTooShort.Message(),
		},
		{
			Name:         "signin if user is not exists",
			Href:         signinURL,
			Src:          &SignInReq{"newsmalex", "123123"},
			Code:         404,
			ErrorMessage: common.ErrUserNotFound.Message(),
		},
		{
			Name: "signup successful",
			Href: signupURL,
			Src:  &SignUpReq{"newsmalex", "newsmalex69@gmail.com", "123123"},
			Code: 200,
		},
		{
			Name:         "signup user already exists",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "newsmalex69@gmail.com", "123"},
			Code:         400,
			ErrorMessage: ErrUsernameExists.Message(),
		},
		{
			Name:         "signup user already exists",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "newsmalex69@gmail.com", "123123"},
			Code:         400,
			ErrorMessage: ErrUsernameExists.Message(),
		},
		{
			Name:         "signup user already exists",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "newsmalex69@gmail.com", "123123"},
			Code:         400,
			ErrorMessage: ErrUsernameExists.Message(),
		},
		{
			Name:         "signup user already exists",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex", "newsmalex69@gmail.com", ""},
			Code:         400,
			ErrorMessage: ErrUsernameExists.Message(),
		},
		{
			Name:         "signup email already exists",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex111", "newsmalex69@gmail.com", "1234567"},
			Code:         400,
			ErrorMessage: ErrEmailExists.Message(),
		},
		{
			Name:         "signin password is too short",
			Href:         signinURL,
			Src:          &SignInReq{"smalex", ""},
			Code:         400,
			ErrorMessage: ErrPasswordTooShort.Message(),
		},
		{
			Name:         "signin wrong password",
			Href:         signinURL,
			Src:          &SignInReq{"newsmalex69@gmail.com", "123456"},
			Code:         400,
			ErrorMessage: ErrPasswordWrong.Message(),
		},
		{
			Name:         "signup email is not valid",
			Href:         signupURL,
			Src:          &SignUpReq{"newsmalex69", "newsmalex69gmailcom", "123456"},
			Code:         400,
			ErrorMessage: ErrEmailIsNotValid.Message(),
		},
		{
			Name:         "signup username is not valid",
			Href:         signupURL,
			Src:          &SignUpReq{"abc", "newsmalex69gmailcom", "123456"},
			Code:         400,
			ErrorMessage: ErrUsernameIsNotValid.Message(),
		},
		{
			Name: "successful signin by username",
			Href: signinURL,
			Src:  &SignInReq{"newsmalex", "123123"},
			Code: 200,
		},
		{
			Name:   "before confirm email user cannot get his own profile",
			Method: "GET",
			Href:   userURL + "1",
			Code:   401,
			Token:  userNew,
		},
		{
			Method:       "GET",
			Name:         "unsuccessful confirmation",
			Href:         "/v1/users-confirm/?id=1",
			ErrorMessage: common.ErrBadRequest.Message(),
			Code:         401,
		},
		{
			Name:  "send confirmation by admin",
			Href:  SEND_CONFIRMATION + "?id=1",
			Code:  200,
			Token: userAdmin,
		},
		{
			Method: "GET",
			Name:   "successful confirm by admin",
			Href:   "/v1/users-confirm/?id=1",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name: "successful signin by username",
			Href: signinURL,
			Src:  &SignInReq{"newsmalex", "123123"},
			Code: 200,
		},
		{
			Name: "successful signin by email",
			Href: signinURL,
			Src:  &SignInReq{"newsmalex69@gmail.com", "123123"},
			Code: 200,
		},
		{
			Name:   "admin get all users",
			Method: "GET",
			Href:   userURL + "?limit=10",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin get all users with query",
			Method: "GET",
			Href:   userURL + "?limit=10&query=admin",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin can get any user",
			Method: "GET",
			Href:   userURL + "1",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin can get any user",
			Method: "GET",
			Href:   userURL + "10000",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:         "user trying to get wrong user",
			Method:       "GET",
			Href:         userURL + "ZZZZ",
			Code:         400,
			ErrorMessage: common.ErrBadRequest.Message(),
			Token:        userAdmin,
		},
		{
			Name:         "get user not found",
			Method:       "GET",
			Href:         userURL + "100",
			Code:         404,
			ErrorMessage: common.ErrNotFound.Message(),
			Token:        userAdmin,
		},
		{
			Name:   "user get his own profile",
			Method: "GET",
			Href:   userURL + "1",
			Code:   200,
			Token:  userNew,
		},
		{
			Name:   "user get settings",
			Method: "GET",
			Href:   "/v1/users-settings/",
			Code:   200,
			Token:  userNew,
		},
		{
			Name:         "unsuccessful user put settings",
			Src:          &PutSettingsReq{NumberCalories: 1000000},
			Href:         "/v1/users-settings/",
			Code:         400,
			Token:        userNew,
			ErrorMessage: ErrNumberCaloriesInvalid.Message(),
		},
		{
			Name:  "successful user put settings",
			Src:   &PutSettingsReq{NumberCalories: 1200},
			Href:  "/v1/users-settings/",
			Code:  200,
			Token: userNew,
		},
		{
			Name: "unsuccessful user put settings without auth",
			Src:  &PutSettingsReq{NumberCalories: 1200},
			Href: "/v1/users-settings/",
			Code: 401,
		},
		{
			Method: "GET",
			Name:   "unsuccessful user put settings without auth",
			Href:   "/v1/users-settings/",
			Code:   401,
		},
		{
			Method:       "GET",
			Href:         userURL + "?limit=10",
			Code:         403,
			Token:        userNew,
			ErrorMessage: common.ErrPermissionDenied.Message(),
		},
		{
			Name:         "user cannot delete himself",
			Method:       "DELETE",
			Href:         userURL + "1",
			Code:         403,
			ErrorMessage: common.ErrPermissionDenied.Message(),
			Token:        userNew,
		},
		{
			Name:   "admin delete user",
			Method: "DELETE",
			Href:   userURL + "1",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "manager allowed to modify admin",
			Method: "POST",
			Href:   userURL,
			Code:   200,
			Token:  userManager,
			Src: gin.H{
				"id":             "10000",
				"username":       "admin",
				"email":          "admin@gmail.com",
				"roleId":         "1",
				"emailConfirmed": true,
				"enabled":        true,
				"created":        "2017-06-06T16:50:45.871269Z",
				"lastVisit":      "0001-01-01T00:00:00Z",
				"numberCalories": 1600,
				"newPassword":    "",
			},
		},
	}
	apptest.CommonApiRunnerAll(r, t, inst, commands)
}
