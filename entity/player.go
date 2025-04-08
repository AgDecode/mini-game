package entity

import "fmt"

type Player struct {
	CurrentRoom *Room
	Inventory   []*Item
	WornItems   []*Item
	Emitter     *EventEmitter
}

func NewPlayer(startRoom *Room) *Player {
	return &Player{
		CurrentRoom: startRoom,
		Inventory:   make([]*Item, 0),
		WornItems:   make([]*Item, 0),
		Emitter:     NewEventEmitter(),
	}
}

func (p *Player) Move(room *Room) string {
	if p.CurrentRoom != nil {
		p.CurrentRoom.OnExit(p)
	}
	p.CurrentRoom = room
	return room.OnEnter(p)
}

func (p *Player) HasBackpack() bool {
	for _, item := range p.WornItems {
		if item.Name == "рюкзак" {
			return true
		}
	}
	return false
}

func (p *Player) Take(item *Item) string {
	if !p.HasBackpack() && !item.IsWearable() {
		return "некуда класть"
	}

	if p.CurrentRoom.HasItem(item.Name) {
		for place, items := range p.CurrentRoom.Items {
			for _, it := range items {
				if it == item {
					p.CurrentRoom.RemoveItem(item, place)
					p.Inventory = append(p.Inventory, item)
					hasItems := false
					for _, items := range p.CurrentRoom.Items {
						if len(items) > 0 {
							hasItems = true
							break
						}
					}
					if !hasItems {
						p.CurrentRoom.Description = "пустая комната"
					}
					return fmt.Sprintf("предмет добавлен в инвентарь: %s", item.Name)
				}
			}
		}
	}
	return "нет такого"
}

func (p *Player) Drop(item *Item) string {
	for i, it := range p.Inventory {
		if it == item {
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			p.CurrentRoom.AddItem(item, "стол") // По умолчанию кладем на стол
			return item.OnDrop(p.CurrentRoom)
		}
	}
	return "нет такого"
}

func (p *Player) Wear(item *Item) string {
	if !item.HasTrait("wearable") {
		return "нельзя надеть"
	}

	if p.CurrentRoom.HasItem(item.Name) {
		item = p.CurrentRoom.GetItem(item.Name)
		for place, items := range p.CurrentRoom.Items {
			for _, it := range items {
				if it == item {
					p.CurrentRoom.RemoveItem(item, place)
					p.WornItems = append(p.WornItems, item)
					hasItems := false
					for _, items := range p.CurrentRoom.Items {
						if len(items) > 0 {
							hasItems = true
							break
						}
					}
					if !hasItems {
						p.CurrentRoom.Description = "пустая комната"
					}
					return fmt.Sprintf("вы надели: %s", item.Name)
				}
			}
		}
	}

	for i, it := range p.Inventory {
		if it == item {
			p.Inventory = append(p.Inventory[:i], p.Inventory[i+1:]...)
			p.WornItems = append(p.WornItems, item)
			return fmt.Sprintf("вы надели: %s", item.Name)
		}
	}
	return "нет такого"
}

func (p *Player) Use(item *Item, target interface{}) string {
	for _, it := range p.Inventory {
		if it == item {
			return item.Use(target)
		}
	}
	return "нет предмета в инвентаре - " + item.Name
}

func (p *Player) HasItem(itemName string) bool {
	for _, item := range p.Inventory {
		if item.Name == itemName {
			return true
		}
	}
	return false
}

func (p *Player) GetItem(itemName string) *Item {
	for _, item := range p.Inventory {
		if item.Name == itemName {
			return item
		}
	}
	return nil
}
