package dbmodels

import (
	"github.com/google/uuid"
)

/*
Это структура слоя хранения данных.
Позже mapper будет преобразовывать её в доменную модель models.User
*/

type UserDTO struct {
	ID           uuid.UUID `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
}
