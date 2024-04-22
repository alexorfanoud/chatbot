package distrib

import (
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func StoreSession(ctx context.Context, uid int, session *model.SessionDTO) error {
	parsed, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return GetRedisClient().Set(ctx, strconv.Itoa(uid), parsed, 0).Err()
}

func GetSession(ctx context.Context, uid int64) (*model.SessionDTO, error) {
	v, err := GetRedisClient().Get(ctx, strconv.FormatInt(uid, 10)).Result()
	if err == redis.Nil {
		return nil, nil
	}

	var s *model.SessionDTO
	err = json.Unmarshal([]byte(v), &s)
	if err != nil {
		utils.Log(ctx, "Unable to parse session from redis store")
		return nil, err
	}

	return s, nil
}

func RemoveSession(ctx context.Context, uid int) error {
	return GetRedisClient().Del(ctx, strconv.Itoa(uid)).Err()
}
