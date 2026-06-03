package database

import (
	"database/sql"
	"encoding/json"
	"github.com/common-nighthawk/go-figure"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
)

/**
 * @brief 初始化数据库连接及内存缓存。
 * @return error 初始化任一步骤失败时返回错误。
 */
func Init() error {
	figlet := figure.NewFigure("urlAPI", "", true)
	figlet.Print()
	if err := connect(); err != nil {
		return err
	}
	migration()
	if err := initRepoMap(); err != nil {
		return err
	}
	if err := initSessionMap(); err != nil {
		return err
	}
	if err := initAppSettings(); err != nil {
		return errors.Wrap(err, "initAppSettings")
	}
	return nil
}

/**
 * @brief 执行数据库表结构迁移。
 */
func migration() {
	localDB.db.AutoMigrate(&AppSetting{})
	localDB.db.AutoMigrate(&Provider{})
	localDB.db.AutoMigrate(&ServiceConfig{})
	localDB.db.AutoMigrate(&Prompt{})
	localDB.db.AutoMigrate(&ConfigListItem{})
	localDB.db.AutoMigrate(&Task{})
	localDB.db.AutoMigrate(&Session{})
	localDB.db.AutoMigrate(&Repo{})
	localDB.db.AutoMigrate(&APIKey{})
	localDB.db.AutoMigrate(&APIKeyUsage{})
}

/**
 * @brief 建立 SQLite 数据库连接。
 * @return error 连接失败时返回错误。
 */
func connect() error {
	var err error
	os.Mkdir("assets", 0777)
	tmp, _ := sql.Open("sqlite3", dbPath)
	tmp.Close()
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return errors.Wrap(err, "gorm")
	}
	SetLocalDB(db)
	log.Println("Connected to database")
	return nil
}

/**
 * @brief 关闭数据库连接。
 */
func Disconnect() {
	sqlDB, _ := localDB.db.DB()
	defer sqlDB.Close()
	log.Println("Disconnected from database")
}

/**
 * @brief 从数据库初始化仓库缓存映射。
 * @return error 读取或反序列化失败时返回错误。
 */
func initRepoMap() error {
	var repos []Repo
	if err := localDB.db.Find(&repos).Error; err != nil {
		return errors.Wrap(err, "db find")
	}
	for _, repo := range repos {
		var repoList []string
		if err := json.Unmarshal([]byte(repo.Content), &repoList); err != nil {
			return errors.Wrap(err, "json")
		}
		RepoMap[repo.API+";"+repo.Info] = repoList
	}
	log.Println("Initialized RepoMap")
	return nil
}

/**
 * @brief 从数据库初始化会话缓存映射。
 * @return error 读取失败时返回错误。
 */
func initSessionMap() error {
	var sessions []Session
	if err := localDB.db.Find(&sessions).Error; err != nil {
		return errors.Wrap(err, "db")
	}
	for _, session := range sessions {
		SessionMap[session.Token] = session
	}
	log.Println("Initialized SessionMap")
	return nil
}

/**
 * @brief 清空任务表并重新创建结构。
 */
func ClearTask() {
	if localDB.db.Migrator().HasTable(&Task{}) {
		if err := localDB.db.Migrator().DropTable(&Task{}); err != nil {
			log.Fatal(errors.Wrap(err, "db"))
		}
		if err := localDB.db.AutoMigrate(&Task{}); err != nil {
			log.Fatal(errors.Wrap(err, "db"))
		}
	}
}

/**
 * @brief 清空会话表并重置内存会话缓存。
 */
func ClearSession() {
	if localDB.db.Migrator().HasTable(&Session{}) {
		if err := localDB.db.Migrator().DropTable(&Session{}); err != nil {
			log.Fatal(errors.Wrap(err, "db"))
		}
		if err := localDB.db.AutoMigrate(&Session{}); err != nil {
			log.Fatal(errors.Wrap(err, "db"))
		}
	}
	SessionMap = make(map[string]Session)
}
