package main

import (
	"fmt"

	redis "github.com/go-redis/redis/v8"
)

type User struct {
	UserId     string `json:"userId"`
	Incorrect  int64  `redis:"incorrect" json:"incorrect"`
	Correct    int64  `redis:"correct" json:"correct"`
	Streak     int64  `redis:"streak" json:"streak"`
	Points     int64  `redis:"points" json:"points"`
	StreakLoss int64  `redis:"streakLoss" json:"streakLoss"`
}

type LbMeta struct {
	Page     int64  `json:"page"`
	PageSize int64  `json:"pageSize"`
	Count    int64  `json:"count"`
	Type     string `json:"type"`
}

type LbResponse struct {
	Meta LbMeta `json:"meta"`
	Data []User `json:"data"`
}

func getMainLeaderboard(orgId string, eventId string, page int64, pageSize int64) ([]User, int64) {

	lbKey := fmt.Sprintf("lb:%s:%s", orgId, eventId)
	from := page * pageSize
	to := page*pageSize + pageSize - 1

	pipe := rdb.Pipeline()
	pipe.ZRevRange(ctx, fmt.Sprintf("{polls}%s", lbKey), from, to)
	pipe.ZCard(ctx, fmt.Sprintf("{polls}%s", lbKey))
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}

	usersIds := cmds[0].(*redis.StringSliceCmd).Val()
	count := cmds[1].(*redis.IntCmd).Val()

	pipe = rdb.Pipeline()
	for _, userId := range usersIds {
		userKey := fmt.Sprintf("{polls}%s:%s", lbKey, userId)
		pipe.HMGet(ctx, userKey, "incorrect", "correct", "points", "streak", "streakLoss")
	}

	cmds, err = pipe.Exec(ctx)
	if err != nil {
		panic(err)
	}

	users := []User{}
	for index, cmd := range cmds {
		var model User
		cmd.(*redis.SliceCmd).Scan(&model)
		model.UserId = usersIds[index]
		users = append(users, model)
	}

	return users, count
}
