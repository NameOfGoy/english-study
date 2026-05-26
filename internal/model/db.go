package model

import (
	"english-study/internal/eventbus"
	"english-study/internal/model/bean"
	"english-study/internal/model/dto"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Model struct {
	DB  *gorm.DB
	Gen *dto.Query
}

// NewModel returns a model for the database table.
func NewModel(dsn string) (*Model, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(bean.Schemas...)
	if err != nil {
		return nil, err
	}
	// 每个用户一份自己的词典表
	if err := userWordTableSync(db); err != nil {
		return nil, err
	}
	// 启动时孤儿清理: 上次进程如果在 PublishAsync(tag.deleted) 后崩了, 这里兜底清掉 word_tags 残留
	if err := eventbus.ReapOrphanWordTags(db); err != nil {
		return nil, err
	}
	// 注册 eventbus 订阅者 (tag 删除级联清理等)
	if err := eventbus.RegisterSubscribers(db); err != nil {
		return nil, err
	}
	return &Model{
		DB:  db,
		Gen: dto.Use(db),
	}, nil
}

func userWordTableSync(db *gorm.DB) error {
	// 查询所有用户
	var users []bean.User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	// 为每个用户创建专属表
	for _, user := range users {
		userId := user.ID

		// 移植word_user_$userId表
		wordTable := (&bean.Word{}).UserTableName(&userId)
		if err := db.Table(wordTable).AutoMigrate(&bean.Word{}); err != nil {
			return err
		}

		// 移植word_pos_user_$userId表
		posTable := (&bean.WordPos{}).UserTableName(&userId)
		if err := db.Table(posTable).AutoMigrate(&bean.WordPos{}); err != nil {
			return err
		}
		
		// 移植word_phrase_user_$userId表
		phraseTable := (&bean.WordPhrase{}).UserTableName(&userId)
		if err := db.Table(phraseTable).AutoMigrate(&bean.WordPhrase{}); err != nil {
			return err
		}
	}

	return nil
}
