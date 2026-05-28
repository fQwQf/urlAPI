package database

import (
	"github.com/pkg/errors"
	"reflect"
	"urlAPI/internal/model"
)

/**
 * @brief 创建任务记录。
 * @param task 待写入的任务对象。
 * @return error 写入失败时返回错误。
 */
func (adapter *SQLiteAdapter) CreateTask(task *model.Task) error {
	return errors.WithStack(adapter.db.Create(task).Error)
}

/**
 * @brief 更新任务记录。
 * @param task 待更新的任务对象。
 * @return error 更新失败时返回错误。
 */
func (adapter *SQLiteAdapter) UpdateTask(task *model.Task) error {
	return errors.WithStack(adapter.db.Save(task).Error)
}

/**
 * @brief 查询任务记录。
 * @param task 查询条件，零值字段会被忽略。
 * @return *model.DBList 查询结果集合。
 * @return error 查询失败时返回错误。
 */
func (adapter *SQLiteAdapter) ReadTask(task model.Task) (*model.DBList, error) {
	var tasks []model.Task
	var err error
	query := adapter.db.Model(&model.Task{})
	if !task.Time.IsZero() {
		start := task.Time
		end := start.AddDate(0, 1, 0)
		query.Where("time >= ? AND time <= ?", start, end)
	} else {
		val := reflect.ValueOf(task)
		if val.Kind() == reflect.Struct {
			for i := 0; i < val.NumField(); i++ {
				field := val.Type().Field(i).Tag.Get("json")
				value := val.Field(i)
				if value.IsZero() {
					continue
				}
				if value.Interface().(string) == "N/A" {
					query.Where(field+"=? OR "+field+" IS NULL", "")
				} else {
					query.Where(field+"=?", value.Interface().(string))
				}
			}
		}
	}
	err = query.Find(&tasks).Error
	ret := model.DBList{
		TaskList: tasks,
	}
	return &ret, errors.WithStack(err)
}

/**
 * @brief 删除任务记录。
 * @param task 待删除的任务对象。
 * @return error 删除失败时返回错误。
 */
func (adapter *SQLiteAdapter) DeleteTask(task *model.Task) error {
	return errors.WithStack(adapter.db.Delete(task).Error)
}

/** @brief 创建任务记录的包级代理函数。 */
func CreateTask(task *model.Task) error { return localDB.CreateTask(task) }

/** @brief 更新任务记录的包级代理函数。 */
func UpdateTask(task *model.Task) error { return localDB.UpdateTask(task) }

/** @brief 查询任务记录的包级代理函数。 */
func ReadTask(task model.Task) (*model.DBList, error) { return localDB.ReadTask(task) }

/** @brief 删除任务记录的包级代理函数。 */
func DeleteTask(task *model.Task) error { return localDB.DeleteTask(task) }
