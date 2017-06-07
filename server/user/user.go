package user

import (
	"crypto/md5"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/smalex-als/health-tracker/server/common"
	"github.com/smalex-als/health-tracker/server/dao"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/search"
	"google.golang.org/appengine/taskqueue"
)

type UserDoc struct {
	Email   search.Atom
	Text    string
	Created time.Time
}

type UserDao struct {
	dao.BaseDao
}

var ErrUsernameIsEmpty = common.NewClientError("username", "Username is empty")
var ErrUsernameIsNotValid = common.NewClientError("username", "Username is not valid format")
var ErrUsernameExists = common.NewClientError("username", "Sorry, that username's taken. Try another?")
var ErrEmailExists = common.NewClientError("email", "Email is already in use")
var ErrEmailIsEmpty = common.NewClientError("email", "Email is empty")
var ErrEmailIsNotValid = common.NewClientError("email", "Email is not valid format")
var ErrPasswordTooShort = common.NewClientError("password", "Password is too short")
var ErrPasswordWrong = common.NewClientError("password", "Incorrect password")
var ErrUserNotFound = common.NewClientError("username", "User not found")
var ErrUserDeleted = common.NewClientError("username", "User deleted")
var ErrNumberCaloriesInvalid = common.NewClientError("numberCalories", "Number calories not valid")

var _ dao.Dao = &UserDao{}

func NewUserDao() *UserDao {
	typeDescUser := dao.NewTypeDesc("types/User.json")
	userDao := &UserDao{dao.BaseDao{typeDescUser}}
	return userDao
}

func (userDao *UserDao) Get(ctx context.Context, id int64, dst interface{}) error {
	if err := userDao.BaseDao.Get(ctx, id, dst); err != nil {
		return err
	}
	u := dst.(*common.User)
	// for new object
	if u.Id == 0 {
		return nil
	}
	if curuser := common.UserFromContext(ctx); curuser != nil {
		if curuser.RoleId == common.ROLE_MANAGER ||
			curuser.RoleId == common.ROLE_ADMIN ||
			curuser.Id == u.Id {
			return nil
		}
	}
	return common.ErrPermissionDenied
}

func (userDao *UserDao) GetAll(ctx context.Context, params map[string]string) (interface{}, error) {
	u := common.UserFromContext(ctx)
	if u == nil || (u.RoleId != common.ROLE_ADMIN && u.RoleId != common.ROLE_MANAGER) {
		return nil, common.ErrPermissionDenied
	}
	query := dao.GetParam(params, "query", "")
	limit := userDao.ValidateLimit(userDao.GetParamInt(params, "limit", 50))
	offset := userDao.ValidateOffset(userDao.GetParamInt(params, "offset", 0))
	if len(query) > 0 {
		searchQuery := fmt.Sprintf(" Text:%s OR Email:%s", query, query)
		ids, err := dao.SearchIndex(ctx, userDao.IndexName(), searchQuery,
			userDao.TypeDesc.ListView.Sort, offset, limit)
		if err != nil {
			return nil, err
		}
		return common.DbGetByIds(ctx, userDao.TypeDesc.Type.Kind, ids)
	} else {
		users := make([]*common.User, 0)
		orderBy := userDao.TypeDesc.ListView.Sort
		q := datastore.NewQuery("User").Order(orderBy).
			Offset(offset).Limit(limit)
		if err := common.DbGetAll(ctx, q, &users); err != nil {
			return nil, err
		}
		return &users, nil
	}
}

func (userDao *UserDao) Put(ctx context.Context, src interface{}) error {
	u := src.(*common.User)
	if u.Id == 0 {
		u.Enabled = true
		u.Created = time.Now()
	}
	if !userDao.UserAccessAllowed(ctx, u) {
		return common.ErrPermissionDenied
	}
	u.Username = strings.TrimSpace(strings.ToLower(u.Username))
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.NewPassword = strings.TrimSpace(u.NewPassword)
	if u.NumberCalories <= 0 || u.NumberCalories > 50000 {
		return ErrNumberCaloriesInvalid
	}
	if err := UserUsernameExists(ctx, u.Username, u.Id); err != nil {
		return err
	}
	if err := UserEmailExists(ctx, u.Email, u.Id); err != nil {
		return err
	}
	isNew := u.Id == 0
	// if new user or user wants to change password
	if isNew || len(u.NewPassword) > 0 {
		if len(u.NewPassword) == 0 || len(u.NewPassword) < 5 {
			return ErrPasswordTooShort
		}
	}

	// save to DB
	if err := userDao.BaseDao.Put(ctx, src); err != nil {
		return err
	}

	newPassword := u.NewPassword
	if len(u.NewPassword) > 0 {
		updPsw := &common.UserPassword{Password: UserEncodePassword(u.Id, u.NewPassword)}
		u.NewPassword = ""
		key := datastore.NewKey(ctx, "UserPassword", "", u.Id, nil)
		if _, err := common.DbPutCached(ctx, key, updPsw); err != nil {
			return err
		}
	}
	if err := userDao.userIndex(ctx, u); err != nil {
		return err
	}
	if isNew && !u.EmailConfirmed {
		if err := userDao.confirmEmail(ctx, u, newPassword); err != nil {
			return err
		}
	}
	return nil
}

func (userDao *UserDao) confirmEmail(ctx context.Context, user *common.User, newPassword string) error {
	params := make(url.Values)
	params.Add("id", strconv.FormatInt(user.Id, 10))
	params.Add("username", user.Username)
	params.Add("password", newPassword)

	_, err := taskqueue.Add(ctx, taskqueue.NewPOSTTask(SEND_CONFIRMATION, params), "")
	return err
}

func (userDao *UserDao) userIndex(ctx context.Context, u *common.User) error {
	index, err := search.Open(userDao.IndexName())
	if err != nil {
		return err
	}
	roleName := "user"
	if u.RoleId == common.ROLE_MANAGER {
		roleName = "manager"
	} else if u.RoleId == common.ROLE_ADMIN {
		roleName = "admin"
	}

	values := strings.Split(u.Email, "@")
	values = append(values, u.Username)
	values = append(values, roleName)
	doc := &UserDoc{
		Created: u.Created,
		Email:   search.Atom(u.Email),
		Text:    strings.Join(values, " "),
	}
	_, err = index.Put(ctx, strconv.FormatInt(u.Id, 10), doc)
	return err
}

func (userDao *UserDao) Delete(ctx context.Context, id int64, src interface{}) error {
	cur := common.UserFromContext(ctx)
	if cur == nil || (cur.RoleId != common.ROLE_ADMIN && cur.RoleId != common.ROLE_MANAGER) {
		return common.ErrPermissionDenied
	}
	u := src.(*common.User)
	if err := userDao.DeleteExpenses(ctx, u.Id, "Expense", "expenses"); err != nil {
		return err
	}
	if err := userDao.DeleteExpenses(ctx, u.Id, "Meal", "meals"); err != nil {
		return err
	}
	key := datastore.NewKey(ctx, "UserPassword", "", u.Id, nil)
	if err := common.DbDeleteCached(ctx, key); err != nil {
		return err
	}
	return userDao.BaseDao.Delete(ctx, id, src)
}

func (userDao *UserDao) DeleteExpenses(ctx context.Context, userId int64, kind, indexName string) error {
	index, err := search.Open(indexName)
	if err != nil {
		return err
	}
	for {
		q := datastore.NewQuery(kind).KeysOnly().Filter("userId=", userId).Limit(100)
		if keys, err := q.GetAll(ctx, nil); err != nil {
			return err
		} else if len(keys) > 0 {
			cacheKeys := make([]string, 0)
			for _, key := range keys {
				cacheKeys = append(cacheKeys, key.String())
				if err := index.Delete(ctx, strconv.FormatInt(key.IntID(), 10)); err != nil {
					return err
				}
			}
			memcache.DeleteMulti(ctx, cacheKeys)
			if err := datastore.DeleteMulti(ctx, keys); err != nil {
				return err
			}
		} else {
			break
		}
	}
	return nil
}

func (userDao *UserDao) UserAccessAllowed(ctx context.Context, subj *common.User) bool {
	u := common.UserFromContext(ctx)
	if u != nil && (u.RoleId == common.ROLE_ADMIN || u.RoleId == common.ROLE_MANAGER) {
		return true
	}
	// if new user or edit yourself
	if (subj != nil) && (u == nil && subj.Id == 0 || u != nil && subj.Id == u.Id) {
		return true
	}
	return false
}

func (userDao *UserDao) UserFindByUsernamePassword(ctx context.Context, username string, password string) (*common.User, error) {
	var email string
	if strings.Contains(username, "@") {
		if err := UserIsEmailValid(username); err != nil {
			return nil, err
		} else {
			email = username
		}
	} else {
		if err := UserIsUsernameValid(username); err != nil {
			return nil, err
		}
	}
	if len(password) == 0 || len(password) < 5 {
		return nil, ErrPasswordTooShort
	}
	q := datastore.NewQuery("User").Limit(1)
	if len(email) > 0 {
		q = q.Filter("email=", email)
	} else {
		q = q.Filter("username=", username)
	}
	var u common.User
	if err := common.DbGetSingle(ctx, q, &u); err != nil {
		return nil, common.ErrUserNotFound
	}
	if !u.Enabled {
		return nil, common.ErrUserDeleted
	}
	var updPsw common.UserPassword
	key := datastore.NewKey(ctx, "UserPassword", "", u.Id, nil)
	if err := common.DbGetCached(ctx, key, &updPsw); err != nil {
		return nil, err
	}
	encPassword := UserEncodePassword(u.Id, password)
	if encPassword != updPsw.Password {
		return nil, ErrPasswordWrong
	}
	return &u, nil
}

func UserIsUsernameValid(username string) error {
	if len(username) == 0 {
		return ErrUsernameIsEmpty
	}
	re := regexp.MustCompile("^[a-z][a-z0-9]{4,32}$")
	if !re.MatchString(username) {
		return ErrUsernameIsNotValid
	}
	return nil
}

func UserIsEmailValid(email string) error {
	if len(email) == 0 {
		return ErrEmailIsEmpty
	}
	e, err := mail.ParseAddress(email)
	if err != nil || email != e.Address {
		return ErrEmailIsNotValid
	}
	return nil
}

func UserUsernameExists(ctx context.Context, username string, id int64) error {
	if err := UserIsUsernameValid(username); err != nil {
		return err
	}
	keys, err := datastore.NewQuery("User").KeysOnly().
		Filter("username=", username).GetAll(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error: search by username %s", err)
	}
	for i := 0; i < len(keys); i++ {
		if keys[i].IntID() != id {
			return ErrUsernameExists
		}
	}
	return nil
}

func UserEmailExists(ctx context.Context, email string, id int64) error {
	if err := UserIsEmailValid(email); err != nil {
		return err
	}
	keys, err := datastore.NewQuery("User").KeysOnly().
		Filter("email=", email).GetAll(ctx, nil)
	if err != nil {
		return fmt.Errorf("Error: search by email %s", err)
	}
	for i := 0; i < len(keys); i++ {
		if keys[i].IntID() != id {
			return ErrEmailExists
		}
	}
	return nil
}

func UserEncodePassword(userId int64, password string) string {
	data := []byte(strconv.FormatInt(userId, 10) + "^&])%([.,$" + password)
	return fmt.Sprintf("%x", md5.Sum(data))
}
