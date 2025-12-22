package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/tubblelive/star/star/config"
	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserStore struct {
	DB     *gorm.DB
	Config *config.Config
}

func (store *UserStore) GetAllUsers() ([]model.User, error) {
	var users []model.User

	if err := store.DB.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (store *UserStore) GetUserByID(id string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("id = ?", id).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsername(username string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("username = ?", username).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByEmail(email string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("email = ?", strings.ToLower(email)).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		
		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsernameOrEmail(query string) (*model.User, error) {
	user := model.User{}

	if err := store.DB.Where("email = ? OR username = ?", query, query).Preload(clause.Associations).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (store *UserStore) GetActiveSuspension(id string) (*model.Suspension, error) {
	suspension := model.Suspension{}

	if err := store.DB.Where("user_id = ?", id).First(&suspension).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &suspension, nil
}

func (store *UserStore) SaveUser(user model.User) error {
	_, err := gorm.G[model.User](store.DB).Where("id = ?", user.ID).Updates(context.Background(), user)
	return err
}

func (store *UserStore) SaveAccount(account *model.Account) error {
	return store.DB.Where("user_id = ? AND provider_id = ?", account.UserID, account.ProviderID).
		Assign(model.Account{ PlatformAvatar: account.PlatformAvatar, Username: account.Username }).
		Attrs(account).
		FirstOrCreate(&account).Error
}

func (store *UserStore) GetUserPermissions(user *model.User) ([]model.Permission, error) {
	// var roleIds []string
	// for _, role := range user.Roles {
	// 	roleIds = append(roleIds, role.ID)
	// }

	// var permissions []string
	// return store.DB.Table("role_permissions").Joi.Select().Find(&permissions)
	// todo: this!!!
	return []model.Permission{}, nil
}