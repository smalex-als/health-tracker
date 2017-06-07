package meal

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/health-tracker/server/common"
	"github.com/smalex-als/health-tracker/server/dao"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/search"
)

type Meal struct {
	Id          int64     `json:"id,string" datastore:"-"`
	TypeId      int64     `json:"typeId,string" datastore:"typeId" meccano:"MealType"`
	UserId      int64     `json:"userId,string" datastore:"userId" meccano:"User"`
	Date        time.Time `json:"date" datastore:"date"`
	Time        int       `json:"time,string" datastore:"minutes"`
	Description string    `json:"description" datastore:"description"`
	Amount      int       `json:"amount" datastore:"amount"`
}

type MealType struct {
	Id   int64  `json:"id,string" datastore:"-"`
	Name string `json:"name" datastore:"name"`
}

var ErrMealAmountNotValid = common.NewClientError("amount", "Incorrect value for number of calories")
var ErrMealTimeNotValid = common.NewClientError("time", "Incorrect value for time")
var ErrMealDateNotValid = common.NewClientError("date", "Incorrect value for date")

func init() {
	common.RegisterType("Meal", &Meal{})
	common.RegisterType("MealType", &MealType{})
}

type MealDoc struct {
	UserId string
	Date   time.Time
	Time   float64
	Text   string
	Amount float64
}

type MealStatsReq struct {
	Date string `form:"date" json:"date"`
}

type MealStatsResp struct {
	Date  string          `json:"date"`
	Prev  string          `json:"prev"`
	Next  string          `json:"next"`
	Items []MealStatsDate `json:"items"`
	common.BaseResp
}

type MealStatsDate struct {
	Date    string `json:"date"`
	Weekday string `json:"weekday"`
	Total   int    `json:"total"`
	Success bool   `json:"success"`
}

type MealDao struct {
	dao.BaseDao
}

var _ dao.Dao = &MealDao{}

func (mealDao *MealDao) Get(ctx context.Context, id int64, dst interface{}) error {
	if err := mealDao.BaseDao.Get(ctx, id, dst); err != nil {
		return err
	}
	meal := dst.(*Meal)
	// for new object
	if meal.Id == 0 {
		return nil
	}
	if user := common.UserFromContext(ctx); user != nil {
		if user.RoleId == common.ROLE_ADMIN || user.Id == meal.UserId {
			return nil
		}
	}
	return common.ErrPermissionDenied
}

func (mealDao *MealDao) GetAll(ctx context.Context, params map[string]string) (interface{}, error) {
	user := common.UserFromContext(ctx)
	if user == nil {
		return nil, common.ErrAuthorizationRequired
	}
	from, to := mealDao.getDateRange(mealDao.GetParam(params, "date", ""), mealDao.GetParamInt(params, "days", 0))
	fromTime := mealDao.GetParamInt(params, "from", 0)
	toTime := mealDao.GetParamInt(params, "to", 0)
	if fromTime < 0 {
		fromTime = 0
	}
	if toTime < 0 {
		toTime = 0
	}
	query := strings.TrimSpace(params["query"])
	limit := mealDao.ValidateLimit(mealDao.GetParamInt(params, "limit", 50))
	offset := mealDao.ValidateOffset(mealDao.GetParamInt(params, "offset", 0))
	if len(query) == 0 && from == nil && fromTime == 0 && toTime == 0 {
		slice := make([]*Meal, 0)
		q := datastore.NewQuery(mealDao.TypeDesc.Type.Kind).
			Order(mealDao.TypeDesc.ListView.Sort).
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
		if fromTime > 0 {
			searchQuery += fmt.Sprintf(" Time >= %d ", fromTime)
		}
		if toTime > 0 {
			searchQuery += fmt.Sprintf(" Time <= %d ", toTime)
		}
		ids, err := dao.SearchIndex(ctx, mealDao.IndexName(), searchQuery,
			mealDao.TypeDesc.ListView.Sort, offset, limit)
		if err != nil {
			return nil, err
		}
		return common.DbGetByIds(ctx, mealDao.TypeDesc.Type.Kind, ids)
	}
}

func (mealDao *MealDao) getDateRange(date string, days int) (*time.Time, *time.Time) {
	if days < 0 || days > 100 {
		days = 7
	}
	t, err := time.Parse("2006-01-02", date)
	if err == nil {
		from := t.AddDate(0, 0, -days)
		// to := t.AddDate(0, 0, 1)
		to := t
		return &from, &to
	}
	return nil, nil
}

func (mealDao *MealDao) Put(ctx context.Context, src interface{}) error {
	meal := src.(*Meal)
	if meal.UserId == 0 {
		// assign to current user
		if user := common.UserFromContext(ctx); user != nil {
			meal.UserId = user.Id
		}
	}
	if meal.Amount <= 0 || meal.Amount > 10000 {
		return ErrMealAmountNotValid
	}
	if meal.Time < 0 || meal.Time%100 >= 60 || meal.Time/100 >= 24 {
		return ErrMealTimeNotValid
	}
	if meal.Date.IsZero() {
		return ErrMealDateNotValid
	}
	dt := meal.Date.Truncate(24 * time.Hour)
	dt = dt.Add(time.Duration(meal.Time%100) * time.Minute)
	dt = dt.Add(time.Duration(meal.Time/100) * time.Hour)
	meal.Date = dt

	// save to DB
	if err := mealDao.BaseDao.Put(ctx, src); err != nil {
		return err
	}

	ec := common.NewEntityContext(ctx)
	ec.Add(meal)
	ec.Load()
	username, _ := ec.GetStringForField(meal, "UserId.Username")
	typeName, _ := ec.GetStringForField(meal, "TypeId.Name")

	index, err := search.Open(mealDao.IndexName())
	if err != nil {
		return err
	}
	text := strings.Join([]string{meal.Description, username, typeName}, " ")

	doc := &MealDoc{
		UserId: strconv.FormatInt(meal.UserId, 10),
		Date:   meal.Date,
		Text:   text,
		Amount: float64(meal.Amount),
		Time:   float64(meal.Time),
	}
	_, err = index.Put(ctx, strconv.FormatInt(meal.Id, 10), doc)
	return err
}

type MealRemoteService struct {
	dao.RemoteService
}

func (es *MealRemoteService) Register(r *common.HandlerRegister) {
	typeDescMeal := dao.NewTypeDesc("types/Meal.json")
	es.RemoteService = dao.RemoteService{
		TypeDesc: typeDescMeal,
		Entity:   &Meal{},
		Dao:      &MealDao{dao.BaseDao{typeDescMeal}},
	}
	r.AddHandler("GET", "/v1/meals-install/", false, false, es.HandleInstall)
	r.AddHandler("GET", "/v1/meals-stats/", true, true, es.HandleMealStats)
	r.AddHandler("GET", "/v1/meals/:id", true, true, es.HandleGet)
	r.AddHandler("GET", "/v1/meals/", true, true, es.HandleList)
	r.AddHandler("GET", "/v1/meals-form/", true, true, es.HandleForm)
	r.AddHandler("POST", "/v1/meals/", true, true, es.HandlePut)
	r.AddHandler("DELETE", "/v1/meals/:id", true, true, es.HandleDelete)
}

func (es *MealRemoteService) HandleInstall(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	if cnt, _ := datastore.NewQuery("MealType").Count(ctx); cnt == 0 {
		deleteIndex(ctx, "meals")
		types := []MealType{
			{1, "Breakfast"},
			{2, "Lunch"},
			{3, "Dinner"},
			{4, "Snack"},
		}
		keys := make([]*datastore.Key, len(types))
		for i, exType := range types {
			keys[i] = datastore.NewKey(ctx, "MealType", "", exType.Id, nil)
		}
		if _, err := datastore.PutMulti(ctx, keys, types); err != nil {
			panic(err)
		}
	}

	c.String(200, "OK")
	return nil
}

func (es *MealRemoteService) HandleMealStats(c *gin.Context) common.AppError {
	var in MealStatsReq
	if c.Bind(&in) != nil {
		return common.ErrBadRequest
	}
	tm, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		tm = time.Now()
	}
	ctx := common.GetAppEngineContext(c)
	cnt := (int(tm.Weekday()) + 6) % 7
	monday := tm.AddDate(0, 0, -cnt)
	monday = monday.Truncate(24 * time.Hour)
	prevMonday := monday.AddDate(0, 0, -7)
	nextMonday := monday.AddDate(0, 0, 7)
	user := common.UserFromContext(ctx)
	q := datastore.NewQuery("Meal").
		Filter("userId=", user.Id).
		Filter("date >=", monday).
		Filter("date <", nextMonday)
	t := q.Run(ctx)
	var total int
	totalPerDay := make(map[int]int, 0)
	for {
		var p Meal
		_, err := t.Next(&p)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return common.AppErrorf(err, "get data failed")
		}
		total += p.Amount
		day := int(p.Date.Weekday())
		total += p.Amount
		totalPerDay[day] += p.Amount
	}
	cur := monday
	items := make([]MealStatsDate, 0)
	for cur.Before(nextMonday) {
		day := int(cur.Weekday())
		item := MealStatsDate{
			Date:    cur.Format("2006-01-02"),
			Weekday: cur.Weekday().String()[0:3],
			Total:   totalPerDay[day],
			Success: totalPerDay[day] <= user.NumberCalories,
		}
		items = append(items, item)
		cur = cur.Add(time.Duration(24) * time.Hour)
	}
	resp := &MealStatsResp{
		Date:  monday.Format("2006-01-02"),
		Next:  nextMonday.Format("2006-01-02"),
		Prev:  prevMonday.Format("2006-01-02"),
		Items: items,
	}
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
