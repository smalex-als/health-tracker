package common

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

func DbGetMulti(c context.Context, q *datastore.Query, dst interface{}) error {
	_, err := q.GetAll(c, &dst)
	return err
	// t := q.Run(c)
	// dv := reflect.ValueOf(dst)
	// if dv.Kind() != reflect.Slice {
	// 	return errors.New("ErrInvalidEntityType")
	// }
	// // if dv.Kind() != reflect.Ptr || dv.IsNil() {
	// // 	return errors.New("ErrInvalidEntityType")
	// // }
	// datastore.GetMulti()
	// dv = dv.Elem()
	// for {
	// 	entity := reflect.New(dv.Type().Elem())
	// 	_, err := t.Next(entity.Interface())
	// 	if err == datastore.Done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		return fmt.Errorf("fetching next: %v", err)
	// 	}
	// 	dv.Set(reflect.Append(dv, entity.Elem()))
	// }
	// return nil
}

func DbGetSingle(c context.Context, q *datastore.Query, dst interface{}) error {
	t := q.Run(c)
	key, err := t.Next(dst)
	if err == datastore.Done {
		return err
	}
	if err != nil {
		log.Warningf(c, "Error: get query result %+v", err)
		return err
	}
	if key.IntID() != 0 {
		if err := SetValueInt64(dst, "Id", key.IntID()); err != nil {
			return err
		}
	}
	if key.IntID() != 0 {
		if err := SetValueInt64(dst, "Id", key.IntID()); err != nil {
			return err
		}
	}
	return nil
}

func DbGetCached(c context.Context, key *datastore.Key, dst interface{}) error {
	strKey := key.String()
	if _, err := memcache.JSON.Get(c, strKey, dst); err == nil {
		return nil
	} else if err != nil && err != memcache.ErrCacheMiss {
		log.Errorf(c, "Error: %s", err)
	}
	if err := datastore.Get(c, key, dst); err != nil {
		return err
	}
	if key.IntID() != 0 {
		if err := SetValueInt64(dst, "Id", key.IntID()); err != nil {
			return err
		}
	}
	return memcache.JSON.Set(c, &memcache.Item{
		Key:        strKey,
		Object:     dst,
		Expiration: 0,
	})
}

func DbDeleteCached(c context.Context, key *datastore.Key) error {
	strKey := key.String()
	if err := datastore.Delete(c, key); err != nil {
		return err
	}
	memcache.Delete(c, strKey)
	return nil
}

func DbPutCached(c context.Context, key *datastore.Key, dst interface{}) (*datastore.Key, error) {
	if newkey, err := datastore.Put(c, key, dst); err != nil {
		return newkey, err
	} else {
		strKey := newkey.String()
		return newkey, memcache.JSON.Set(c, &memcache.Item{
			Key:        strKey,
			Object:     dst,
			Expiration: 0,
		})
	}
}

func DbGet(c context.Context, key *datastore.Key, dst interface{}) {
	if err := datastore.Get(c, key, dst); err != nil {
		panic(err)
	}
}

func DbGetAll(c context.Context, q *datastore.Query, dst interface{}) error {
	if keys, err := q.GetAll(c, dst); err != nil {
		return err
	} else {
		s := reflect.ValueOf(dst)
		if s.Kind() == reflect.Ptr {
			s = s.Elem()
		}
		for i := 0; i < s.Len(); i++ {
			item := s.Index(i).Interface()
			if err := SetValueInt64(item, "Id", keys[i].IntID()); err != nil {
				return err
			}
		}
	}
	return nil
}

func DbGetByIds(c context.Context, kind string, ids []int64) (interface{}, error) {
	keys := make([]*datastore.Key, len(ids))
	for i, id := range ids {
		keys[i] = datastore.NewKey(c, kind, "", id, nil)
	}
	slice := makeSlice(kind, len(keys))
	if err := datastore.GetMulti(c, keys, slice.Interface()); err != nil {
		log.Warningf(c, "entities not found %+v", keys)
		return nil, err
	}

	s := reflect.ValueOf(slice.Interface())
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	for i := 0; i < s.Len(); i++ {
		SetValueInt64(s.Index(i).Interface(), "Id", keys[i].IntID())
	}
	// log.Infof(c, "getAllByIds = %+v", slice.Interface())

	return slice.Interface(), nil
}

func DbMustPut(c context.Context, key *datastore.Key, src interface{}) *datastore.Key {
	newkey, err := datastore.Put(c, key, src)
	if err != nil {
		panic(err)
	}
	return newkey
}

func SetValueInt64(dst interface{}, field string, value int64) error {
	s := reflect.ValueOf(dst)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	f := s.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("field id not found %+v %+T", dst, dst)
	}
	f.SetInt(value)
	return nil
}

func GetValueInt64(dst interface{}, field string) (int64, error) {
	s := reflect.ValueOf(dst)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	f := s.FieldByName(field)
	if !f.IsValid() {
		return 0, fmt.Errorf("field id not found %+v %+T", dst, dst)
	}
	return f.Int(), nil
}
