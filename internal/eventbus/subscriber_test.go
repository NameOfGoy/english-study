package eventbus

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockDB 用 sqlmock 假 PG driver, 不打真库
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	rawDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 rawDB,
		PreferSimpleProtocol: true, // 让 sqlmock 期望和实际 SQL 直接对得上
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return gormDB, mock, rawDB
}

func TestTagDeletedHandler_SystemTag_DeletesAllUsers(t *testing.T) {
	db, mock, raw := newMockDB(t)
	defer raw.Close()

	// 系统标签删除应只用 tag_id 作条件, 不带 user_id
	mock.ExpectExec(`^DELETE FROM "word_tags" WHERE tag_id = \$1$`).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 17)) // 假装清了 17 行

	makeTagDeletedHandler(db)(TagDeletedPayload{
		TagID:    42,
		IsSystem: true,
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

func TestTagDeletedHandler_UserTag_ScopedToOwner(t *testing.T) {
	db, mock, raw := newMockDB(t)
	defer raw.Close()

	// 用户标签删除必须带上 user_id 否则会误清别人
	mock.ExpectExec(`^DELETE FROM "word_tags" WHERE tag_id = \$1 AND user_id = \$2$`).
		WithArgs(7, 123).
		WillReturnResult(sqlmock.NewResult(0, 3))

	makeTagDeletedHandler(db)(TagDeletedPayload{
		TagID:    7,
		IsSystem: false,
		UserID:   123,
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

func TestTagDeletedHandler_DBError_DoesNotPanic(t *testing.T) {
	db, mock, raw := newMockDB(t)
	defer raw.Close()

	mock.ExpectExec(`DELETE FROM "word_tags"`).
		WillReturnError(sql.ErrConnDone)

	// 必须无 panic; 只 log
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("handler panicked on DB error: %v", r)
		}
	}()
	makeTagDeletedHandler(db)(TagDeletedPayload{TagID: 1, IsSystem: true})
}

func TestReapOrphanWordTags(t *testing.T) {
	db, mock, raw := newMockDB(t)
	defer raw.Close()

	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM word_tags WHERE tag_id NOT IN (SELECT id FROM tags)",
	)).WillReturnResult(sqlmock.NewResult(0, 5))

	if err := ReapOrphanWordTags(db); err != nil {
		t.Fatalf("ReapOrphanWordTags: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}
