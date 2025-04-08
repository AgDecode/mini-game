package entity

import "fmt"

type EventType string

const (
	EventItemPicked  EventType = "item_picked"
	EventItemDropped EventType = "item_dropped"
	EventItemUsed    EventType = "item_used"
	EventRoomEntered EventType = "room_entered"
	EventRoomExited  EventType = "room_exited"
)

type Event struct {
	Type      EventType
	Source    interface{} // Объект, вызвавший событие
	Target    interface{} // Объект, на который направлено событие
	Data      map[string]interface{}
	Prevented bool
}

type EventHandler func(*Event) error

type EventEmitter struct {
	handlers map[EventType][]EventHandler
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		handlers: make(map[EventType][]EventHandler),
	}
}

func (e *EventEmitter) On(eventType EventType, handler EventHandler) {
	if e.handlers[eventType] == nil {
		e.handlers[eventType] = make([]EventHandler, 0)
	}
	e.handlers[eventType] = append(e.handlers[eventType], handler)
}

func (e *EventEmitter) Emit(event *Event) error {
	handlers := e.handlers[event.Type]
	if handlers == nil {
		return nil
	}

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			return fmt.Errorf("error handling event %s: %v", event.Type, err)
		}
		if event.Prevented {
			break
		}
	}
	return nil
}

func (e *EventEmitter) RemoveHandler(eventType EventType, handler EventHandler) {
	handlers := e.handlers[eventType]
	if handlers == nil {
		return
	}

	for i, h := range handlers {
		if fmt.Sprintf("%v", h) == fmt.Sprintf("%v", handler) {
			e.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}
