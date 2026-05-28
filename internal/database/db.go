package database

import "gorm.io/gorm"

/**
 * @brief SQLite 数据库适配器。
 */
type SQLiteAdapter struct {
	db *gorm.DB
}

var localDB = &SQLiteAdapter{}

/**
 * @brief 设置全局数据库连接。
 * @param db GORM 数据库实例。
 * @return *SQLiteAdapter 更新后的数据库适配器。
 */
func SetLocalDB(db *gorm.DB) *SQLiteAdapter {
	localDB.db = db
	return localDB
}

/**
 * @brief 获取全局数据库适配器。
 * @return *SQLiteAdapter 当前全局数据库适配器。
 */
func GetLocalDB() *SQLiteAdapter {
	return localDB
}

/**
 * @brief 获取底层 GORM 数据库实例。
 * @return *gorm.DB 数据库连接对象。
 */
func (adapter *SQLiteAdapter) DB() *gorm.DB {
	return adapter.db
}
