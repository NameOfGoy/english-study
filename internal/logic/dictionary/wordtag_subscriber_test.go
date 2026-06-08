package dictionary

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"english-study/internal/eventbus"

	"github.com/DATA-DOG/go-sqlmock"
	resbus "github.com/NameOfGoy/resource-eventbus"
	"github.com/NameOfGoy/resource-eventbus/generator"
	asaskbus "github.com/asaskevich/EventBus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockDBForWordTag 用 sqlmock 假 PG driver, 不打真库.
// 命名带后缀避免和同包别处可能的 newMockDB 撞名.
func newMockDBForWordTag(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

// resetBusForWordTag 给每个 round-trip 测试一个全新的底层 bus, 避免相互污染
func resetBusForWordTag(t *testing.T) {
	t.Helper()
	resbus.SetBus(asaskbus.New())
}

func TestTagDeletedHandler_SystemTag_DeletesAllUsers(t *testing.T) {
	db, mock, raw := newMockDBForWordTag(t)
	defer raw.Close()

	// 系统标签删除应只用 tag_id 作条件, 不带 user_id
	mock.ExpectExec(`^DELETE FROM "word_tags" WHERE tag_id = \$1$`).
		WithArgs(42).
		WillReturnResult(sqlmock.NewResult(0, 17))

	makeTagDeletedHandler(db)(eventbus.TagDeletedPayload{
		TagID:    42,
		IsSystem: true,
	}, nil)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

func TestTagDeletedHandler_UserTag_ScopedToOwner(t *testing.T) {
	db, mock, raw := newMockDBForWordTag(t)
	defer raw.Close()

	// 用户标签删除必须带上 user_id 否则会误清别人
	mock.ExpectExec(`^DELETE FROM "word_tags" WHERE tag_id = \$1 AND user_id = \$2$`).
		WithArgs(7, 123).
		WillReturnResult(sqlmock.NewResult(0, 3))

	makeTagDeletedHandler(db)(eventbus.TagDeletedPayload{
		TagID:    7,
		IsSystem: false,
		UserID:   123,
	}, nil)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}

func TestTagDeletedHandler_DBError_DoesNotPanic(t *testing.T) {
	db, mock, raw := newMockDBForWordTag(t)
	defer raw.Close()

	mock.ExpectExec(`DELETE FROM "word_tags"`).
		WillReturnError(sql.ErrConnDone)

	// 必须无 panic; 只 log
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("handler panicked on DB error: %v", r)
		}
	}()
	makeTagDeletedHandler(db)(eventbus.TagDeletedPayload{TagID: 1, IsSystem: true}, nil)
}

func TestReapOrphanWordTags(t *testing.T) {
	db, mock, raw := newMockDBForWordTag(t)
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

// TestSubscribeTagEvents_EndToEnd 验证 publish→subscribe 链路实际打到 handler,
// 且 handler 能正确发 SQL.
func TestSubscribeTagEvents_EndToEnd(t *testing.T) {
	resetBusForWordTag(t)
	db, mock, raw := newMockDBForWordTag(t)
	defer raw.Close()

	done := make(chan struct{}, 1)
	mock.ExpectExec(`^DELETE FROM "word_tags" WHERE tag_id = \$1$`).
		WithArgs(99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := eventbus.Tag().Deleted().SubscribeAsync(
		func(p eventbus.TagDeletedPayload, _ *generator.EventMetadata) {
			makeTagDeletedHandler(db)(p, nil)
			done <- struct{}{}
		}, false); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	eventbus.Tag().Deleted().PublishAsync(eventbus.TagDeletedPayload{
		TagID:    99,
		IsSystem: true,
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("async handler never fired")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("未满足的 SQL 期望: %v", err)
	}
}
