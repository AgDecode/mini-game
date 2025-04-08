package entity

import (
	"fmt"
)

type Item struct {
	Name        string
	Description string
	Traits      map[string]interface{}
	Emitter     *EventEmitter
}

func NewItem(name, description string) *Item {
	return &Item{
		Name:        name,
		Description: description,
		Traits:      make(map[string]interface{}),
		Emitter:     NewEventEmitter(),
	}
}

func (i *Item) GetDescription() string {
	return i.Description
}

func (i *Item) HasTrait(trait string) bool {
	_, exists := i.Traits[trait]
	return exists
}

func (i *Item) GetTrait(trait string) interface{} {
	return i.Traits[trait]
}

func (i *Item) SetTrait(trait string, value interface{}) {
	i.Traits[trait] = value
}

func (i *Item) Use(target interface{}) string {
	event := &Event{
		Type:   EventItemUsed,
		Source: i,
		Target: target,
		Data:   make(map[string]interface{}),
	}
	err := i.Emitter.Emit(event)
	if err != nil {
		return ""
	}
	return i.Description
}

func (i *Item) OnPickup(player *Player) string {
	event := &Event{
		Type:   EventItemPicked,
		Source: i,
		Target: player,
		Data:   make(map[string]interface{}),
	}
	err := i.Emitter.Emit(event)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("предмет добавлен в инвентарь: %s", i.Name)
}

func (i *Item) OnDrop(room *Room) string {
	event := &Event{
		Type:   EventItemDropped,
		Source: i,
		Target: room,
		Data:   make(map[string]interface{}),
	}
	err := i.Emitter.Emit(event)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("предмет выброшен: %s", i.Name)
}

func (i *Item) IsWearable() bool {
	return i.HasTrait("wearable")
}
