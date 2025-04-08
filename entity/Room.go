package entity

import (
	"fmt"
	"strings"
)

type RoomBehavior struct {
	OnEnter func(*Player) string
	OnLook  func(*Player) string
}

type Room struct {
	Name        string
	Description string
	Items       map[string][]*Item
	Connections map[string]*Room
	Traits      map[string]interface{}
	Emitter     *EventEmitter
	WasVisited  bool
	HasHint     bool
}

func NewRoom(name, description string) *Room {
	return &Room{
		Name:        name,
		Description: description,
		Items:       make(map[string][]*Item),
		Connections: make(map[string]*Room),
		Traits:      make(map[string]interface{}),
		Emitter:     NewEventEmitter(),
	}
}

func (r *Room) Look() string {
	return r.GetDescription()
}

func (r *Room) GetDescription() string {
	var parts []string

	if r.Description != "" {
		parts = append(parts, r.Description)
	}

	for place, items := range r.Items {
		if len(items) > 0 {
			itemNames := make([]string, len(items))
			for i, item := range items {
				itemNames[i] = item.Name
			}
			parts = append(parts, fmt.Sprintf("на %s: %s", place, strings.Join(itemNames, ", ")))
		}
	}

	var exits []string
	for dir := range r.Connections {
		exits = append(exits, dir)
	}
	if len(exits) > 0 {
		parts = append(parts, fmt.Sprintf("можно пройти - %s", strings.Join(exits, ", ")))
	}

	return strings.Join(parts, ", ")
}

func (r *Room) FormatDescription(parts []string) string {
	return strings.Join(parts, ". ")
}

func (r *Room) HasTrait(trait string) bool {
	_, exists := r.Traits[trait]
	return exists
}

func (r *Room) GetTrait(trait string) interface{} {
	return r.Traits[trait]
}

func (r *Room) SetTrait(trait string, value interface{}) {
	r.Traits[trait] = value
}

func (r *Room) OnEnter(player *Player) string {
	event := &Event{
		Type:   EventRoomEntered,
		Source: player,
		Target: r,
		Data:   make(map[string]interface{}),
	}
	err := r.Emitter.Emit(event)
	if err != nil {
		return ""
	}
	if event.Prevented {
		return event.Data["message"].(string)
	}
	return r.GetDescription()
}

func (r *Room) OnExit(player *Player) string {
	event := &Event{
		Type:   EventRoomExited,
		Source: player,
		Target: r,
		Data:   make(map[string]interface{}),
	}
	err := r.Emitter.Emit(event)
	if err != nil {
		return ""
	}
	return ""
}

func (r *Room) AddItem(item *Item, place string) {
	if r.Items == nil {
		r.Items = make(map[string][]*Item)
	}
	r.Items[place] = append(r.Items[place], item)
	event := &Event{
		Type:   EventItemDropped,
		Source: item,
		Target: r,
		Data:   make(map[string]interface{}),
	}
	err := r.Emitter.Emit(event)
	if err != nil {
		return
	}
}

func (r *Room) RemoveItem(item *Item, place string) {
	if items, ok := r.Items[place]; ok {
		for i, it := range items {
			if it == item {
				r.Items[place] = append(items[:i], items[i+1:]...)
				event := &Event{
					Type:   EventItemPicked,
					Source: item,
					Target: r,
					Data:   make(map[string]interface{}),
				}
				err := r.Emitter.Emit(event)
				if err != nil {
					return
				}
				return
			}
		}
	}
}

func (r *Room) GetItems() []*Item {
	var allItems []*Item
	for _, items := range r.Items {
		allItems = append(allItems, items...)
	}
	return allItems
}

func (r *Room) GetItemsByPlace(place string) []*Item {
	return r.Items[place]
}

func (r *Room) HasItem(itemName string) bool {
	for _, items := range r.Items {
		for _, item := range items {
			if item.Name == itemName {
				return true
			}
		}
	}
	return false
}

func (r *Room) GetItem(itemName string) *Item {
	for _, items := range r.Items {
		for _, item := range items {
			if item.Name == itemName {
				return item
			}
		}
	}
	return nil
}

func (r *Room) AddConnection(direction string, room *Room) {
	r.Connections[direction] = room
}

func (r *Room) CanGo(direction string) bool {
	_, exists := r.Connections[direction]
	return exists
}

func (r *Room) GetNeighbors() string {
	var neighbors []string
	for name := range r.Connections {
		neighbors = append(neighbors, name)
	}
	return strings.Join(neighbors, ", ")
}

func (r *Room) GetItemsOnFurniture(place string) []string {
	var items []string
	for _, item := range r.Items[place] {
		items = append(items, item.Name)
	}
	return items
}
