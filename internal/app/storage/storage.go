package storage

type Storage interface {
	Insert(value string) (uint32, error)
	Select(id uint32) (string, error)
}
