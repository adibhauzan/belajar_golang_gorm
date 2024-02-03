package belajar_golang_gorm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strconv"
	"testing"
)

func OpenConnection() *gorm.DB {
	dialect := mysql.Open("root:@tcp(localhost:3306)/belajar_golang_gorm?charset=utf8mb4&parseTime=true&loc=Local")
	db, err := gorm.Open(dialect, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	return db
}

var db = OpenConnection()

func TestConnection(t *testing.T) {
	assert.NotNil(t, db)
}

type Sample struct {
	Id   string
	Name string
}

func TestExecuteSQL(t *testing.T) {
	err := db.Exec("INSERT INTO users(id, name, password) values(?, ?, ?)", "2", "adib", "adib123").Error
	assert.Nil(t, err)
}

func TestExecuteDeleteSQL(t *testing.T) {
	err := db.Exec("DELETE FROM users WHERE id = ?", "3").Error
	assert.Nil(t, err)
}

func TestRawSQL(t *testing.T) {
	var sample Sample

	err := db.Raw("SELECT id, name from sample where id = ?", "1").Scan(&sample).Error
	assert.Nil(t, err)
	assert.Equal(t, "adib", sample.Name)

	var samples []Sample
	err = db.Raw("SELECT id, name from sample").Scan(&samples).Error
	assert.Nil(t, err)
	assert.Equal(t, 3, len(samples))
}

func TestSQlRow(t *testing.T) {
	var samples []Sample

	rows, err := db.Raw("SELECT id, name FROM sample").Rows()
	assert.Nil(t, err)
	defer rows.Close()

	for rows.Next() {
		var id string
		var name string

		err := rows.Scan(&id, &name)
		assert.Nil(t, err)

		samples = append(samples, Sample{
			Id:   id,
			Name: name,
		})
	}

	assert.Equal(t, 3, len(samples))
}
func TestScanRow(t *testing.T) {
	var samples []Sample

	rows, err := db.Raw("SELECT id, name FROM sample").Rows()
	assert.Nil(t, err)
	defer rows.Close()

	for rows.Next() {
		err := db.ScanRows(rows, &samples)
		assert.Nil(t, err)
	}

	assert.Equal(t, 3, len(samples))
}

func TestCreateUser(t *testing.T) {
	user := User{
		ID: "1",
		Name: Name{
			FirstName:  "Adib",
			MiddleName: "Hauzan",
			LastName:   "Sofyan",
		},
		Password:    "adib123",
		Information: "ini akan di ignore",
	}

	response := db.Create(&user)
	assert.Nil(t, response.Error)
	assert.Equal(t, int64(1), response.RowsAffected)
}

func TestBatchInsert(t *testing.T) {
	var users []User

	for i := 2; i <= 10; i++ {
		users = append(users, User{
			ID: strconv.Itoa(i),
			Name: Name{
				FirstName: "User " + strconv.Itoa(i),
			},
			Password: "Rahasia" + strconv.Itoa(i),
		})
	}

	result := db.Create(&users)
	assert.Nil(t, result.Error)
	assert.Equal(t, 9, int(result.RowsAffected))
}

func TestTransactionSuccess(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&User{
			ID: "11",
			Name: Name{
				FirstName: "adib lagi",
			},
			Password: "adib123",
		}).Error
		if err != nil {
			return err
		}

		err = tx.Create(&User{
			ID: "12",
			Name: Name{
				FirstName: "adib lagi dan lagi",
			},
			Password: "adib123",
		}).Error
		if err != nil {
			return err
		}

		err = tx.Create(&User{
			ID: "13",
			Name: Name{
				FirstName: "adib lagi dan lagi dan lagi",
			},
			Password: "adib123",
		}).Error
		if err != nil {
			return err
		}

		return nil
	})

	assert.Nil(t, err)
}

func TestTransactionFailed(t *testing.T) {
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&User{
			ID: "14",
			Name: Name{
				FirstName: "adib lagi",
			},
			Password: "adib123",
		}).Error
		if err != nil {
			return err
		}

		err = tx.Create(&User{
			ID: "12",
			Name: Name{
				FirstName: "adib lagi dan lagi",
			},
			Password: "adib123",
		}).Error
		if err != nil {
			return err
		}
		return nil
	})

	assert.Nil(t, err)
}

func TestManualTransactionSuccess(t *testing.T) {
	tx := db.Begin()
	defer tx.Rollback()

	err := tx.Create(&User{
		ID: "14",
		Name: Name{
			FirstName: "adib lagi dan lagi",
		},
		Password: "adib123",
	}).Error

	assert.Nil(t, err)

	err = tx.Create(&User{
		ID: "15",
		Name: Name{
			FirstName: "adib lagi dan lagi",
		},
		Password: "adib123",
	}).Error

	assert.Nil(t, err)

	if err == nil {
		tx.Commit()
	}
}

func TestQuerySingleObject(t *testing.T) {
	user := User{}
	result := db.First(&user)
	assert.Nil(t, result.Error)
	assert.Equal(t, "1", user.ID)

	user = User{}
	result = db.Last(&user)
	assert.Nil(t, result.Error)
	assert.Equal(t, "9", user.ID)
}

func TestQuerySingleObjectInlineCondition(t *testing.T) {
	user := User{}
	err := db.First(&user, "id = ?", "5").Error
	assert.Nil(t, err)
	assert.Equal(t, "5", user.ID)
	assert.Equal(t, "User 5", user.Name.FirstName)
}

func TestQueryCondition(t *testing.T) {
	var users []User
	err := db.Where("first_name like ?", "%User%").Where("password = ?", "Rahasia123").Find(&users).Error

	assert.Nil(t, err)
	assert.Equal(t, 0, len(users))
}

func TestQueryOrCondition(t *testing.T) {
	var users []User
	err := db.Where("first_name like ?", "%User%").Or("password = ?", "Rahasia123").Find(&users).Error

	assert.Nil(t, err)
	assert.Equal(t, 9, len(users))
}

func TestSelectField(t *testing.T) {
	var users []User
	err := db.Select("id", "first_name").Find(&users).Error
	assert.Nil(t, err)

	for _, user := range users {
		assert.NotNil(t, user.ID)
		assert.NotEqual(t, "", user.Name.FirstName)
	}

	assert.Equal(t, 15, len(users))
}

func TestStructCondition(t *testing.T) {
	userCondition := User{
		Name: Name{
			MiddleName: "Hauzan",
		},
	}

	var users []User
	err := db.Where(userCondition).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 1, len(users))
	assert.NotEqual(t, 5, len(users))
}

func TestLimitAndOffset(t *testing.T) {
	var users []User
	err := db.Order("id asc, first_name desc").Limit(5).Offset(5).Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 5, len(users))
	assert.NotEqual(t, 6, len(users))
}

type UserResponse struct {
	Id        string
	FirstName string
	LastName  string
}

func TestQueryNonModel(t *testing.T) {
	var users []UserResponse
	err := db.Model(&User{}).Select("id", "first_name", "last_name").Find(&users).Error
	assert.Nil(t, err)
	assert.Equal(t, 15, len(users))
	fmt.Println(users)
}
