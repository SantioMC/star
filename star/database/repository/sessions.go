package repository

import (
	"context"
	"errors"
	"time"

	"github.com/tubblelive/star/star/database/model"
	"gorm.io/gorm"
)

type SessionStore struct {
	DB *gorm.DB
}

func (store SessionStore) GetSessionByID(id string) (*model.Session, error) {
	session, err := gorm.G[model.Session](store.DB).Where("id = ?", id).First(context.Background())

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &session, nil
}

func (store SessionStore) GetUserSessions(userId string) ([]model.Session, error) {
	return gorm.G[model.Session](store.DB).Where("user_id = ?", userId).Find(context.Background())
}

func (store SessionStore) CreateSession(session *model.Session) error {
	return gorm.G[model.Session](store.DB).Create(context.Background(), session)
}

func (store SessionStore) InvalidateSession(id string) error {
	_, err := gorm.G[model.Session](store.DB).Where("id = ?", id).Update(context.Background(), "invalidated_at", time.Now().Unix())
	return err
}

func (store SessionStore) UpdateSessionFingerprint(id string, fingerprint *string) error {
	_, err := gorm.G[model.Session](store.DB).Where("id = ?", id).Update(context.Background(), "fingerprint", fingerprint)
	return err
}

func (store SessionStore) InvalidateAllSessions(userId string) error {
	_, err := gorm.G[model.Session](store.DB).Where("user_id = ?", userId).Update(context.Background(), "invalidated_at", time.Now().Unix())
	return err
}

func (store *SessionStore) GetCount() int64 {
	sessions, err := gorm.G[model.Session](store.DB).Count(context.Background(), "*")
	if err != nil {
		return 0
	}

	return sessions
}

func (store SessionStore) DeleteOldSessions() error {
	invalidatedAt := time.Now().Add(-time.Hour * 24)
	createdAt := time.Now().Add(-time.Hour * 24 * 14)

	_, err := gorm.G[model.Session](store.DB.Unscoped()).
		Where("invalidated_at <= ? OR created_at <= ?", invalidatedAt.Unix(), createdAt.Unix()).
		Delete(context.Background())
	
	return err
}