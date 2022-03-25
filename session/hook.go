package session

import "context"

type BeforeQuery[T any] interface {
	BeforeQuery(ctx context.Context, db *DB[T]) error
}

type AfterQuery[T any] interface {
	AfterQuery(ctx context.Context, db *DB[T]) error
}

type BeforeUpdate[T any] interface {
	BeforeUpdate(ctx context.Context, db *DB[T]) error
}

type AfterUpdate[T any] interface {
	AfterUpdate(ctx context.Context, db *DB[T]) error
}

type BeforeDelete[T any] interface {
	BeforeDelete(ctx context.Context, db *DB[T]) error
}

type AfterDelete[T any] interface {
	AfterDelete(ctx context.Context, db *DB[T]) error
}

type BeforeInsert[T any] interface {
	BeforeInsert(ctx context.Context, db *DB[T]) error
}

type AfterInsert[T any] interface {
	AfterInsert(ctx context.Context, db *DB[T]) error
}

type BeforeExecute[T any] interface {
	BeforeExecute(ctx context.Context, db *DB[T])
}

type AfterExecute[T any] interface {
	AfterExecute(ctx context.Context, db *DB[T])
}
