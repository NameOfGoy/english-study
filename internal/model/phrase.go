package model

import (
	"context"
	"english-study/internal/model/bean"
)

func (m *Model) InsertWordPhrase(ctx context.Context, w *bean.WordPhrase, userId *uint) error {
	return m.DB.WithContext(ctx).Table(w.UserTableName(userId)).Create(w).Error
}

func (m *Model) GetWordPhraseById(ctx context.Context, id uint, userId *uint) (*bean.WordPhrase, error) {
	var w bean.WordPhrase
	err := m.DB.WithContext(ctx).Table(w.UserTableName(userId)).Where("id = ?", id).Take(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *Model) GetWordPhraseByPhrase(ctx context.Context, phrase string, userId *uint) (*bean.WordPhrase, error) {
	var w bean.WordPhrase
	err := m.DB.WithContext(ctx).Table(w.UserTableName(userId)).Where("phrase = ?", phrase).Take(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}
