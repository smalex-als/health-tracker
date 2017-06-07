package dao

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smalex-als/health-tracker/server/common"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/search"
)

type Dao interface {
	IndexName() string
	GetAll(ctx context.Context, params map[string]string) (interface{}, error)
	Get(ctx context.Context, id int64, dst interface{}) error
	Put(ctx context.Context, src interface{}) error
	Delete(ctx context.Context, id int64, src interface{}) error
}

type BaseDao struct {
	TypeDesc *TypeDesc
}

var _ Dao = &BaseDao{}

func (s *BaseDao) IndexName() string {
	return s.TypeDesc.Type.IndexName
}

func (dao *BaseDao) Get(ctx context.Context, id int64, dst interface{}) error {
	key := datastore.NewKey(ctx, dao.TypeDesc.Type.Kind, "", id, nil)
	if err := common.DbGetCached(ctx, key, dst); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return common.ErrNotFound
		} else {
			return err
		}
	}
	return nil
}

func (dao *BaseDao) GetAll(ctx context.Context, params map[string]string) (interface{}, error) {
	limit := dao.ValidateLimit(dao.GetParamInt(params, "limit", 50))
	offset := dao.ValidateOffset(dao.GetParamInt(params, "offset", 0))
	slice := common.MakePtrSlice(dao.TypeDesc.Type.Kind, 0)
	q := datastore.NewQuery(dao.TypeDesc.Type.Kind).Order(dao.TypeDesc.ListView.Sort).Offset(offset).Limit(limit)
	if err := common.DbGetAll(ctx, q, slice); err != nil {
		return nil, err
	}
	return slice, nil
}

func (dao *BaseDao) Put(ctx context.Context, dst interface{}) error {
	var err error
	var id int64
	id, err = common.GetValueInt64(dst, "Id")
	if err != nil {
		return err
	}
	if id == 0 {
		id, _, err = datastore.AllocateIDs(ctx, dao.TypeDesc.Type.Kind, nil, 1)
		if err != nil {
			return err
		}
		if err = common.SetValueInt64(dst, "Id", id); err != nil {
			return err
		}
	}
	key := datastore.NewKey(ctx, dao.TypeDesc.Type.Kind, "", id, nil)
	if key, err = common.DbPutCached(ctx, key, dst); err != nil {
		return err
	}
	return nil
}

func (dao *BaseDao) Delete(ctx context.Context, id int64, src interface{}) error {
	log.Infof(ctx, "Delete %s", dao.IndexName())
	if dao.IndexName() != "" {
		if index, err := search.Open(dao.IndexName()); err != nil {
			return err
		} else {
			if err := index.Delete(ctx, strconv.FormatInt(id, 10)); err != nil {
				return err
			}
		}
	}
	key := datastore.NewKey(ctx, dao.TypeDesc.Type.Kind, "", id, nil)
	return common.DbDeleteCached(ctx, key)
}

func (dao *BaseDao) ValidateOffset(offset int) int {
	if offset < 0 {
		offset = 0
	}
	if offset > 1000 {
		offset = 1000
	}
	return offset
}

func (dao *BaseDao) ValidateLimit(limit int) int {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	return limit
}

func (dao *BaseDao) GetFormParamInt64(c *gin.Context, name string, defValue int64) int64 {
	val := strings.TrimSpace(c.Query(name))
	res, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		return res
	} else {
		return defValue
	}
}

func (dao *BaseDao) GetParamInt(params map[string]string, name string, defValue int) int {
	val := strings.TrimSpace(params[name])
	res, err := strconv.ParseInt(val, 10, 32)
	if err == nil {
		return int(res)
	} else {
		return defValue
	}
}

func (dao *BaseDao) GetParam(params map[string]string, name string, defValue string) string {
	val := strings.TrimSpace(params[name])
	if val != "" {
		return val
	} else {
		return defValue
	}
}
