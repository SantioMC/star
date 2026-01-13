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

func (store *UserStore) GetAllUsers(page int, size int) ([]model.User, error) {
	var users []model.User

	if err := store.DB.Order("created_at DESC").Limit(size).Offset((page - 1) * size).Preload(clause.Associations).Find(&users).Error; err != nil {
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

func (store *UserStore) SearchUsers(query string, page int, size int) ([]model.User, error) {
	if strings.Contains(query, "%") {
		return []model.User{}, nil
	}

	search := "%" + query + "%"
	var users []model.User

	err := store.DB.
		Where("username LIKE ? OR email LIKE ? OR id LIKE ?", search, search, search).
		Order("created_at DESC").
		Limit(size).
		Offset((page - 1) * size).
		Preload(clause.Associations).
		Find(&users).Error

	return users, err
}

func (store *UserStore) SaveUser(user model.User) error {
	return store.DB.Unscoped().Transaction(func(tx *gorm.DB) error {
		if err := store.DB.Unscoped().
			Table("user_roles").
			Where("user_id = ?", user.ID).
			Delete(nil).Error; 
			err != nil {
				return err
			}

		_, err := gorm.G[model.User](tx).
			Where("id = ?", user.ID).
			Select("*").
			Updates(context.Background(), user)
		
		return err
	})
}

func (store *UserStore) DeleteUser(userId string) error {
	_, err := gorm.G[model.User](store.DB).Where("id = ?", userId).Delete(context.Background())
	return err
}

func (store *UserStore) SetLastLogin(userId string, time int64) error {
	_, err := gorm.G[model.User](store.DB).Where("id = ?", userId).Update(context.Background(), "last_login_at", time)
	return err
}

func (store *UserStore) SaveAccount(account *model.Account) error {
	return store.DB.Where("user_id = ? AND provider_id = ?", account.UserID, account.ProviderID).
		Assign(model.Account{ PlatformAvatar: account.PlatformAvatar, Username: account.Username }).
		Attrs(account).
		FirstOrCreate(&account).Error
}

func (store *UserStore) UpdateAccountPassword(accountId string, password string) error {
	_, err := gorm.G[model.Account](store.DB).Where("id = ?", accountId).Update(context.Background(), "password", password)
	return err
}

func (store *UserStore) GetCount() int64 {
	users, err := gorm.G[model.User](store.DB).Count(context.Background(), "*")
	if err != nil {
		return 0
	}

	return users
}

func (store *UserStore) SetMetadata(userId string, metadata *model.UserMetadata) error {
	_, err := gorm.G[model.User](store.DB).Where("id = ?", userId).Update(context.Background(), "metadata", metadata)
	return err
}

func (store *UserStore) SuspendUser(userId string, issuerId string, reason string, duration *int64) error {
	suspension := model.NewSuspension(userId, issuerId, reason, duration)
	return gorm.G[model.Suspension](store.DB).Create(context.Background(), &suspension)
}

func (store *UserStore) UnsuspendUser(userId string) error {
	_, err := gorm.G[model.Suspension](store.DB).Where("user_id = ?", userId).Delete(context.Background())
	return err
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