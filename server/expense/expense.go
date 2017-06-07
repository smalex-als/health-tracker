package expense

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/health-tracker/server/common"
	"github.com/smalex-als/health-tracker/server/dao"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

type Expense struct {
	Id          int64     `json:"id,string" datastore:"-"`
	TypeId      int64     `json:"typeId,string" datastore:"typeId" meccano:"ExpenseType"`
	UserId      int64     `json:"userId,string" datastore:"userId" meccano:"User"`
	Date        time.Time `json:"date" datastore:"date"`
	Description string    `json:"description" datastore:"description"`
	Amount      float64   `json:"amount" datastore:"amount"`
	Comment     string    `json:"comment" datastore:"comment,noindex"`
}

type ExpenseType struct {
	Id          int64  `json:"id,string" datastore:"-"`
	Name        string `json:"name" datastore:"name"`
	Description string `json:"description" datastore:"description"`
}

var ErrExpenseAmountNotValid = common.NewClientError("amount", "Amount is not valid format")

func init() {
	common.RegisterType("Expense", &Expense{})
	common.RegisterType("ExpenseType", &ExpenseType{})
}

type ExpenseDoc struct {
	UserId string
	Date   time.Time
	Text   string
	Amount float64
}

type ExpenseStatsReq struct {
	Date string `form:"date" json:"date"`
}

type ExpenseStatsResp struct {
	Date   string                `json:"date"`
	Prev   string                `json:"prev"`
	Next   string                `json:"next"`
	Total  float64               `json:"total"`
	Avg    float64               `json:"avg"`
	Errors []*common.ClientError `json:"errors,omitempty"`
}

type ExpenseDao struct {
	dao.BaseDao
}

var _ dao.Dao = &ExpenseDao{}

func (expenseDao *ExpenseDao) Get(ctx context.Context, id int64, dst interface{}) error {
	if err := expenseDao.BaseDao.Get(ctx, id, dst); err != nil {
		return err
	}
	expense := dst.(*Expense)
	// for new object
	if expense.Id == 0 {
		return nil
	}
	if user := common.UserFromContext(ctx); user != nil {
		if user.RoleId == common.ROLE_ADMIN || user.Id == expense.UserId {
			return nil
		}
	}
	return common.ErrPermissionDenied
}

func (expenseDao *ExpenseDao) GetAll(ctx context.Context, params map[string]string) (interface{}, error) {
	user := common.UserFromContext(ctx)
	if user == nil {
		return nil, common.ErrAuthorizationRequired
	}
	from, to := expenseDao.getDateRange(expenseDao.GetParam(params, "date", ""),
		expenseDao.GetParamInt(params, "days", 0))
	query := params["query"]
	limit := expenseDao.ValidateLimit(expenseDao.GetParamInt(params, "limit", 50))
	offset := expenseDao.ValidateOffset(expenseDao.GetParamInt(params, "offset", 0))
	if len(query) == 0 && from == nil {
		slice := make([]*Expense, 0)
		q := datastore.NewQuery(expenseDao.TypeDesc.Type.Kind).
			Order(expenseDao.TypeDesc.ListView.Sort).
			Offset(offset).Limit(limit)
		if user.RoleId < common.ROLE_ADMIN {
			q = q.Filter("userId=", user.Id)
		}
		if err := common.DbGetAll(ctx, q, &slice); err != nil {
			return nil, err
		}
		return &slice, nil
	} else {
		var searchQuery string
		if user.RoleId < common.ROLE_ADMIN {
			searchQuery += fmt.Sprintf(" UserId:%d ", user.Id)
		}
		if len(query) > 0 {
			searchQuery += fmt.Sprintf(" Text:%s ", query)
		}
		if from != nil {
			searchQuery += fmt.Sprintf(" Date > %s ", from.Format("2006-01-02"))
		}
		if to != nil {
			searchQuery += fmt.Sprintf(" Date <= %s ", to.Format("2006-01-02"))
		}
		ids, err := dao.SearchIndex(ctx, expenseDao.IndexName(), searchQuery,
			expenseDao.TypeDesc.ListView.Sort, offset, limit)
		if err != nil {
			return nil, err
		}
		return common.DbGetByIds(ctx, expenseDao.TypeDesc.Type.Kind, ids)
	}
}

func (expenseDao *ExpenseDao) getDateRange(date string, days int) (*time.Time, *time.Time) {
	if days < 0 || days > 100 {
		days = 7
	}
	t, err := time.Parse("2006-01-02", date)
	if err == nil {
		from := t.AddDate(0, 0, -days)
		to := t.AddDate(0, 0, 1)
		return &from, &to
	}
	return nil, nil
}

func (expenseDao *ExpenseDao) Put(ctx context.Context, src interface{}) error {
	expense := src.(*Expense)
	if expense.UserId == 0 {
		// assign to current user
		if user := common.UserFromContext(ctx); user != nil {
			expense.UserId = user.Id
		}
	}
	// normalize amount
	amount, _ := strconv.ParseFloat(fmt.Sprintf("%.02f", expense.Amount), 64)
	expense.Amount = amount

	if !(expense.Amount > 0) {
		return ErrExpenseAmountNotValid
	}

	// save to DB
	if err := expenseDao.BaseDao.Put(ctx, src); err != nil {
		return err
	}

	index, err := search.Open(expenseDao.IndexName())
	if err != nil {
		return err
	}
	doc := &ExpenseDoc{
		UserId: strconv.FormatInt(expense.UserId, 10),
		Date:   expense.Date,
		Text:   expense.Description + " " + expense.Comment,
		Amount: expense.Amount,
	}
	_, err = index.Put(ctx, strconv.FormatInt(expense.Id, 10), doc)
	return err
}

type ExpenseRemoteService struct {
	dao.RemoteService
}

func (es *ExpenseRemoteService) Register(r *common.HandlerRegister) {
	typeDescExpense := dao.NewTypeDesc("types/Expense.json")
	es.RemoteService = dao.RemoteService{
		TypeDesc: typeDescExpense,
		Entity:   &Expense{},
		Dao:      &ExpenseDao{dao.BaseDao{typeDescExpense}},
	}
	r.AddHandler("GET", "/v1/expenses-install/", false, false, es.HandleInstall)
	r.AddHandler("GET", "/v1/expenses-stats/", true, true, es.HandleExpenseStats)
	r.AddHandler("GET", "/v1/expenses/:id", true, true, es.HandleGet)
	r.AddHandler("GET", "/v1/expenses/", true, true, es.HandleList)
	r.AddHandler("GET", "/v1/expenses-form/", true, true, es.HandleForm)
	r.AddHandler("POST", "/v1/expenses/", true, true, es.HandlePut)
	r.AddHandler("DELETE", "/v1/expenses/:id", true, true, es.HandleDelete)
}

func (es *ExpenseRemoteService) HandleInstall(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)

	if cnt, _ := datastore.NewQuery("ExpenseType").Count(ctx); cnt == 0 {
		deleteIndex(ctx, "expenses")
		types := []ExpenseType{
			{1, "Food", "groceries and eating out"},
			{2, "Home", "mortgage/rent, furniture, appliances, maintenance & improvement"},
			{3, "Car", "insurance, gas, repairs, and any payments"},
			{4, "Giving", "tithing, or general donations"},
			{5, "Medical", "insurance, copays, and prescriptions"},
			{6, "Utilities", "gas, electric, water, cable, internet, and phone"},
			{7, "Personal Care", "toiletries, hair, makeup, clothes, etc"},
			{8, "Gifts", "Christmas, birthday, and cards"},
			{9, "Other", "pets, kids, entertainment, hobbies, and miscellaneous household items"},
		}
		keys := make([]*datastore.Key, len(types))
		for i, exType := range types {
			keys[i] = datastore.NewKey(ctx, "ExpenseType", "", exType.Id, nil)
		}
		if _, err := datastore.PutMulti(ctx, keys, types); err != nil {
			panic(err)
		}
	}
	c.String(200, "OK")
	return nil
}

func (es *ExpenseRemoteService) HandleExpenseStats(c *gin.Context) common.AppError {
	resp := &ExpenseStatsResp{}
	var in ExpenseStatsReq
	if c.Bind(&in) != nil {
		return common.AppErrorf(common.ErrBadRequest, "bad request")
	}
	tm, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		tm = time.Now()
	}
	monday := tm.AddDate(0, 0, -int(tm.Weekday()))
	monday = monday.Truncate(24 * time.Hour)
	prevMonday := monday.AddDate(0, 0, -7)
	nextMonday := monday.AddDate(0, 0, 7)
	ctx := common.GetAppEngineContext(c)
	user := common.UserFromContext(ctx)
	q := datastore.NewQuery("Expense").
		Filter("userId=", user.Id).
		Filter("date >=", monday).
		Filter("date <", nextMonday)
	t := q.Run(ctx)
	var total float64
	totalPerDay := make(map[int]float64, 0)
	for {
		var p Expense
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return common.AppErrorf(err, "get data failed")
		}
		day := int(p.Date.Weekday())
		total += p.Amount
		totalPerDay[day] += p.Amount
	}
	resp.Total = total
	if len(totalPerDay) > 0 {
		resp.Avg = total / 7
	}
	resp.Date = monday.Format("2006-01-02")
	resp.Next = nextMonday.Format("2006-01-02")
	resp.Prev = prevMonday.Format("2006-01-02")
	c.JSON(http.StatusOK, resp)
	return nil
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
