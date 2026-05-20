package model

import (
	"context"
	"english-study/internal/model/bean"

	"github.com/zeromicro/go-zero/core/logx"
)

func (m *Model) CreateUser(ctx context.Context, user *bean.User) (err error) {
	// 开启事务
	tx := m.DB.Begin()
	defer func() {
		if err != nil {
			if txe := tx.Rollback().Error; txe != nil {
				logx.Errorf("Rollback failed: %v", txe)
			}
		} else {
			if txe := tx.Commit().Error; txe != nil {
				logx.Errorf("Commit failed: %v", txe)
			}
		}
	}()
	
	// 创建用户
	if err = tx.Table("users").WithContext(ctx).Create(user).Error; err != nil {
		return err
	}

	// 创建用户专属表
	userId := user.ID

	// 移植word_user_$userId表
	wordTable := (&bean.Word{}).UserTableName(&userId)
	if err := m.DB.Table(wordTable).AutoMigrate(&bean.Word{}); err != nil {
		return err
	}

	// 移植word_pos_user_$userId表
	posTable := (&bean.WordPos{}).UserTableName(&userId)
	if err := m.DB.Table(posTable).AutoMigrate(&bean.WordPos{}); err != nil {
		return err
	}

	return nil
}
