package mappers

import (
	"project03/internal/datasource/dbmodels"
	"project03/internal/domain/models"
)

func UserToDomain(dto dbmodels.UserDTO) models.User {
	return models.User{
		ID:           dto.ID,
		Login:        dto.Login,
		PasswordHash: dto.PasswordHash,
	}
}

func UserToDTO(user models.User) dbmodels.UserDTO {
	return dbmodels.UserDTO{
		ID:           user.ID,
		Login:        user.Login,
		PasswordHash: user.PasswordHash,
	}
}
