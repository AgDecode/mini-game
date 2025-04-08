package game

import (
	"fmt"
	"github.com/AgDecode/mini-game/entity"
	"strings"
)

func (s *State) HandleCommand(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "неизвестная команда"
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "осмотреться":
		return s.handleLook()
	case "идти":
		if len(args) == 0 {
			return "не указано направление"
		}
		return s.handleGo(args[0])
	case "взять":
		if len(args) == 0 {
			return "не указан предмет"
		}
		return s.handleTake(args[0])
	case "надеть":
		if len(args) == 0 {
			return "не указан предмет"
		}
		return s.handleWear(args[0])
	case "применить":
		if len(args) < 2 {
			return "не указаны предметы"
		}
		return s.handleUse(args[0], args[1])
	default:
		return "неизвестная команда"
	}
}

func (s *State) handleLook() string {
	room := s.Player.CurrentRoom
	var parts []string

	if room.Name == "кухня" {
		parts = append(parts, "ты находишься на кухне")
	}

	var itemParts []string

	if items, ok := room.Items["столе"]; ok && len(items) > 0 {
		itemNames := make([]string, len(items))
		for i, item := range items {
			itemNames[i] = item.Name
		}
		itemParts = append(itemParts, fmt.Sprintf("на столе: %s", strings.Join(itemNames, ", ")))
	}

	if items, ok := room.Items["стуле"]; ok && len(items) > 0 {
		itemNames := make([]string, len(items))
		for i, item := range items {
			itemNames[i] = item.Name
		}
		itemParts = append(itemParts, fmt.Sprintf("на стуле: %s", strings.Join(itemNames, ", ")))
	}

	if len(itemParts) > 0 {
		parts = append(parts, strings.Join(itemParts, ", "))
	} else if room.Name != "кухня" {
		parts = append(parts, "пустая комната")
	}

	if room.HasHint {
		if room.Name == "кухня" {
			if s.Player.HasBackpack() {
				parts = append(parts, "надо идти в универ")
			} else {
				parts = append(parts, "надо собрать рюкзак и идти в универ")
			}
		}
	}

	if len(room.Connections) > 0 {
		directions := make([]string, 0, len(room.Connections))
		for dir := range room.Connections {
			directions = append(directions, dir)
		}
		parts = append(parts, fmt.Sprintf("можно пройти - %s", strings.Join(directions, ", ")))
	}

	if room.Name == "кухня" {
		return strings.Join(parts[:len(parts)-1], ", ") + ". " + parts[len(parts)-1]
	}
	return strings.Join(parts, ". ")
}

func (s *State) handleGo(direction string) string {
	room := s.Player.CurrentRoom
	if nextRoom, ok := room.Connections[direction]; ok {
		if nextRoom.Name == "улица" && !s.DoorOpened {
			return "дверь закрыта"
		}

		s.Player.CurrentRoom = nextRoom
		s.LastCommand = "идти"

		event := &entity.Event{
			Type:   entity.EventType("enter_room"),
			Source: s.Player,
			Target: nextRoom,
			Data:   make(map[string]interface{}),
		}
		err := s.EventEmitter.Emit(event)
		if err != nil {
			return ""
		}

		switch nextRoom.Name {
		case "коридор":
			return "ничего интересного. можно пройти - кухня, комната, улица"
		case "комната":
			return "ты в своей комнате. можно пройти - коридор"
		case "кухня":
			return "кухня, ничего интересного. можно пройти - коридор"
		case "улица":
			return "на улице весна. можно пройти - домой"
		default:
			return s.handleLook()
		}
	}
	return "нет пути в " + direction
}

func (s *State) handleTake(itemName string) string {
	room := s.Player.CurrentRoom

	var item *entity.Item
	var place string
	for p, items := range room.Items {
		for _, it := range items {
			if it.Name == itemName {
				item = it
				place = p
				break
			}
		}
		if item != nil {
			break
		}
	}

	if item == nil {
		return "нет такого"
	}

	hasBackpack := s.Player.HasBackpack()

	// Для рюкзака не нужен рюкзак :)
	if !hasBackpack && item.Name != "рюкзак" {
		return "некуда класть"
	}

	for i, it := range room.Items[place] {
		if it == item {
			room.Items[place] = append(room.Items[place][:i], room.Items[place][i+1:]...)
			break
		}
	}

	s.Player.Inventory = append(s.Player.Inventory, item)

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

	return fmt.Sprintf("предмет добавлен в инвентарь: %s", itemName)
}

func (s *State) handleWear(itemName string) string {
	room := s.Player.CurrentRoom
	var item *entity.Item
	var place string
	var inRoom bool = false

	for p, items := range room.Items {
		for _, it := range items {
			if it.Name == itemName {
				item = it
				place = p
				inRoom = true
				break
			}
		}
		if item != nil {
			break
		}
	}

	if item == nil {
		for _, it := range s.Player.Inventory {
			if it.Name == itemName {
				item = it
				break
			}
		}
	}

	if item == nil {
		return "нет такого"
	}

	if item.Name != "рюкзак" && !item.HasTrait("wearable") {
		return "нельзя надеть"
	}

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

	s.Player.WornItems = append(s.Player.WornItems, item)

	if inRoom {
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

	return fmt.Sprintf("вы надели: %s", itemName)
}

func (s *State) handleUse(itemName, targetName string) string {
	if itemName == "ключи" && targetName == "дверь" && s.DoorOpened {
		return "дверь открыта"
	}

	var item *entity.Item
	for _, invItem := range s.Player.Inventory {
		if invItem.Name == itemName {
			item = invItem
			break
		}
	}
	if item == nil {
		return "нет предмета в инвентаре - " + itemName
	}

	if targetName == "дверь" {
		if item.Name == "ключи" {
			event := &entity.Event{
				Type:   entity.EventType("use"),
				Source: item,
				Target: &entity.Item{Name: "дверь"},
				Data:   make(map[string]interface{}),
			}
			err := s.EventEmitter.Emit(event)
			if err != nil {
				return ""
			}

			s.DoorOpened = true

			return "дверь открыта"
		}
		return "нельзя применить"
	}

	return "не к чему применить"
}
