package session

type BeforeQuery[T any] interface {
	BeforeQuery(s *DB[T]) error
}
type AfterQuery[T any] interface {
	AfterQuery(s *DB[T]) error
}
type BeforeUpdate[T any] interface {
	BeforeUpdate(s *DB[T]) error
}
type AfterUpdate[T any] interface {
	AfterUpdate(s *DB[T]) error
}
type BeforeDelete[T any] interface {
	BeforeDelete(s *DB[T]) error
}
type AfterDelete[T any] interface {
	AfterDelete(s *DB[T]) error
}
type BeforeInsert[T any] interface {
	BeforeInsert(s *DB[T]) error
}
type AfterInsert[T any] interface {
	AfterInsert(s *DB[T]) error
}
