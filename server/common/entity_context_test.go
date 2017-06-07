package common

import (
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

type Expense struct {
	Id     int64 `json:"id,string" datastore:"-"`
	UserId int64 `json:"userId,string" datastore:"userId" meccano:"User"`
}

func TestEntityContextInit(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	user := addUser(ctx, "User", 1, &User{Id: 1, Username: "smalex69"})
	addUser(ctx, "User", 2, &User{Id: 2, Username: "manager"})
	addUser(ctx, "User", 3, &User{Id: 3, Username: "admin"})
	ec := NewEntityContext(ctx)
	RegisterType("User", &User{})
	RegisterType("Expense", &Expense{})
	ec.Add(user)
	ec.Add(&Expense{Id: 2, UserId: 1})
	ec.Add(&Expense{Id: 3, UserId: 1})
	expense := &Expense{Id: 4, UserId: 2}
	expense2 := &Expense{Id: 5}
	expense3 := &Expense{Id: 5, UserId: 151}
	ec.Add(expense)
	ec.Add(expense2)
	ec.Add(expense3)
	ec.Load()

	username, err := ec.GetStringForField(expense, "UserId.Username")
	if err != nil || username != "manager" {
		t.Errorf("username should be manager, but was %s", username)
	}

	if _, err := ec.GetStringForField(expense, "UserId.Username2"); err == nil {
		t.Errorf("Must be error")
	} else {
		t.Log(err)
	}
	if _, err := ec.GetStringForField(expense, "User.Username2"); err == nil {
		t.Errorf("Must be error")
	} else {
		t.Log(err)
	}
	if _, err := ec.GetStringForField(expense, "UserId.User.Name"); err == nil {
		t.Errorf("Must be error")
	} else {
		t.Log(err)
	}
	if _, err := ec.GetStringForField(expense, ".User.Name"); err == nil {
		t.Errorf("Must be error")
	} else {
		t.Log(err)
	}
	if _, err := ec.GetStringForField(expense2, "UserId.Username"); err != nil {
		t.Errorf("error was not expected %+v", err)
	}
	if _, err := ec.GetStringForField(expense3, "UserId.Username"); err == nil {
		t.Errorf("Must be error")
	} else {
		t.Log(err)
	}
	// if val, ok := toUser.(*auth.User); ok {
	// 	fmt.Printf("toUser = %+v %+T\n", val, val)
	// 	username, _ = ec.getStringValue(val, "Username")
	// }
}

func addUser(c context.Context, kind string, id int64, ref interface{}) interface{} {
	key := datastore.NewKey(c, kind, "", id, nil)
	if _, err := datastore.Put(c, key, ref); err != nil {
		panic(err)
	}
	return ref
}
