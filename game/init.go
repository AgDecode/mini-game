package game

import (
	"github.com/AgDecode/mini-game/entity"
)

type State struct {
	Player       *entity.Player
	Rooms        map[string]*entity.Room
	DoorOpened   bool
	LastCommand  string
	EventEmitter *entity.EventEmitter
}

func InitGame() *State {
	state := &State{
		Rooms:        make(map[string]*entity.Room),
		EventEmitter: entity.NewEventEmitter(),
	}

	kitchen := entity.NewRoom("кухня", "ты находишься на кухне")
	kitchen.HasHint = true
	corridor := entity.NewRoom("коридор", "ничего интересного")
	room := entity.NewRoom("комната", "ты в своей комнате")
	street := entity.NewRoom("улица", "на улице весна")
	home := entity.NewRoom("домой", "ты дома")

	tea := entity.NewItem("чай", "")
	keys := entity.NewItem("ключи", "")
	notes := entity.NewItem("конспекты", "")
	backpack := entity.NewItem("рюкзак", "")

	backpack.SetTrait("wearable", true)
	keys.SetTrait("can_open", true)

	kitchen.AddItem(tea, "столе")
	room.AddItem(keys, "столе")
	room.AddItem(notes, "столе")
	room.AddItem(backpack, "стуле")

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

	registerEventHandlers(state)

	return state
}

func registerEventHandlers(state *State) {
	state.EventEmitter.On(entity.EventType("use"), func(event *entity.Event) error {
		item, ok := event.Source.(*entity.Item)
		if !ok {
			return nil
		}
		target, ok := event.Target.(*entity.Item)
		if !ok {
			return nil
		}

		if item.Name == "ключи" && target.Name == "дверь" {
			state.DoorOpened = true

			doorOpenedEvent := &entity.Event{
				Type:   entity.EventType("door_opened"),
				Source: item,
				Target: target,
				Data:   make(map[string]interface{}),
			}
			err := state.EventEmitter.Emit(doorOpenedEvent)
			if err != nil {
				return err
			}
		}
		return nil
	})

	state.EventEmitter.On(entity.EventType("enter_room"), func(event *entity.Event) error {
		room, ok := event.Target.(*entity.Room)
		if !ok {
			return nil
		}

		switch room.Name {
		case "кухня":
			if state.LastCommand == "идти" {
				room.Description = "ничего интересного"
			}
			if len(room.GetItemsByPlace("стол")) > 0 {
				room.HasHint = true
			}
		case "комната":
			if len(room.GetItems()) == 0 {
				room.Description = "пустая комната"
			}
		}
		return nil
	})
}
