package model

import (
	"context"
	"english-study/internal/model/bean"

	"github.com/zeromicro/go-zero/core/logx"
)

func (m *Model) InsertWord(ctx context.Context, word *bean.Word, userId *uint) (err error) {
	// READ COMMITTED 即可：word 表有 unique 约束防止并发重复，无需 SERIALIZABLE 的谓词锁开销
	tx := m.DB.Begin()
	defer func() {
		if err != nil {
			if txe := tx.Rollback().Error; txe != nil {
				logx.Errorf("Rollback failed: %v", txe)
			}
			return
		}
		// commit 出错必须传回给调用方, 否则会出现"以为成功实际丢数据"
		if txe := tx.Commit().Error; txe != nil {
			logx.Errorf("Commit failed: %v", txe)
			err = txe
		}
	}()
	if err = tx.Table(word.UserTableName(userId)).WithContext(ctx).Create(word).Error; err != nil {
		return err
	}
	for _, pos := range word.Pos {
		pos.WordID = word.ID
	}
	if err = tx.Table((&bean.WordPos{}).UserTableName(userId)).WithContext(ctx).Create(word.Pos).Error; err != nil {
		return err
	}
	return nil
}

// 根据单词获取字典表的单词(带词性)
func (m *Model) GetWordWithPosByWord(ctx context.Context, word string, userId *uint) (*bean.Word, error) {
	bw := &bean.Word{}
	err := m.DB.Table(bw.UserTableName(userId)).Where("word = ?", word).WithContext(ctx).Take(bw).Error
	if err != nil {
		return nil, err
	}
	wps := make([]*bean.WordPos, 0)
	err = m.DB.Table((&bean.WordPos{}).UserTableName(userId)).Where("word_id = ?", bw.ID).WithContext(ctx).Find(&wps).Error
	if err != nil {
		return nil, err
	}
	bw.Pos = wps
	return bw, nil
}

// 根据单词ID获取单词表的单词(带词性)
func (m *Model) GetWordWithPosById(ctx context.Context, wordId uint, userId *uint) (*bean.Word, error) {
	bw := &bean.Word{}
	err := m.DB.Table(bw.UserTableName(userId)).Where("id = ?", wordId).WithContext(ctx).Take(bw).Error
	if err != nil {
		return nil, err
	}
	wps := make([]*bean.WordPos, 0)
	err = m.DB.Table((&bean.WordPos{}).UserTableName(userId)).Where("word_id = ?", wordId).WithContext(ctx).Find(&wps).Error
	if err != nil {
		return nil, err
	}
	bw.Pos = wps
	return bw, nil
}

func (m *Model) GetWordPos(ctx context.Context, wpId uint, userId *uint) (*bean.WordPos, error) {
	wp := &bean.WordPos{}
	err := m.DB.Table((&bean.WordPos{}).UserTableName(userId)).Where("id = ?", wpId).WithContext(ctx).Take(wp).Error
	if err != nil {
		return nil, err
	}
	return wp, nil
}
