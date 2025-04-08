package game

import (
	"github.com/AgDecode/mini-game/entity"
)

func InitGame() *State {
	state := NewState()
	registerCommands(state)
	registerInteractionRules(state)
	state.RegisterEventHandlers()

	kitchen := entity.NewRoom("кухня", "ты находишься на кухне", "кухня, ничего интересного. можно пройти - коридор")
	corridor := entity.NewRoom("коридор", "ничего интересного", "ничего интересного. можно пройти - кухня, комната, улица")
	room := entity.NewRoom("комната", "ты в своей комнате", "ты в своей комнате. можно пройти - коридор")
	street := entity.NewRoom("улица", "на улице весна", "на улице весна. можно пройти - домой")
	home := entity.NewRoom("домой", "ты дома", "ты дома. можно пройти - улица")
	kitchen.HasHint = true

	tea := entity.NewItem("чай", "")
	keys := entity.NewItem("ключи", "")
	notes := entity.NewItem("конспекты", "")
	backpack := entity.NewItem("рюкзак", "")
	door := entity.NewItem("дверь", "закрытая дверь на улицу")

	backpack.SetTrait("wearable", true)
	keys.SetTrait("can_open", true)
	door.SetTrait("openable", true)
	door.SetTrait("is_open", false)

	kitchen.AddItem(tea, "столе")
	room.AddItem(keys, "столе")
	room.AddItem(notes, "столе")
	room.AddItem(backpack, "стуле")
	corridor.AddItem(door, "стене")

	kitchen.Connections["коридор"] = corridor
	corridor.Connections["кухня"] = kitchen
	corridor.Connections["комната"] = room
	corridor.Connections["улица"] = street
	room.Connections["коридор"] = corridor
	street.Connections["домой"] = home
	home.Connections["улица"] = street

	state.Rooms["кухня"] = kitchen
	state.Rooms["коридор"] = corridor
	state.Rooms["комната"] = room
	state.Rooms["улица"] = street
	state.Rooms["домой"] = home

	state.Player = entity.NewPlayer(kitchen)

	return state
}

func registerCommands(state *State) {
	state.RegisterCommand("осмотреться", func(s *State, args []string) string {
		return s.handleLook()
	})

	state.RegisterCommand("идти", func(s *State, args []string) string {
		if len(args) == 0 {
			return MsgNoDirection
		}
		return s.handleGo(args[0])
	})

	state.RegisterCommand("взять", func(s *State, args []string) string {
		if len(args) == 0 {
			return MsgNoItem
		}
		return s.handleTake(args[0])
	})

	state.RegisterCommand("надеть", func(s *State, args []string) string {
		if len(args) == 0 {
			return MsgNoItem
		}
		return s.handleWear(args[0])
	})

	state.RegisterCommand("применить", func(s *State, args []string) string {
		if len(args) < 2 {
			return MsgNoItems
		}
		return s.handleUse(args[0], args[1])
	})
}

func registerInteractionRules(state *State) {
	state.RegisterInteractionRule(InteractionRule{
		SourceTraits: map[string]interface{}{
			"can_open": true,
		},
		TargetTraits: map[string]interface{}{
			"openable": true,
			"is_open":  false,
		},
		StateModifier: func(s *State, source, target *entity.Item) {
			s.DoorOpened = true
			target.SetTrait("is_open", true)
		},
		EventEmitter: func(s *State, source, target *entity.Item) {
			doorOpenedEvent := &entity.Event{
				Type:   "door_opened",
				Source: source,
				Target: target,
				Data:   make(map[string]interface{}),
			}
			err := s.EventEmitter.Emit(doorOpenedEvent)
			if err != nil {
				return
			}
		},
		ResultHandler: func(s *State, source, target *entity.Item) string {
			return "дверь открыта"
		},
	})
}
