package game

import (
	"fmt"
	"github.com/AgDecode/mini-game/entity"
	"strings"
)

const (
	MsgItemNotFound      = "нет такого"
	MsgNoBackpack        = "некуда класть"
	MsgNoDirection       = "не указано направление"
	MsgNoItem            = "не указан предмет"
	MsgNoItems           = "не указаны предметы"
	MsgUnknownCommand    = "неизвестная команда"
	MsgItemAdded         = "предмет добавлен в инвентарь: %s"
	MsgWearing           = "вы надели: %s"
	MsgCannotWear        = "нельзя надеть"
	MsgDoorClosed        = "дверь закрыта"
	MsgNoPath            = "нет пути в %s"
	MsgNoItemInInventory = "нет предмета в инвентаре - %s"
	MsgNothingToApply    = "не к чему применить"
)

func (s *State) HandleCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return MsgUnknownCommand
	}

	cmd := parts[0]
	args := parts[1:]

	if handler, exists := s.Commands[cmd]; exists {
		return handler(s, args)
	}

	return MsgUnknownCommand
}

func (s *State) handleLook() string {
	room := s.Player.CurrentRoom
	parts := []string{}

	if room.Name == "кухня" {
		parts = append(parts, "ты находишься на кухне")
	}

	itemParts := s.getItemDescriptionParts(room)
	if len(itemParts) > 0 {
		parts = append(parts, strings.Join(itemParts, ", "))
	} else if room.Name != "кухня" {
		parts = append(parts, "пустая комната")
	}

	s.addHintsToParts(room, &parts)

	s.addExitsToParts(room, &parts)

	return s.formatRoomDescription(room, parts)
}

func (s *State) getItemDescriptionParts(room *entity.Room) []string {
	var itemParts []string

	if items, ok := room.Items["столе"]; ok && len(items) > 0 {
		itemNames := s.getItemNames(items)
		itemParts = append(itemParts, fmt.Sprintf("на столе: %s", strings.Join(itemNames, ", ")))
	}

	if items, ok := room.Items["стуле"]; ok && len(items) > 0 {
		itemNames := s.getItemNames(items)
		itemParts = append(itemParts, fmt.Sprintf("на стуле: %s", strings.Join(itemNames, ", ")))
	}

	return itemParts
}

func (s *State) getItemNames(items []*entity.Item) []string {
	itemNames := make([]string, len(items))
	for i, item := range items {
		itemNames[i] = item.Name
	}
	return itemNames
}

func (s *State) addHintsToParts(room *entity.Room, parts *[]string) {
	if room.HasHint {
		if room.Name == "кухня" {
			if s.Player.HasBackpack() {
				*parts = append(*parts, "надо идти в универ")
			} else {
				*parts = append(*parts, "надо собрать рюкзак и идти в универ")
			}
		}
	}
}

func (s *State) addExitsToParts(room *entity.Room, parts *[]string) {
	if len(room.Connections) > 0 {
		directions := make([]string, 0, len(room.Connections))
		for dir := range room.Connections {
			directions = append(directions, dir)
		}
		*parts = append(*parts, fmt.Sprintf("можно пройти - %s", strings.Join(directions, ", ")))
	}
}

func (s *State) formatRoomDescription(room *entity.Room, parts []string) string {
	if room.Name == "кухня" {
		return strings.Join(parts[:len(parts)-1], ", ") + ". " + parts[len(parts)-1]
	}
	return strings.Join(parts, ". ")
}

func (s *State) handleGo(direction string) string {
	room := s.Player.CurrentRoom
	nextRoom, exists := room.Connections[direction]

	if !exists {
		return fmt.Sprintf(MsgNoPath, direction)
	}

	if nextRoom.Name == "улица" && !s.DoorOpened {
		return MsgDoorClosed
	}

	s.Player.CurrentRoom = nextRoom
	s.LastCommand = "идти"

	s.emitEnterRoomEvent(nextRoom)

	return s.getRoomEnterMessage(nextRoom)
}

func (s *State) emitEnterRoomEvent(room *entity.Room) {
	event := &entity.Event{
		Type:   "enter_room",
		Source: s.Player,
		Target: room,
		Data:   make(map[string]interface{}),
	}
	err := s.EventEmitter.Emit(event)
	if err != nil {
		return
	}
}

func (s *State) getRoomEnterMessage(room *entity.Room) string {
	if room.EnterMessage != "" {
		return room.EnterMessage
	}
	return s.handleLook()
}

func (s *State) handleTake(itemName string) string {
	room := s.Player.CurrentRoom

	item, place := s.findItemWithLocation(room, itemName)
	if item == nil {
		return MsgItemNotFound
	}

	if !s.Player.HasBackpack() && item.Name != "рюкзак" {
		return MsgNoBackpack
	}

	s.removeItemFromRoom(room, item, place)

	s.Player.Inventory = append(s.Player.Inventory, item)

	s.updateRoomDescriptionIfEmpty(room)

	return fmt.Sprintf(MsgItemAdded, itemName)
}

func (s *State) findItemWithLocation(room *entity.Room, itemName string) (*entity.Item, string) {
	for place, items := range room.Items {
		for _, item := range items {
			if item.Name == itemName {
				return item, place
			}
		}
	}
	return nil, ""
}

func (s *State) findItemInRoom(room *entity.Room, itemName string) *entity.Item {
	for _, items := range room.Items {
		for _, item := range items {
			if item.Name == itemName {
				return item
			}
		}
	}
	return nil
}

func (s *State) removeItemFromRoom(room *entity.Room, item *entity.Item, place string) {
	for i, it := range room.Items[place] {
		if it == item {
			room.Items[place] = append(room.Items[place][:i], room.Items[place][i+1:]...)
			break
		}
	}
}

func (s *State) updateRoomDescriptionIfEmpty(room *entity.Room) {
	hasItems := false
	for _, items := range room.Items {
		if len(items) > 0 {
			hasItems = true
			break
		}
	}
	if !hasItems {
		room.Description = "пустая комната"
	}
}

func (s *State) handleWear(itemName string) string {
	room := s.Player.CurrentRoom

	item, place, inRoom := s.findItemForWearing(room, itemName)
	if item == nil {
		return MsgItemNotFound
	}

	if item.Name != "рюкзак" && !item.HasTrait("wearable") {
		return MsgCannotWear
	}

	s.removeItemForWearing(room, item, place, inRoom)

	s.Player.WornItems = append(s.Player.WornItems, item)

	if inRoom {
		s.updateRoomDescriptionIfEmpty(room)
	}

	return fmt.Sprintf(MsgWearing, itemName)
}

func (s *State) findItemForWearing(room *entity.Room, itemName string) (*entity.Item, string, bool) {
	item, place := s.findItemWithLocation(room, itemName)
	if item != nil {
		return item, place, true
	}

	for _, item := range s.Player.Inventory {
		if item.Name == itemName {
			return item, "", false
		}
	}

	return nil, "", false
}

func (s *State) removeItemForWearing(room *entity.Room, item *entity.Item, place string, inRoom bool) {
	if inRoom {
		for i, it := range room.Items[place] {
			if it == item {
				room.Items[place] = append(room.Items[place][:i], room.Items[place][i+1:]...)
				break
			}
		}
	} else {
		for i, it := range s.Player.Inventory {
			if it == item {
				s.Player.Inventory = append(s.Player.Inventory[:i], s.Player.Inventory[i+1:]...)
				break
			}
		}
	}
}

func (s *State) handleUse(itemName, targetName string) string {
	item := s.findItemInInventory(itemName)
	if item == nil {
		return fmt.Sprintf(MsgNoItemInInventory, itemName)
	}

	var target *entity.Item

	target = s.findItemInRoom(s.Player.CurrentRoom, targetName)

	if target == nil && s.Rooms["коридор"] != nil {
		target = s.findItemInRoom(s.Rooms["коридор"], targetName)
	}

	if target == nil {
		return MsgNothingToApply
	}

	return s.ApplyInteraction(item, target)
}

func (s *State) findItemInInventory(itemName string) *entity.Item {
	for _, item := range s.Player.Inventory {
		if item.Name == itemName {
			return item
		}
	}
	return nil
}
