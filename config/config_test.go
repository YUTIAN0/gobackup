package config

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func init() {
	Init("../gobackup_test.yml")
}

func TestModelsLength(t *testing.T) {
	assert.Equal(t, Exist, true)
	assert.Equal(t, len(Models), 4)
}

func TestModel(t *testing.T) {
	model := GetModelConfigByName("base_test")

	assert.Equal(t, model.Name, "base_test")

	// compress_with
	assert.Equal(t, model.CompressWith.Type, "tgz")
	assert.NotNil(t, model.CompressWith.Viper)

	// encrypt_with
	assert.Equal(t, model.EncryptWith.Type, "openssl")
	assert.NotNil(t, model.EncryptWith.Viper)

	assert.Equal(t, model.Storages["local"].Type, "local")
	assert.Equal(t, model.Storages["local"].Viper.GetString("path"), "/Users/jason/Downloads/backup1")

	assert.Equal(t, model.Storages["scp"].Type, "scp")
	assert.Equal(t, model.Storages["scp"].Viper.GetString("host"), "your-host.com")

	// databases
	assert.Len(t, model.Databases, 3)

	// mysql
	db := model.GetDatabaseByName("dummy_test")
	assert.Equal(t, db.Name, "dummy_test")
	assert.Equal(t, db.Type, "mysql")
	assert.Equal(t, db.Viper.GetString("host"), "localhost")
	assert.Equal(t, db.Viper.GetString("port"), "3306")
	assert.Equal(t, db.Viper.GetString("database"), "dummy_test")
	assert.Equal(t, db.Viper.GetString("username"), "root")
	assert.Equal(t, db.Viper.GetString("password"), "123456")

	// redis
	db = model.GetDatabaseByName("redis1")
	assert.Equal(t, db.Name, "redis1")
	assert.Equal(t, db.Type, "redis")
	assert.Equal(t, db.Viper.GetString("mode"), "sync")
	assert.Equal(t, db.Viper.GetString("rdb_path"), "/var/db/redis/dump.rdb")
	assert.Equal(t, db.Viper.GetBool("invoke_save"), true)
	assert.Equal(t, db.Viper.GetString("password"), "456123")

	// redis
	db = model.GetDatabaseByName("postgresql")
	assert.Equal(t, db.Name, "postgresql")
	assert.Equal(t, db.Type, "postgresql")
	assert.Equal(t, db.Viper.GetString("host"), "localhost")

	// archive
	includes := model.Archive.GetStringSlice("includes")
	assert.Len(t, includes, 4)
	assert.Contains(t, includes, "/home/ubuntu/.ssh/")
	assert.Contains(t, includes, "/etc/nginx/nginx.conf")

	excludes := model.Archive.GetStringSlice("excludes")
	assert.Len(t, excludes, 2)
	assert.Contains(t, excludes, "/home/ubuntu/.ssh/known_hosts")

	// schedule
	schedule := model.Schedule
	assert.Equal(t, true, schedule.Enabled)
	assert.Equal(t, "5 4 * * sun", schedule.Cron)
}

func Test_otherModels(t *testing.T) {
	model := GetModelConfigByName("normal_files")

	// schedule
	schedule := model.Schedule
	assert.Equal(t, true, schedule.Enabled)
	assert.Equal(t, "", schedule.Cron)
	assert.Equal(t, "1day", schedule.Every)
	assert.Equal(t, "0:30", schedule.At)

	model = GetModelConfigByName("test_model")
	assert.Equal(t, false, model.Schedule.Enabled)
}

func Test_ScheduleConfig_String(t *testing.T) {
	schedule := ScheduleConfig{
		Enabled: true,
		Every:   "1day",
		At:      "0:30",
	}
	assert.Equal(t, schedule.String(), "every 1day at 0:30")

	schedule = ScheduleConfig{
		Enabled: true,
		Every:   "1day",
	}
	assert.Equal(t, schedule.String(), "every 1day")

	schedule = ScheduleConfig{
		Enabled: true,
		Cron:    "5 4 * * sun",
	}

	assert.Equal(t, schedule.String(), "cron 5 4 * * sun")

	schedule = ScheduleConfig{
		Enabled: false,
	}
	assert.Equal(t, schedule.String(), "disabled")
}
