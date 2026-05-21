package model

import (
	"context"
	"english-study/internal/model/bean"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

// CreateUser 创建用户 + 用户专属三张表 (word_user_{uid}, word_pos_user_{uid}, word_phrase_user_{uid}).
//
// 复杂在 DDL 不能放进事务 (PG 的 CREATE TABLE 有 implicit commit), 我们的目标:
//   1. 任意一步失败时不能留"用户已建但没表" / "表已建但没用户" 的半状态
//   2. 失败后用户应能立刻重试注册成功 (账号未占用 + 旧 DDL 清理掉)
//
// 实现策略:
//   - tx.Create(user) 但不立即 commit (拿到 user.ID 即可)
//   - 在 tx 外串行 AutoMigrate 三张表, 用 createdTables 记录已建成功的
//   - 任何一步失败: 回滚用户 + DROP TABLE 已建的表 → 彻底回到空状态
//   - 都成功: 提交用户行
//   - 极端 race (commit 后进程崩在 DROP/DDL 之间) 由启动时的 userWordTableSync 兜底
func (m *Model) CreateUser(ctx context.Context, user *bean.User) (err error) {
	tx := m.DB.Begin()
	createdTables := make([]string, 0, 3)

	defer func() {
		if err != nil {
			if txe := tx.Rollback().Error; txe != nil {
				logx.Errorf("CreateUser rollback failed: %v", txe)
			}
			// 同步 DROP 掉已建的表, 避免孤儿; 失败也只能 log, 没法再做更多
			for _, t := range createdTables {
				if dropErr := m.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", t)).Error; dropErr != nil {
					logx.Errorf("CreateUser: 回滚 DROP TABLE %s 失败 (孤儿表, 等启动 reconcile): %v", t, dropErr)
				}
			}
			return
		}
		if txe := tx.Commit().Error; txe != nil {
			logx.Errorf("CreateUser commit failed: %v", txe)
			err = txe
			// commit 失败也走清理
			for _, t := range createdTables {
				if dropErr := m.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", t)).Error; dropErr != nil {
					logx.Errorf("CreateUser: commit 失败后 DROP TABLE %s 失败: %v", t, dropErr)
				}
			}
		}
	}()

	if err = tx.Table("users").WithContext(ctx).Create(user).Error; err != nil {
		return err
	}

	userId := user.ID

	wordTable := (&bean.Word{}).UserTableName(&userId)
	if err = m.DB.Table(wordTable).AutoMigrate(&bean.Word{}); err != nil {
		return fmt.Errorf("AutoMigrate %s failed: %w", wordTable, err)
	}
	createdTables = append(createdTables, wordTable)

	posTable := (&bean.WordPos{}).UserTableName(&userId)
	if err = m.DB.Table(posTable).AutoMigrate(&bean.WordPos{}); err != nil {
		return fmt.Errorf("AutoMigrate %s failed: %w", posTable, err)
	}
	createdTables = append(createdTables, posTable)

	phraseTable := (&bean.WordPhrase{}).UserTableName(&userId)
	if err = m.DB.Table(phraseTable).AutoMigrate(&bean.WordPhrase{}); err != nil {
		return fmt.Errorf("AutoMigrate %s failed: %w", phraseTable, err)
	}
	createdTables = append(createdTables, phraseTable)

	return nil
}
