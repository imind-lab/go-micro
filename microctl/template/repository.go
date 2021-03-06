/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package template

import (
	"os"
	"text/template"
)

// 生成repository
func CreateRepository(data *Data) error {
	var tpl = `/**
 *  MindLab
 *
 *  Create by songli on {{.Date}}
 *  Copyright © {{.Year}} imind.tech All rights reserved.
 */

package repository

import (
	"context"
	"{{.Domain}}/{{.Project}}/{{.Service}}/domain/{{.Service}}/repository/model"
)

type {{.Svc}}Repository interface {
	Create{{.Svc}}(ctx context.Context, m model.{{.Svc}}) (model.{{.Svc}}, error)

	Get{{.Svc}}ById(ctx context.Context, id int32, opt ...{{.Svc}}ByIdOption) (model.{{.Svc}}, error)

	Find{{.Svc}}ById(ctx context.Context, id int32) (model.{{.Svc}}, error)
	Get{{.Svc}}List(ctx context.Context, status, lastId, pageSize, page int32) ([]model.{{.Svc}}, int, error)

	Update{{.Svc}}Status(ctx context.Context, id, status int32) (int64, error)
	Update{{.Svc}}Count(ctx context.Context, id, num int32, column string) (int64, error)

	Delete{{.Svc}}ById(ctx context.Context, id int32) (int64, error)
}

`

	t, err := template.New("repository").Parse(tpl)
	if err != nil {
		return err
	}

	t.Option()
	dir := "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/domain/" + data.Service + "/repository/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName := dir + "repository.go"

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	tpl = `package repository

import (
	"time"
)

type {{.Svc}}ByIdOptions struct {
	RandExpire time.Duration
}

func New{{.Svc}}ByIdOptions(randExpire time.Duration) *{{.Svc}}ByIdOptions {
	return &{{.Svc}}ByIdOptions{RandExpire: randExpire}
}

type {{.Svc}}ByIdOption func(*{{.Svc}}ByIdOptions)

func {{.Svc}}ByIdRandExpire(expire time.Duration) {{.Svc}}ByIdOption {
	return func(o *{{.Svc}}ByIdOptions) {
		o.RandExpire = expire
	}
}

`

	t, err = template.New("repoption").Parse(tpl)
	if err != nil {
		return err
	}

	t.Option()
	dir = "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/domain/" + data.Service + "/repository/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName = dir + "options.go"

	f, err = os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	tpl = `/**
 *  MindLab
 *
 *  Create by songli on {{.Date}}
 *  Copyright © {{.Year}} imind.tech All rights reserved.
 */

package persistence

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	errorsx "github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"{{.Domain}}/{{.Project}}/{{.Service}}/domain/{{.Service}}/repository"
	"{{.Domain}}/{{.Project}}/{{.Service}}/domain/{{.Service}}/repository/model"
	"{{.Domain}}/{{.Project}}/{{.Service}}/infrastructure/pkg/constant"
	utilx "{{.Domain}}/{{.Project}}/{{.Service}}/infrastructure/pkg/util"
	"{{.Domain}}/{{.Project}}/micro/dao"
	redisx "{{.Domain}}/{{.Project}}/micro/redis"
	"{{.Domain}}/{{.Project}}/micro/util"
)

type {{.Service}}Repository struct {
	dao.Dao
}

//New{{.Svc}}Repository 创建用户仓库实例
func New{{.Svc}}Repository() repository.{{.Svc}}Repository {
	rep := dao.NewDao(constant.DBName)
	repo := {{.Service}}Repository{
		Dao: rep,
	}
	return repo
}

func (repo {{.Service}}Repository) Create{{.Svc}}(ctx context.Context, m model.{{.Svc}}) (model.{{.Svc}}, error) {
	if err := repo.DB(ctx).Create(&m).Error; err != nil {
		return m, errorsx.Wrap(err, "{{.Service}}Repository.Create{{.Svc}}")
	}
	repo.Cache{{.Svc}}(ctx, m)
	return m, nil
}

func (repo {{.Service}}Repository) Cache{{.Svc}}(ctx context.Context, m model.{{.Svc}}) error {
	key := utilx.CacheKey("{{.Service}}_", strconv.Itoa(int(m.Id)))
	expire := constant.CacheMinute5
	redisx.SetHashTable(ctx, repo.Redis(), key, m, expire)
	return nil
}

func (repo {{.Service}}Repository) Get{{.Svc}}ById(ctx context.Context, id int32, opt ...repository.{{.Svc}}ByIdOption) (model.{{.Svc}}, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "{{.Service}}Repository"), zap.String("func", "Get{{.Svc}}ById"))

	opts := repository.New{{.Svc}}ByIdOptions(util.RandDuration(120))
	for _, o := range opt {
		o(opts)
	}

	var m model.{{.Svc}}
	key := utilx.CacheKey("{{.Service}}_", strconv.Itoa(int(id)))
	err := redisx.HGet(ctx, repo.Redis(), key, &m)
	logger.Debug("redis.HGetAll", zap.Any("{{.Service}}", m), zap.String("key", key), zap.Error(err))
	if err == nil {
		return m, nil
	}

	m, err = repo.Find{{.Svc}}ById(ctx, id)
	if err != nil {
		return m, errorsx.WithMessage(err, "{{.Service}}Repository.Get{{.Svc}}ById")
	}

	expire := constant.CacheMinute5 + opts.RandExpire
	if m.IsEmpty() {
		expire = constant.CacheMinute1
	}
	redisx.SetHashTable(ctx, repo.Redis(), key, m, expire)
	return m, nil
}

func (repo {{.Service}}Repository) Find{{.Svc}}ById(ctx context.Context, id int32) (model.{{.Svc}}, error) {
	var m model.{{.Svc}}
	err := repo.DB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return m, nil
		}
		return m, errorsx.Wrap(err, "{{.Service}}Repository.Find{{.Svc}}ById")
	}
	return m, nil
}

func (repo {{.Service}}Repository) Get{{.Svc}}sCount(ctx context.Context, status int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "{{.Service}}Repository"), zap.String("func", "Get{{.Svc}}sCount"))

	key := utilx.CacheKey("{{.Service}}_cnt_", strconv.Itoa(int(status)))
	cnt, err := redisx.GetNumber(ctx, repo.Redis(), key)
	if err == nil {
		return cnt, nil
	}
	cnt, err = repo.Find{{.Svc}}sCount(ctx, status)
	if err != nil {
		return 0, errorsx.WithMessage(err, "{{.Service}}Repository.Get{{.Svc}}sCount")
	}
	err = repo.Redis().Set(ctx, key, cnt, constant.CacheMinute5).Err()
	if err != nil {
		logger.Error("redis.Set", zap.String("key", key), zap.Error(err))
	}
	return cnt, nil
}

func (repo {{.Service}}Repository) Find{{.Svc}}sCount(ctx context.Context, status int32) (int64, error) {
	var count int64
	tx := repo.DB(ctx).Model(model.{{.Svc}}{}).Select("count(id)")
	tx = tx.Where("status=?", status)
	if err := tx.Count(&count).Error; err != nil {
		return 0, errorsx.Wrap(err, "{{.Service}}Repository.Find{{.Svc}}sCount")
	}
	return count, nil
}

func (repo {{.Service}}Repository) Get{{.Svc}}List(ctx context.Context, status, lastId, pageSize, page int32) ([]model.{{.Svc}}, int, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "{{.Service}}Repository"), zap.String("func", "Get{{.Svc}}List"))

	ids, cnt, err := repo.Get{{.Svc}}ListIds(ctx, status, lastId, pageSize, page)
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "{{.Service}}Repository.Get{{.Svc}}List.Get{{.Svc}}ListIds")
	}

	ctx1, cancel := context.WithTimeout(ctx, constant.CRequestTimeout)
	defer cancel()

	{{.Service}}s, err := repo.Get{{.Svc}}List4Concurrent(ctx1, ids, repo.Get{{.Svc}}ById)
	logger.Debug("Get{{.Svc}}List4Concurrent", zap.Any("{{.Service}}s", {{.Service}}s), zap.Error(err))
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "{{.Service}}Repository.Get{{.Svc}}List.Get{{.Svc}}List4Concurrent")
	}
	return {{.Service}}s, cnt, nil
}

func (repo {{.Service}}Repository) Get{{.Svc}}ListIds(ctx context.Context, status, lastId, pageSize, page int32) ([]int32, int, error) {
	key := utilx.CacheKey("{{.Service}}_ids_", strconv.Itoa(int(status)))

	ids, cnt, err := redisx.ZRevRangeWithCard(ctx, repo.Redis(), key, lastId, pageSize, page)
	if err == nil {
		return ids, cnt, nil
	}

	ids, args, err := repo.Find{{.Svc}}ListIds(ctx, status, lastId, pageSize)
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "{{.Service}}Repository.Get{{.Svc}}List")
	}
	expire := constant.CacheMinute5 + util.RandDuration(120)
	redisx.SetSortedSet(ctx, repo.Redis(), key, args, expire)
	return ids, len(args), nil
}

func (repo {{.Service}}Repository) Find{{.Svc}}ListIds(ctx context.Context, status, lastId, pageSize int32) ([]int32, []*redis.Z, error) {

	tx := repo.DB(ctx).Model(model.{{.Svc}}{}).Select("id")
	tx = tx.Where("status=?", status)
	tx = tx.Order("id DESC")
	rows, err := tx.Rows()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []int32{}, []*redis.Z{}, nil
		}
		return nil, nil, errorsx.Wrap(err, "{{.Service}}Repository.Find{{.Svc}}ListIds.Rows")
	}
	defer rows.Close()

	var ids []int32
	var args []*redis.Z
	for rows.Next() {
		var (
			id int32
		)
		err = rows.Scan(&id)
		if err != nil {
			return nil, nil, errorsx.Wrap(err, "{{.Service}}Repository.Find{{.Svc}}ListIds.Scan")
		}

		check := false
		if lastId == 0 {
			check = true
		} else if lastId > id {
			check = true
		}
		if check {
			if len(ids) < int(pageSize) {
				ids = append(ids, id)
			}
		}
		args = append(args, &redis.Z{Score: float64(id), Member: id})
	}
	if err = rows.Err(); err != nil {
		return nil, nil, errorsx.Wrap(err, "{{.Service}}Repository.Find{{.Svc}}ListIds.Err")
	}
	return ids, args, nil
}

func (repo {{.Service}}Repository) Get{{.Svc}}List4Concurrent(ctx context.Context, ids []int32, fn func(context.Context, int32, ...repository.{{.Svc}}ByIdOption) (model.{{.Svc}}, error)) ([]model.{{.Svc}}, error) {
	var wg sync.WaitGroup

	count := len(ids)
	outputs := make([]*concurrent{{.Svc}}Output, count)
	wg.Add(count)

	for idx, id := range ids {
		go func(idx int, id int32, wg *sync.WaitGroup) {
			defer wg.Done()
			{{.Service}}, err := fn(ctx, id)
			outputs[idx] = &concurrent{{.Svc}}Output{
				object: {{.Service}},
				err:    err,
			}
		}(idx, id, &wg)
	}
	wg.Wait()

	{{.Service}}s := make([]model.{{.Svc}}, 0, count)
	for _, output := range outputs {
		if output.err == nil {
			{{.Service}}s = append({{.Service}}s, output.object)
		}
	}
	return {{.Service}}s, nil
}

type concurrent{{.Svc}}Output struct {
	object model.{{.Svc}}
	err    error
}

func (repo {{.Service}}Repository) Update{{.Svc}}Status(ctx context.Context, id, status int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "{{.Service}}Repository"), zap.String("func", "Update{{.Svc}}Status"))

	logger.Debug("invoke info", zap.Int32("id", id), zap.Int32("status", status))
	tx := repo.DB(ctx).Model(model.{{.Svc}}{}).Where("id = ?", id)
	tx = tx.Update("status", status)
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "{{.Service}}Repository.Update{{.Svc}}Status")
	}
	key := utilx.CacheKey("{{.Service}}_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	if err != nil {
		logger.Warn("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))
	}
	return tx.RowsAffected, nil
}

func (repo {{.Service}}Repository) Update{{.Svc}}Count(ctx context.Context, id, num int32, column string) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "{{.Service}}Repository"), zap.String("func", "Update{{.Svc}}Count"))

	logger.Debug("invoke info", zap.Int32("id", id), zap.Int32("num", num), zap.String("column", column))
	tx := repo.DB(ctx).Model(model.{{.Svc}}{}).Where("id = ?", id)
	tx = tx.Updates(map[string]interface{}{column: gorm.Expr(column+" + ?", num)})
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "{{.Service}}Repository.Update{{.Svc}}Count")
	}
	key := utilx.CacheKey("{{.Service}}_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	if err != nil {
		logger.Warn("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))
	}
	return tx.RowsAffected, nil
}

func (repo {{.Service}}Repository) Delete{{.Svc}}ById(ctx context.Context, id int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "{{.Service}}Repository"), zap.String("func", "Delete{{.Svc}}ById"))

	logger.Debug("invoke info", zap.Int32("id", id))
	tx := repo.DB(ctx).Delete(&model.{{.Svc}}{}, id)
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "{{.Service}}Repository.Delete{{.Svc}}ById")
	}
	key := utilx.CacheKey("{{.Service}}_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	logger.Debug("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))

	status := []int{0, 1}
	for _, s := range status {
		key := utilx.CacheKey("{{.Service}}_ids_", strconv.Itoa(s))
		err := repo.Redis().ZRem(ctx, key, id).Err()
		if err != nil {
			logger.Warn("redis.ZRem", zap.String("key", key), zap.Int32("id", id), zap.Error(err))
		}
	}

	return tx.RowsAffected, nil
}

`

	t, err = template.New("reppersistence").Parse(tpl)
	if err != nil {
		return err
	}

	t.Option()
	dir = "./" + data.Domain + "/" + data.Project + "/" + data.Service + "/domain/" + data.Service + "/repository/persistence/"

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	fileName = dir + data.Service + ".go"

	f, err = os.Create(fileName)
	if err != nil {
		return err
	}
	err = t.Execute(f, data)
	if err != nil {
		return err
	}
	f.Close()

	return nil
}
