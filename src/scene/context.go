package scene

import (
	"context"
)

type SceneState struct {
	Scene string
	Stage int
	Data  interface{}
}

type sceneMap map[int64]SceneState

type Ctx struct {
	Ctx context.Context
}

var sceneKey = "SceneKey"

func (ctx *Ctx) Start(userId int64, scene string) {
	ctx.SetValue(userId, SceneState{
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

func (s Ctx) GetValue(userId int64) (SceneState, bool) {
	sMap, _ := s.Ctx.Value(sceneKey).(sceneMap)
	value, exist := sMap[userId]

	return value, exist
}

func (s *Ctx) SetValue(userId int64, state SceneState) {
	sMap, ok := s.Ctx.Value(sceneKey).(sceneMap)

	if ok {
		sMap[userId] = state
	} else {
		sMap = sceneMap{userId: state}
	}

	s.Ctx = context.WithValue(s.Ctx, sceneKey, sMap)
}

func NewSceneContext() Ctx {
	ctx := context.Background()
	defer ctx.Done()

	return Ctx{
		Ctx: ctx,
	}
}
