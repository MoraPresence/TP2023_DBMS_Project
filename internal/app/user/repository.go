package user

import "TP2023_DBMS_Project/internal/app/user/models"

type Repository interface {
	AddUser(user models.UserInfo) error

	GetUserByNickAndEmail(nickname, email string) ([]models.UserInfo, error)
	GetUserByNick(nickname string) (models.UserInfo, error)
	GetUsersByForumSlug(slug string, lim int, from string, desc bool) ([]models.UserInfo, error)

	UpdateUser(user models.UserInfo) (models.UserInfo, error)
}
