package expense

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/expense-tracker/server/apptest"
	"github.com/smalex-als/health-tracker/server/common"
	"github.com/smalex-als/health-tracker/server/dao"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestInit(t *testing.T) {
	r := gin.New()
	register := &common.HandlerRegister{r}
	(&ExpenseRemoteService{}).Register(register)
	http.Handle("/", r)

	inst, err := aetest.NewInstance(&aetest.Options{StronglyConsistentDatastore: true})
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()
	req, _ := inst.NewRequest("GET", "/", nil)
	ctx := appengine.NewContext(req)

	users := []*common.User{
		{Id: 1, Username: "admin", Enabled: true, RoleId: common.ROLE_ADMIN, EmailConfirmed: true},
		{Id: 2, Username: "manager", Enabled: true, RoleId: common.ROLE_MANAGER, EmailConfirmed: true},
		{Id: 3, Username: "smalex", Enabled: true, RoleId: common.ROLE_USER, EmailConfirmed: true},
		{Id: 4, Username: "admin2", Enabled: true, RoleId: common.ROLE_ADMIN, EmailConfirmed: true},
	}
	keys := make([]*datastore.Key, len(users))
	for i, u := range users {
		keys[i] = datastore.NewKey(ctx, "User", "", u.Id, nil)
	}
	if _, err := datastore.PutMulti(ctx, keys, users); err != nil {
		panic(err)
	}
	types := []*ExpenseType{
		{1, "Food", "groceries and eating out"},
		{2, "Home", "mortgage/rent, furniture, appliances, maintenance & improvement"},
	}
	keys = make([]*datastore.Key, len(types))
	for i, exType := range types {
		keys[i] = datastore.NewKey(ctx, "ExpenseType", "", exType.Id, nil)
	}
	if _, err := datastore.PutMulti(ctx, keys, types); err != nil {
		panic(err)
	}

	userAdmin, _ := common.CreateToken("admin", 1, common.ROLE_ADMIN, 1)
	userManager, _ := common.CreateToken("manager", 2, common.ROLE_MANAGER, 1)
	userOrdinary, _ := common.CreateToken("smalex", 3, common.ROLE_USER, 1)
	userAdmin2, _ := common.CreateToken("admin2", 4, common.ROLE_ADMIN, 1)

	const expenseURL = "/v1/expenses/"
	commands := []*apptest.TestCommand{
		{
			Name: "admin create expense",
			Href: expenseURL,
			Src: &Expense{
				Amount:      5.99,
				Comment:     "Coffee",
				Description: "very tasty",
				Date:        time.Now(),
			},
			Dst:   &dao.CommonPutResp{},
			Code:  200,
			Token: userAdmin,
			Validate: func(t *testing.T, cmd *apptest.TestCommand) {
				if cmd.Dst != nil {
					resp := (cmd.Dst).(*dao.CommonPutResp)
					values := (resp.Data).(map[string]interface{})
					if values["id"] != "1" {
						t.Fatal("expected id = 1")
					}
					if values["comment"] != "Coffee" {
						t.Fatal("expected comment = 'Coffee'")
					}
					fmt.Printf("cmd.Dst = %+v\n", values)
				}
			},
		},
		{
			Name: "admin can update his own expense",
			Href: expenseURL,
			Src: &Expense{
				Id:          1,
				Amount:      5.99,
				Comment:     "Coffee",
				Description: "very tasty",
				Date:        time.Now(),
			},
			Code:  200,
			Token: userAdmin,
		},
		{
			Name: "manager create expense",
			Href: expenseURL,
			Src: &Expense{
				Amount:      3.99,
				Comment:     "Tea",
				Description: "and biscit",
				Date:        time.Now(),
			},
			Code:  200,
			Token: userManager,
		},
		{
			Name: "manager can update his own expense",
			Href: expenseURL,
			Src: &Expense{
				Id:          2,
				Amount:      4.99,
				Comment:     "Tea",
				Description: "and biscit",
				Date:        time.Now(),
			},
			Code:  200,
			Token: userManager,
		},
		{
			Name: "admin can update any expense",
			Href: expenseURL,
			Src: &Expense{
				Id:          2,
				Amount:      4.99,
				Comment:     "Tea",
				Description: "and biscit",
				Date:        time.Now(),
			},
			Code:  200,
			Token: userAdmin,
		},
		{
			Name:         "expense not found",
			Method:       "GET",
			Href:         expenseURL + "77777",
			Code:         404,
			ErrorMessage: common.ErrNotFound.Message(),
			Token:        userAdmin,
		},
		{
			Name:   "admin get all expenses",
			Method: "GET",
			Href:   expenseURL,
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin get all expenses with query",
			Method: "GET",
			Href:   expenseURL + "?query=tea",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "user get expense form",
			Method: "GET",
			Href:   "/v1/expenses-form/",
			Code:   200,
			Token:  userManager,
		},
		{
			Name:   "user get all expenses",
			Method: "GET",
			Href:   expenseURL,
			Code:   200,
			Token:  userManager,
		},
		{
			Name:   "user get all",
			Method: "GET",
			Href:   expenseURL,
			Code:   401,
		},
		{
			Name:   "user get all expenses with query",
			Method: "GET",
			Href:   expenseURL + "?query=tea&date=" + time.Now().Format("2006-01-02"),
			Code:   200,
			Token:  userManager,
		},
		{
			Name:   "user get stats",
			Method: "GET",
			Href:   "/v1/expenses-stats/",
			Code:   200,
			Token:  userManager,
		},
		{
			Name:   "admin get expense his own",
			Method: "GET",
			Href:   expenseURL + "1",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin get expense any expense",
			Method: "GET",
			Href:   expenseURL + "2",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "admin2 get expense any expense",
			Method: "GET",
			Href:   expenseURL + "1",
			Code:   200,
			Token:  userAdmin2,
		},
		{
			Name:         "manager cannot see expense if he is not owner",
			Method:       "GET",
			Href:         expenseURL + "1",
			Code:         403,
			ErrorMessage: common.ErrPermissionDenied.Message(),
			Token:        userManager,
		},
		{
			Name: "manager cannot update expense if he is not owner",
			Href: expenseURL,
			Src: &Expense{
				Id:          1,
				Amount:      5.99,
				Comment:     "Coffee",
				Description: "very tasty",
				Date:        time.Now(),
			},
			Code:         403,
			ErrorMessage: common.ErrPermissionDenied.Message(),
			Token:        userManager,
		},
		{
			Name:         "user cannot get expense if he is not owner",
			Method:       "GET",
			Href:         expenseURL + "1",
			Code:         403,
			ErrorMessage: common.ErrPermissionDenied.Message(),
			Token:        userOrdinary,
		},
		{
			Name:         "manager cannot delete expense if he is not owner",
			Method:       "DELETE",
			Href:         expenseURL + "1",
			Code:         403,
			ErrorMessage: common.ErrPermissionDenied.Message(),
			Token:        userManager,
		},
		{
			Name:         "user cannot delete expense if he is not owner",
			Method:       "DELETE",
			Href:         expenseURL + "1",
			Code:         403,
			ErrorMessage: common.ErrPermissionDenied.Message(),
			Token:        userOrdinary,
		},
		{
			Name:   "admin delete expense his own expense",
			Method: "DELETE",
			Href:   expenseURL + "1",
			Code:   200,
			Token:  userAdmin,
		},
		{
			Name:   "manager delete expense his own expense",
			Method: "DELETE",
			Href:   expenseURL + "2",
			Code:   200,
			Token:  userManager,
		},
		{
			Name:         "admin cannot delete expense if it is not exists",
			Method:       "DELETE",
			Href:         expenseURL + "7777",
			Code:         404,
			ErrorMessage: common.ErrNotFound.Message(),
			Token:        userAdmin,
		},
		{
			Name:   "if user is not logged he cannot view list of expenses",
			Method: "GET",
			Href:   expenseURL,
			Code:   401,
		},
		{
			Name:   "if user is not logged he cannot view expense",
			Method: "GET",
			Href:   expenseURL + "1",
			Code:   401,
		},
	}
	apptest.CommonApiRunnerAll(r, t, inst, commands)
}
