package model

import (
	"context"
	"english-study/internal/model/bean"

	"github.com/zeromicro/go-zero/core/logx"
)

func (m *Model) CreateUser(ctx context.Context, user *bean.User) (err error) {
	// 1. 用户行写入: 用事务保证回滚一致性
	tx := m.DB.Begin()
	defer func() {
		if err != nil {
			if txe := tx.Rollback().Error; txe != nil {
				logx.Errorf("Rollback failed: %v", txe)
			}
			return
		}
		if txe := tx.Commit().Error; txe != nil {
			logx.Errorf("Commit failed: %v", txe)
			err = txe
		}
	}()

	if err = tx.Table("users").WithContext(ctx).Create(user).Error; err != nil {
		return err
	}

	// 2. 用户专属表的 DDL (AutoMigrate) 不能放在 tx 里:
	//   - GORM AutoMigrate 在 PG 上每条 DDL 自己一个 implicit commit;
	//   - 强行用 tx 会出现"用户行未 commit 时已经 ALTER 了 schema"的诡异半状态.
	// 这里的取舍: DDL 失败时, 用户行也回滚, 保证不会出现"用户存在但表没建" — 表 + 用户必须共存亡.
	userId := user.ID

	wordTable := (&bean.Word{}).UserTableName(&userId)
	if err = m.DB.Table(wordTable).AutoMigrate(&bean.Word{}); err != nil {
		return err
	}

	posTable := (&bean.WordPos{}).UserTableName(&userId)
	if err = m.DB.Table(posTable).AutoMigrate(&bean.WordPos{}); err != nil {
		return err
	}

	// **必须**: 之前漏建 word_phrase_user_{uid}, 新注册用户首次写短语就 500
	phraseTable := (&bean.WordPhrase{}).UserTableName(&userId)
	if err = m.DB.Table(phraseTable).AutoMigrate(&bean.WordPhrase{}); err != nil {
		return err
	}

	return nil
}
