package distrib

import (
	"chat/internal/model"
	"chat/internal/utils"
	"context"
	"encoding/json"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func getKey(uid int64, ct model.ChannelType) string {
	return strconv.FormatInt(uid, 10) + strconv.Itoa(int(ct))
}

func StoreSession(ctx context.Context, session *model.SessionDTO) error {
	parsed, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return GetRedisClient().Set(ctx, getKey(session.UserID, session.ChannelType), parsed, 0).Err()
}

func GetSession(ctx context.Context, uid int64, ct model.ChannelType) (*model.SessionDTO, error) {
	v, err := GetRedisClient().Get(ctx, getKey(uid, ct)).Result()
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

func RemoveSession(ctx context.Context, uid int64, ct model.ChannelType) error {
	return GetRedisClient().Del(ctx, getKey(uid, ct)).Err()
}
