package scene

import (
	"bot/src/bot"
	"context"
)

type sceneMap map[int64]bot.SceneState

type Ctx struct {
	Ctx context.Context
}

var sceneKey = "SceneKey"

func (ctx *Ctx) Start(userId int64, scene string) {
	ctx.SetCtxValue(userId, bot.SceneState{
		Scene: scene,
		Stage: 1,
	})
}

func (ctx *Ctx) End(userId int64) {
	sMap := ctx.Ctx.Value(sceneKey).(sceneMap)

	delete(sMap, userId)

	ctx.Ctx = context.WithValue(ctx.Ctx, sceneKey, sMap)
}

func (s *Ctx) Next(userId int64) {
	sMap := s.Ctx.Value(sceneKey).(sceneMap)
	state := sMap[userId]

	state.Stage++
	sMap[userId] = state

	s.Ctx = context.WithValue(s.Ctx, sceneKey, sMap)

}

func (s Ctx) GetCtxValue(userId int64) (bot.SceneState, bool) {
	sMap, _ := s.Ctx.Value(sceneKey).(sceneMap)
	value, exist := sMap[userId]

	return value, exist
}

func (s *Ctx) SetCtxValue(userId int64, state bot.SceneState) {
	sMap, ok := s.Ctx.Value(sceneKey).(sceneMap)

	if ok {
		sMap[userId] = state
	} else {
		sMap = sceneMap{userId: state}
	}

	s.Ctx = context.WithValue(s.Ctx, sceneKey, sMap)
}
