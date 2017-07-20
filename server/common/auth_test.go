package common

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/health-tracker/server/apptest"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestTokenParserKeys(t *testing.T) {
	parser := NewTokenParserKeys()
	tokenString, err := parser.CreateToken("smalex", 777, 1, 1)
	if err != nil {
		t.Fatalf("create token failed %+v", err)
	}
	claims, err := parser.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("parse token failed %+v", err)
	}
	if claims.Name != "smalex" {
		t.Fatalf("expected name failed")
	}
	if claims.UserId != 777 {
		t.Fatalf("expected userId failed")
	}
	if claims.RoleId != 1 {
		t.Fatalf("expected userId failed")
	}
}

func TestTokenParserString(t *testing.T) {
	parser := NewTokenParserString([]byte("abcdefghij"))
	tokenString, err := parser.CreateToken("smalex", 777, 1, 1)
	if err != nil {
		t.Fatalf("create token failed %+v", err)
	}
	claims, err := parser.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("parse token failed %+v", err)
	}
	if claims.Name != "smalex" {
		t.Fatalf("expected name failed")
	}
	if claims.UserId != 777 {
		t.Fatalf("expected userId failed")
	}
	if claims.RoleId != 1 {
		t.Fatalf("expected userId failed")
	}
}

func TestAuth(t *testing.T) {
	r := gin.New()
	register := &HandlerRegister{r}
	register.AddHandler("GET", "/secure/", true, true, func(g *gin.Context) AppError {
		g.JSON(200, gin.H{"Status": "OK"})

		user := UserFromContext(GetAppEngineContext(g))
		if user == nil {
			t.Fail()
		}
		return nil
	})
	register.AddHandler("GET", "/public/", false, false, func(g *gin.Context) AppError {
		g.JSON(200, gin.H{"Status": "OK"})
		return nil
	})

	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()
	req, _ := inst.NewRequest("GET", "/", nil)
	ctx := appengine.NewContext(req)

	users := []*User{
		{Id: 10000, Username: "admin", Enabled: true, RoleId: ROLE_ADMIN, EmailConfirmed: true},
		{Id: 20000, Username: "manager", Enabled: true, RoleId: ROLE_MANAGER, EmailConfirmed: true},
		{Id: 30000, Username: "smalex", Enabled: true, RoleId: ROLE_USER, EmailConfirmed: true},
	}
	keys := make([]*datastore.Key, len(users))
	for i, u := range users {
		keys[i] = datastore.NewKey(ctx, "User", "", u.Id, nil)
	}
	if _, err := datastore.PutMulti(ctx, keys, users); err != nil {
		panic(err)
	}

	userAdmin, _ := CreateToken("admin", 10000, ROLE_ADMIN, 1)
	userExpired, _ := CreateToken("admin", 10000, ROLE_ADMIN, -1)
	// userManager, _ := auth.CreateToken("manager", 2, auth.ROLE_MANAGER)
	// userOrdinary, _ := auth.CreateToken("smalex", 30000, auth.ROLE_USER)
	// userAdmin2, _ := auth.CreateToken("admin2", 4, auth.ROLE_ADMIN
	// userNew, _ := auth.CreateToken("newsmalex", 1, auth.ROLE_USER)

	commands := []*apptest.TestCommand{
		{
			Name:   "admin successful secure",
			Method: "GET",
			Href:   "/secure/",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin successful public",
			Method: "GET",
			Href:   "/secure/",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "using expired token secure",
			Method: "GET",
			Href:   "/secure/",
			Code:   401,
			Token:  userExpired,
		},
		{
			Name:   "using expired token public",
			Method: "GET",
			Href:   "/public/",
			Code:   200,
			Token:  userExpired,
		},
	}
	apptest.CommonApiRunnerAll(r, t, inst, commands)
}
