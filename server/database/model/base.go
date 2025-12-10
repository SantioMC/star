package model

import (
	"math/rand"

	"gorm.io/plugin/soft_delete"
)

var idAlphabet = []byte("0123456789abcdefghijklmnopqrstuvwxyz")

func CreateId(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = idAlphabet[rand.Intn(len(idAlphabet))]
	}
	return string(b)
}

func CreateTypedId(prefix string, length int) string {
	return prefix + "_" + CreateId(length)
}

type BaseModel struct {
	ID        string                `json:"id" gorm:"primaryKey"`
	CreatedAt int64                 `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64                 `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index"`
}
