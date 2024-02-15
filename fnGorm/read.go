package fnGorm

import (
	"context"
	"gorm.io/gorm"
)

type (
	IFindQuery[T any] interface {
		Query(tx *gorm.DB) *gorm.DB
	}

	IPager interface {
		Page() int64
		Size() int64
	}

	ResultList[T any] struct {
		Page  int64
		Size  int64
		Total int64
		List  []*T
	}
)

// FindOne
// T 는 반드시 struct 타입이어야 한다. 포인터는 안됨
func FindOne[T any](
	ctx context.Context,
	i IFindQuery[T],
) (res *T, err error) {
	var rows *gorm.DB
	var query = GetDBP(ctx).Model(new(T))
	query = i.Query(query)

	res = new(T)
	if rows = query.Take(res); rows.Error != nil {
		err = rows.Error
		return
	}

	return
}

func FindAll[T any](
	ctx context.Context,
	i IFindQuery[T],
) (ls []*T, err error) {
	ls = make([]*T, 0)
	var query = GetDBP(ctx).Model(new(T))
	query = i.Query(query)

	var rows *gorm.DB
	if rows = query.Find(&ls); rows.Error != nil {
		err = rows.Error
		return
	}

	return
}

func FindList[T any](
	ctx context.Context,
	i IFindQuery[T],
	p IPager,
) (res *ResultList[T], err error) {
	res = &ResultList[T]{
		Page:  p.Page(),
		Size:  p.Size(),
		Total: 0,
		List:  make([]*T, 0),
	}

	var query = GetDBP(ctx).Model(new(T))
	query = i.Query(query)

	var rows *gorm.DB

	if rows = query.Count(&res.Total); rows.Error != nil {
		err = rows.Error
		return
	}

	if rows = query.
		Offset(int(p.Size() * p.Page())).
		Limit(int(p.Size())).
		Find(&res.List); rows.Error != nil {
		err = rows.Error
		return
	}

	return
}
