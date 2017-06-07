package dao

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/smalex-als/health-tracker/server/common"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type CommonGetResp struct {
	Data       interface{}            `json:"data,omitempty"`
	References map[string]interface{} `json:"references,omitempty"`
	common.BaseResp
}

type CommonFormResp struct {
	EditView   *EditView                  `json:"editView,omitempty"`
	References map[string][]ListBoxOption `json:"references,omitempty"`
	common.BaseResp
}

type CommonDeleteResp struct {
	common.BaseResp
}

type CommonPutResp struct {
	Data interface{} `json:"data,omitempty"`
	common.BaseResp
}

type ListRow struct {
	Id     int64    `json:"id,string"`
	Values []string `json:"values"`
}

type CommonListResp struct {
	Columns []*ListViewColumn `json:"columns,omitempty"`
	Rows    []ListRow         `json:"rows,omitempty"`
	Items   interface{}       `json:"items,omitempty"`
	common.BaseResp
}

type ListBoxOption struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

type RemoteService struct {
	TypeDesc *TypeDesc
	Entity   interface{}
	Dao      Dao
	common.ServiceRegister
}

func (service *RemoteService) NewEntity() interface{} {
	return reflect.New(reflect.TypeOf(service.Entity).Elem()).Interface()
}

func (service *RemoteService) HandleList(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	resp := &CommonListResp{}

	reqValues := c.Request.URL.Query()
	values := make(map[string]string)
	for k, _ := range reqValues {
		values[k] = reqValues.Get(k)
	}
	items, err := service.Dao.GetAll(ctx, values)
	if err != nil {
		return common.AppErrorf(err, "get all failed")
	}
	if GetParam(values, "mode", "table") == "items" {
		resp.Items = items
	} else {
		rows, err := service.createListRowsResp(ctx, items)
		if err != nil {
			return common.AppErrorf(err, "create list rows failed")
		}
		resp.Columns = service.TypeDesc.ListView.Columns
		resp.Rows = rows
	}
	c.JSON(http.StatusOK, resp)
	return nil
}

func (service *RemoteService) HandleDelete(c *gin.Context) common.AppError {
	var dst interface{}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return common.AppErrorf(common.ErrBadRequest, "parsing id failed %+v", err)
	}
	ctx := common.GetAppEngineContext(c)
	dst = service.NewEntity()
	if err := service.Dao.Get(ctx, id, dst); err != nil {
		return common.AppErrorf(err, "getting from db failed")
	}
	if err := service.Dao.Delete(ctx, id, dst); err != nil {
		return common.AppErrorf(err, "deleting failed")
	}
	c.JSON(http.StatusOK, gin.H{})
	return nil
}

func (service *RemoteService) HandleForm(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	resp := &CommonFormResp{}
	resp.EditView = service.TypeDesc.EditView
	resp.References = make(map[string][]ListBoxOption)
	for _, w := range service.TypeDesc.EditView.Widgets {
		if w.DataSource != nil {
			options, err := service.loadDataSource(ctx, w.DataSource)
			if err != nil {
				return common.AppErrorf(err, "load data source failed %+v", w.DataSource)
			}
			resp.References[w.Id] = options
		}
	}
	c.JSON(http.StatusOK, resp)
	return nil
}

func (service *RemoteService) loadDataSource(ctx context.Context, ds *DataSource) ([]ListBoxOption, error) {
	slice := common.MakePtrSlice(ds.Kind, 0)
	err := common.DbGetAll(ctx, datastore.NewQuery(ds.Kind), slice)
	if err != nil {
		panic(err)
	}
	s := reflect.ValueOf(slice)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	res := make([]ListBoxOption, 0)
	if ds.EmptyRow {
		res = append(res, ListBoxOption{Value: "", Name: "..."})
	}
	for i := 0; i < s.Len(); i++ {
		v := s.Index(i)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		value, err := service.getValue(v, "Id")
		if err != nil {
			return nil, err
		}
		name, err := service.getValue(v, ds.ColumnName)
		if err != nil {
			return nil, err
		}
		res = append(res, ListBoxOption{Value: value, Name: name})
	}
	return res, nil
}

func (service *RemoteService) getValue(s reflect.Value, name string) (string, error) {
	var tm time.Time
	var val string
	name = common.CapitalizeFirstLetter(name)
	f := s.FieldByName(name)
	if !f.IsValid() {
		return "", fmt.Errorf("field %q not found in %s", name, s.Type().Name())
	}
	switch f.Kind() {
	case reflect.Float64:
		val = fmt.Sprintf("%.06f", f.Float())
	case reflect.Int64:
		val = strconv.FormatInt(f.Int(), 10)
	case reflect.Int:
		val = strconv.FormatInt(f.Int(), 10)
	case reflect.String:
		val = f.String()
	case reflect.Bool:
		val = strconv.FormatBool(f.Bool())
	case reflect.Struct:
		switch f.Type() {
		case reflect.TypeOf(tm):
			tm1 := f.Interface().(time.Time)
			val = tm1.Format(time.RFC3339)
		}
	default:
		return "", fmt.Errorf("unsupported field %+v", f.Kind())
	}
	return val, nil
}

func (service *RemoteService) HandleGet(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	resp := &CommonGetResp{}
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return common.AppErrorf(common.ErrBadRequest, "parsing id failed %+v", err)
	}
	dst := service.NewEntity()
	if err := service.Dao.Get(ctx, id, dst); err != nil {
		return common.AppErrorf(err, "getting from db failed")
	}
	resp.Data = dst
	ec := common.NewEntityContext(ctx)
	ec.Add(dst)
	ec.Load()
	references := make(map[string]interface{})
	for _, f := range service.TypeDesc.Type.Fields {
		if f.FieldType != FieldTypeReference {
			continue
		}
		ref, err := ec.GetEntityForField(dst, f.Id)
		if err != nil {
			return common.AppErrorf(err, "getting reference for field %+v failed", f.Id)
		}
		name := f.Id
		if strings.HasSuffix(name, "Id") {
			name = name[:len(name)-2]
		}
		references[strings.ToLower(name)] = ref
	}
	if len(references) > 0 {
		resp.References = references
	}
	c.JSON(http.StatusOK, resp)
	return nil
}

func (service *RemoteService) createListRowsResp(ctx context.Context, dst interface{}) ([]ListRow, error) {
	typeDesc := service.TypeDesc
	slice := reflect.ValueOf(dst)
	if slice.Kind() == reflect.Ptr {
		slice = slice.Elem()
	}
	if slice.Kind() != reflect.Slice {
		return nil, fmt.Errorf("createListRowsResp: wrong dst type")
	}
	ec := common.NewEntityContext(ctx)
	for i := 0; i < slice.Len(); i++ {
		ec.Add(slice.Index(i).Interface())
	}
	ec.Load()
	rows := make([]ListRow, 0)
	columns := typeDesc.ListView.Columns
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		s := reflect.ValueOf(item)
		if s.Kind() == reflect.Ptr {
			s = s.Elem()
		}
		fieldId := s.FieldByName("Id")
		if !fieldId.IsValid() || fieldId.Kind() != reflect.Int64 {
			return nil, fmt.Errorf("createListRowsResp: could not find Id field")
		}
		row := ListRow{
			Id:     fieldId.Int(),
			Values: make([]string, len(columns)),
		}
		for i, column := range columns {
			var val string
			var err error
			if strings.Contains(column.Path, ".") {
				val, err = ec.GetStringForField(item, column.Path)
				if err != nil {
					return nil, err
				}
			} else {
				if column.Id == "id" {
					val = strconv.FormatInt(fieldId.Int(), 10)
				} else {
					val, err = service.getValue(s, column.Id)
					if err != nil {
						return nil, err
					}
				}
			}
			if val != "" && len(column.Function) > 0 {
				switch column.Function {
				case "human_float":
					m, _ := strconv.ParseFloat(val, 64)
					val = humanize.Ftoa(m)
				case "human_datetime":
					tm, _ := time.Parse(time.RFC3339, val)
					val = humanize.Time(tm)
				case "human_date":
					tm, _ := time.Parse(time.RFC3339, val)
					val = tm.Format("2006-01-02")
				case "human_time":
					if len(val) < 4 {
						val = "0000"[len(val):] + val
					}
					val = val[0:2] + ":" + val[2:]
				}
			}
			row.Values[i] = val
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (service *RemoteService) HandlePut(c *gin.Context) common.AppError {
	ctx := common.GetAppEngineContext(c)
	in := map[string]interface{}{}
	if err := c.BindJSON(&in); err != nil {
		return common.AppErrorf(common.ErrBadRequest, "bad request")
	}
	var id int64
	if idVal, ok := in["id"]; ok {
		var err error
		if idStr, ok := idVal.(string); !ok {
			return common.AppErrorf(common.ErrBadRequest, "id not specified")
		} else if id, err = strconv.ParseInt(idStr, 10, 64); err != nil {
			return common.AppErrorf(common.ErrBadRequest, "parsing id failed %+v", err)
		}
	}
	dst := service.NewEntity()
	if id != 0 {
		if err := service.Dao.Get(ctx, id, dst); err != nil {
			return common.AppErrorf(err, "getting from db failed")
		}
	}
	if err := service.formToRecord(ctx, dst, in); err != nil {
		return common.AppErrorf(err, "parse form failed")
	}
	if err := service.Dao.Put(ctx, dst); err != nil {
		return common.AppErrorf(err, "put failed")
	}
	c.JSON(http.StatusOK, &CommonPutResp{Data: dst})
	return nil
}

func (service *RemoteService) formToRecord(ctx context.Context, record interface{}, values map[string]interface{}) error {
	typeDesc := service.TypeDesc
	s := reflect.ValueOf(record)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("FromMap only accepts structs; got %T", s))
	}
	var tm time.Time
	for _, widget := range typeDesc.EditView.Widgets {
		val := values[widget.Id]
		newValue := reflect.ValueOf(val)
		// exported field
		name := common.CapitalizeFirstLetter(widget.Id)
		f := s.FieldByName(name)
		if !f.IsValid() {
			return errors.New(fmt.Sprintf("field %q not found %s %+v", name, s.Kind()))
		}
		if !f.CanSet() {
			return errors.New(fmt.Sprintf("field %q cannot be set", name))
		}
		// log.Infof(ctx, "newvalue %+v %+T %+v req %+v", widget.Id, val, newValue.Type(), f.Kind())
		switch f.Kind() {
		case reflect.Bool:
			f.SetBool(newValue.Bool())
		case reflect.Float64:
			switch newValue.Kind() {
			case reflect.Float64:
				f.SetFloat(newValue.Float())
			case reflect.Int64:
				f.SetFloat(float64(newValue.Int()))
			default:
				return fmt.Errorf("field %q unsupported type %+T", name, val)
			}
		case reflect.Int:
			switch newValue.Kind() {
			case reflect.Float64:
				f.SetInt(int64(newValue.Float()))
			case reflect.String:
				if len(newValue.String()) > 0 {
					intval, err := strconv.ParseInt(newValue.String(), 10, 64)
					if err != nil {
						return fmt.Errorf("field %q cannot parse int %+v", name, err)
					}
					f.SetInt(intval)
				} else {
					f.SetInt(0)
				}
			default:
				return fmt.Errorf("field %q unsupported type %+T", name, val)
			}
		case reflect.Int64:
			switch newValue.Kind() {
			case reflect.String:
				if len(newValue.String()) > 0 {
					intval, err := strconv.ParseInt(newValue.String(), 10, 64)
					if err != nil {
						return fmt.Errorf("field %q cannot parse int %+v", name, err)
					}
					f.SetInt(intval)
				} else {
					f.SetInt(0)
				}
			default:
				return fmt.Errorf("field %q unsupported type %+T", name, val)
			}
		case reflect.String:
			switch newValue.Kind() {
			case reflect.String:
				f.SetString(newValue.String())
			}
		case reflect.Struct:
			switch f.Type() {
			case reflect.TypeOf(tm):
				if newValue.String() != "" {
					tm, err := time.Parse(time.RFC3339, newValue.String())
					if err != nil {
						return err
					}
					f.Set(reflect.ValueOf(tm))
				} else {
					f.Set(reflect.ValueOf(tm))
				}
			}
		default:

			log.Infof(ctx, "failed newvalue %+v %+T %+v req %+v", widget.Id, val, newValue.Type(), f.Kind())
		}
	}
	return nil
}

type SearchDoc struct {
	c *gin.Context
}

func (d *SearchDoc) Load(fields []search.Field, meta *search.DocumentMetadata) error {
	ctx := common.GetAppEngineContext(d.c)
	log.Infof(ctx, "fields = %+v\n", fields)
	return nil
}

func (d *SearchDoc) Save() ([]search.Field, *search.DocumentMetadata, error) {
	return nil, nil, nil
}

func SearchIndex(c context.Context, indexName, query, orderBy string, offset, limit int) ([]int64, error) {
	log.Infof(c, query)
	var expr string
	var reverse bool
	if strings.HasPrefix(orderBy, "-") {
		expr = orderBy[1:]
		reverse = false
	} else {
		expr = orderBy
		reverse = true
	}
	searchSort := &search.SortOptions{
		Expressions: []search.SortExpression{
			{Expr: common.CapitalizeFirstLetter(expr), Reverse: reverse},
		},
	}
	opts := &search.SearchOptions{
		Offset: offset,
		Limit:  limit,
		Sort:   searchSort,
	}
	index, err := search.Open(indexName)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0)
	for t := index.Search(c, query, opts); ; {
		id, err := t.Next(nil)
		if err == search.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		val, _ := strconv.ParseInt(id, 10, 64)
		ids = append(ids, val)
	}
	return ids, nil
}

func GetParam(params map[string]string, name string, defValue string) string {
	val := strings.TrimSpace(params[name])
	if val != "" {
		return val
	} else {
		return defValue
	}
}
