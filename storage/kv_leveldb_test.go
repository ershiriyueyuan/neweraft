package storage

import (
	"testing"
	"os"
)

// TestLevelDBKvStore_OpenClose 测试数据库的打开和关闭功能
func TestLevelDBKvStore_OpenClose(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例
	db := MakeLevelDBKvStore()
	db.Path = tempDir // 使用临时目录

	// 测试Open方法
	err = db.Open()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// 测试Close方法
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}
}

// TestLevelDBKvStore_PutGet 测试Put和Get方法的基本功能
func TestLevelDBKvStore_PutGet(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例并打开
	db := MakeLevelDBKvStore()
	db.Path = tempDir
	if err := db.Open(); err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 测试数据
	key := "test_key"
	value := "test_value"

	// 测试Put方法
	err = db.Put(key, value)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// 测试Get方法
	gotValue, err := db.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// 验证值是否正确
	if gotValue != value {
		t.Errorf("Get returned wrong value: got %q want %q", gotValue, value)
	}
}

// TestLevelDBKvStore_Del 测试Del方法功能
func TestLevelDBKvStore_Del(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例并打开
	db := MakeLevelDBKvStore()
	db.Path = tempDir
	if err := db.Open(); err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 测试数据
	key := "delete_key"
	value := "delete_value"

	// 先存储数据
	db.Put(key, value)

	// 测试Del方法
	err = db.Del(key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证数据是否已删除
	_, err = db.Get(key)
	if err == nil {
		t.Errorf("Get should return error after Delete, but got nil")
	}
}

// TestLevelDBKvStore_NotOpened 测试未打开数据库时的错误处理
func TestLevelDBKvStore_NotOpened(t *testing.T) {
	// 创建数据库实例但不打开
	db := MakeLevelDBKvStore()

	// 测试未打开时调用Put
	err := db.Put("key", "value")
	if err == nil {
		t.Error("Put should return error when database is not opened")
	}

	// 测试未打开时调用Get
	_, err = db.Get("key")
	if err == nil {
		t.Error("Get should return error when database is not opened")
	}

	// 测试未打开时调用Del
	err = db.Del("key")
	if err == nil {
		t.Error("Del should return error when database is not opened")
	}
}

// TestLevelDBKvStore_GetNonExistent 测试获取不存在的键
func TestLevelDBKvStore_GetNonExistent(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例并打开
	db := MakeLevelDBKvStore()
	db.Path = tempDir
	if err := db.Open(); err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 测试获取不存在的键
	_, err = db.Get("non_existent_key")
	if err == nil {
		t.Error("Get should return error for non-existent key")
	}
}

// TestEngineerFactory 测试工厂方法功能
func TestEngineerFactory(t *testing.T) {
	// 测试创建leveldb存储
	kvStore := Engineerfactory("leveldb")
	if kvStore == nil {
		t.Fatal("Engineerfactory returned nil for 'leveldb'")
	}

	// 验证返回的实例类型
	ldb, ok := kvStore.(*LevelDBKvStore)
	if !ok {
		t.Errorf("Engineerfactory returned wrong type for 'leveldb': %v", kvStore)
	}

	// 测试创建无效存储类型
	invalidStore := Engineerfactory("invalid_type")
	if invalidStore != nil {
		t.Errorf("Engineerfactory should return nil for invalid type, but got %v", invalidStore)
	}

	// 清理资源
	if ldb != nil {
		// 如果已经打开，需要关闭
		if ldb.db != nil {
			ldb.Close()
		}
	}
}