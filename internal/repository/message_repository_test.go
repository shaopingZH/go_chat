package repository

import (
	"strings"
	"testing"

	"go-chat/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "dryrun:dryrun@tcp(localhost:3306)/dryrun?charset=utf8mb4&parseTime=True&loc=Local",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}

	return db
}

func TestApplyHistoryKeywordFilter_ExcludesImageMessages(t *testing.T) {
	db := newDryRunDB(t)
	var messages []model.Message

	stmt := applyHistoryKeywordFilter(db.Model(&model.Message{}), "5").Find(&messages).Statement
	sql := stmt.SQL.String()

	if !strings.Contains(sql, "content LIKE ?") {
		t.Fatalf("expected keyword LIKE filter in SQL, got %q", sql)
	}
	if !strings.Contains(sql, "msg_type = ?") {
		t.Fatalf("expected text-message filter in SQL, got %q", sql)
	}

	if len(stmt.Vars) < 2 {
		t.Fatalf("expected at least 2 SQL vars, got %d (%v)", len(stmt.Vars), stmt.Vars)
	}
	if stmt.Vars[0] != 1 {
		t.Fatalf("expected first SQL var to enforce text msg_type=1, got %#v", stmt.Vars[0])
	}
	if stmt.Vars[1] != "%5%" {
		t.Fatalf("expected keyword SQL var %%5%%, got %#v", stmt.Vars[1])
	}
}
