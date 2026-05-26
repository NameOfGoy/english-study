package eventbus

import (
	"sync/atomic"
	"testing"
	"time"

	resbus "github.com/NameOfGoy/resource-eventbus"

	asaskbus "github.com/asaskevich/EventBus"
)

// resetBus 给每个测试一个全新的 bus, 避免测试间订阅相互污染
func resetBus(t *testing.T) {
	t.Helper()
	resbus.SetBus(asaskbus.New())
}

func TestTagDeleted_RoundTrip_SystemTag(t *testing.T) {
	resetBus(t)
	var got TagDeletedPayload
	var calls int32
	err := TagDeleted.Subscribe(func(p TagDeletedPayload) {
		got = p
		atomic.AddInt32(&calls, 1)
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}

	TagDeleted.Publish(TagDeletedPayload{
		TagID:    42,
		IsSystem: true,
		// UserID 在系统标签场景下被订阅者忽略, 但发送方传 0 是规范
	})

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("handler should fire once, got %d", calls)
	}
	if got.TagID != 42 || !got.IsSystem {
		t.Errorf("payload mismatch: %+v", got)
	}
}

func TestTagDeleted_RoundTrip_UserTag(t *testing.T) {
	resetBus(t)
	var got TagDeletedPayload
	_ = TagDeleted.Subscribe(func(p TagDeletedPayload) { got = p })

	TagDeleted.Publish(TagDeletedPayload{
		TagID:    7,
		IsSystem: false,
		UserID:   123,
	})

	if got.TagID != 7 || got.IsSystem || got.UserID != 123 {
		t.Errorf("payload mismatch: %+v", got)
	}
}

func TestTagDeleted_PublishAsync_Eventually(t *testing.T) {
	resetBus(t)
	done := make(chan struct{})
	_ = TagDeleted.SubscribeAsync(func(_ TagDeletedPayload) { close(done) }, false)

	TagDeleted.PublishAsync(TagDeletedPayload{TagID: 1, IsSystem: true})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("async handler never fired")
	}
}
