package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/RedHatInsights/sources-api-go/redis"
	"github.com/go-gormigrate/gormigrate/v2"
	uuidpkg "github.com/google/uuid"
	"gorm.io/gorm"
)

var migrationsCollection = []*gormigrate.Migration{
	InitialSchema(),
	AddOrgIdToTenants(),
	TranslateEbsAccountNumbersToOrgIds(),
	SourceTypesAddCategoryColumn(),
	AddRetryCounterToApplications(),
	AddTenantsExternalTenantOrgIdUnique(),
}

var ctx = context.Background()

// redisLockKey is the key which will be used for the Redis lock when performing the migrations.
const redisLockKey = "sources-api-go-redis-lock"

// redisSleepTime is the amount of time that the client will wait until the next retry to obtain the lock.
const redisSleepTime = 3 * time.Second

// redisLockExpirationTime is the time in milliseconds that the lock will be held before it expires.
const redisLockExpirationTime = 30 * time.Second

// Migrate migrates the database schema to the latest version. Implements the single instance lock algorithm detailed
// in https://redis.io/topics/distlock#correct-implementation-with-a-single-instance.
func Migrate(db *gorm.DB) error {
	// Using UUID as the lock value since it's a safe way of obtaining a unique string among all the clients.
	uuid, err := uuidpkg.NewUUID()
	if err != nil {
		return err
	}

	// Before doing anything, check for the existence of the lock.
	exists, err := redis.Client.Exists(ctx, redisLockKey).Result()
	if err != nil {
		return err
	}

	// If the lock is present, we must wait in order to be able to obtain it ourselves.
	lockExists := exists != 0
	for lockExists {
		time.Sleep(redisSleepTime)

		exists, err = redis.Client.Exists(ctx, redisLockKey).Result()
		if err != nil {
			return err
		}

		lockExists = exists != 0
	}

	// Set the migrations lock.
	err = redis.Client.Set(ctx, redisLockKey, uuid.String(), redisLockExpirationTime).Err()
	if err != nil {
		return err
	}

	// Perform the migrations and store the error for a proper return.
	migrateTool := gormigrate.New(db, gormigrate.DefaultOptions, migrationsCollection)
	migrationErr := migrateTool.Migrate()

	// Once the migrations have finished, get the lock's value to attempt to release it.
	value, err := redis.Client.Get(ctx, redisLockKey).Result()
	if err != nil {
		return err
	}

	// The lock's value should coincide with the one we set above. If it doesn't something very wrong happened.
	if value == uuid.String() {
		err = redis.Client.Del(ctx, redisLockKey).Err()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf(`migrations lock release failed. Expecting lock with value "%s", got "%s"`, uuid.String(), value)
	}

	// Return the result of the migration process, in case something went wrong.
	return migrationErr
}
