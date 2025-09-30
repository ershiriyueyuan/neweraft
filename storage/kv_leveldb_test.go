package storage

import (
	"testing"
	"os"
)

// TestLevelDBKvStore_PutGet 测试Put和Get方法的基本功能
func TestLevelDBKvStore_PutGet(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例
	db,err := MakeLevelDBKvStore(tempDir)
	if db.db == nil {
		t.Fatalf("创建LevelDB实例失败")
	}
	defer db.Close()

	// 测试数据
	key := "test_key"
	value := "test_value"

	// 测试Put方法
	err = db.Put(key, value)
	if err != nil {
		t.Fatalf("Put操作失败: %v", err)
	}

	// 测试Get方法
	gotValue, err := db.Get(key)
	if err != nil {
		t.Fatalf("Get操作失败: %v", err)
	}

	// 验证值是否正确
	if gotValue != value {
		t.Errorf("Get返回错误的值: 得到 %q 期望 %q", gotValue, value)
	}
}

// TestLevelDBKvStore_Del 测试Del方法功能
func TestLevelDBKvStore_Del(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例
	db,err := MakeLevelDBKvStore(tempDir)
	if db.db == nil {
		t.Fatalf("创建LevelDB实例失败")
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
		t.Fatalf("Delete操作失败: %v", err)
	}

	// 验证数据是否已删除
	_, err = db.Get(key)
	if err == nil {
		t.Errorf("删除后Get操作应该返回错误，但没有返回错误")
	}
}

// TestLevelDBKvStore_NotOpened 测试未打开数据库时的错误处理
func TestLevelDBKvStore_NotOpened(t *testing.T) {
	// 创建一个未初始化的数据库实例
	db := &LevelDBKvStore{}

	// 测试未打开时调用Put
	err := db.Put("key", "value")
	if err == nil {
		t.Error("数据库未打开时，Put操作应该返回错误")
	}

	// 测试未打开时调用Get
	_, err = db.Get("key")
	if err == nil {
		t.Error("数据库未打开时，Get操作应该返回错误")
	}

	// 测试未打开时调用Del
	err = db.Del("key")
	if err == nil {
		t.Error("数据库未打开时，Del操作应该返回错误")
	}
}

// TestLevelDBKvStore_GetNonExistent 测试获取不存在的键
func TestLevelDBKvStore_GetNonExistent(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例
	db,err := MakeLevelDBKvStore(tempDir)
	if db.db == nil {
		t.Fatalf("创建LevelDB实例失败")
	}
	defer db.Close()

	// 测试获取不存在的键
	_, err = db.Get("non_existent_key")
	if err == nil {
		t.Error("获取不存在的键时应该返回错误")
	}
}

// TestEngineerFactory 测试工厂方法功能
func TestEngineerFactory(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 测试创建leveldb存储
	kvStore := Engineerfactory("leveldb", tempDir)
	if kvStore == nil {
		t.Fatal("Engineerfactory创建'leveldb'存储失败，返回nil")
	}

	// 验证返回的实例类型
	ldb, ok := kvStore.(*LevelDBKvStore)
	if !ok {
		t.Errorf("Engineerfactory返回了错误的类型: %T", kvStore)
	}

	// 测试创建无效存储类型
	invalidStore := Engineerfactory("invalid_type", "")
	if invalidStore != nil {
		t.Errorf("Engineerfactory为无效类型应该返回nil，但返回了 %v", invalidStore)
	}

	// 清理资源
	if ldb != nil {
		ldb.Close()
	}
}

// TestLevelDBKvStore_MultipleOperations 测试多个连续操作
func TestLevelDBKvStore_MultipleOperations(t *testing.T) {
	// 创建一个临时目录用于测试
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试结束后清理

	// 创建数据库实例
	db,err := MakeLevelDBKvStore(tempDir)
	if db.db == nil {
		t.Fatalf("创建LevelDB实例失败")
	}
	defer db.Close()

	// 测试多个键值对
	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// 存储所有数据
	for k, v := range data {
		err = db.Put(k, v)
		if err != nil {
			t.Fatalf("Put操作失败 (key=%s): %v", k, err)
		}
	}

	// 验证所有数据
	for k, expected := range data {
		got, err := db.Get(k)
		if err != nil {
			t.Fatalf("Get操作失败 (key=%s): %v", k, err)
		}
		if got != expected {
			t.Errorf("Get返回错误的值 (key=%s): 得到 %q 期望 %q", k, got, expected)
		}
	}

	// 删除部分数据
	db.Del("key2")

	// 验证删除后的数据
	_, err = db.Get("key2")
	if err == nil {
		t.Error("删除key2后，Get操作应该返回错误")
	}

	// 验证未删除的数据
	got1, err := db.Get("key1")
	if err != nil || got1 != "value1" {
		t.Errorf("key1应该保持不变，但得到 %q (err: %v)", got1, err)
	}
}