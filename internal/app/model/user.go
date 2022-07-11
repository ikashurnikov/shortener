package model

type UserID int

const (
	InvalidUserID = -1
)

func (id UserID) IsValid() bool {
	return id >= 0
}
