package game

import (
	"github.com/AgDecode/mini-game/entity"
)

func (s *State) RegisterEventHandlers() {
	s.EventEmitter.On("use", s.handleUseEvent)

	s.EventEmitter.On("enter_room", s.handleEnterRoomEvent)
}

func (s *State) handleUseEvent(event *entity.Event) error {
	item, ok := event.Source.(*entity.Item)
	if !ok {
		return nil
	}

	target, ok := event.Target.(*entity.Item)
	if !ok {
		return nil
	}

	canOpen, _ := item.GetTrait("can_open").(bool)
	isOpenable, _ := target.GetTrait("openable").(bool)

	if canOpen && isOpenable {
		s.DoorOpened = true
		target.SetTrait("is_open", true)

		doorOpenedEvent := &entity.Event{
			Type:   "door_opened",
			Source: item,
			Target: target,
			Data:   make(map[string]interface{}),
		}

		return s.EventEmitter.Emit(doorOpenedEvent)
	}

	return nil
}

func (s *State) handleEnterRoomEvent(event *entity.Event) error {
	room, ok := event.Target.(*entity.Room)
	if !ok {
		return nil
	}

	switch room.Name {
	case "кухня":
		if s.LastCommand == "идти" {
			room.Description = "ничего интересного"
		}
		if len(room.GetItemsByPlace("столе")) > 0 {
			room.HasHint = true
		}
	case "комната":
		if len(room.GetItems()) == 0 {
			room.Description = "пустая комната"
		}
	}

	return nil
}
