package game

import (
	"github.com/AgDecode/mini-game/entity"
)

type CommandHandler func(*State, []string) string

type InteractionRule struct {
	SourceTraits  map[string]interface{}
	TargetTraits  map[string]interface{}
	ResultHandler func(*State, *entity.Item, *entity.Item) string
	StateModifier func(*State, *entity.Item, *entity.Item)
	EventEmitter  func(*State, *entity.Item, *entity.Item)
}

type State struct {
	Player           *entity.Player
	Rooms            map[string]*entity.Room
	DoorOpened       bool
	LastCommand      string
	EventEmitter     *entity.EventEmitter
	Commands         map[string]CommandHandler
	InteractionRules []InteractionRule
}

func NewState() *State {
	state := &State{
		Rooms:            make(map[string]*entity.Room),
		EventEmitter:     entity.NewEventEmitter(),
		Commands:         make(map[string]CommandHandler),
		InteractionRules: make([]InteractionRule, 0),
	}

	return state
}

func (s *State) RegisterCommand(command string, handler CommandHandler) {
	if s.Commands == nil {
		s.Commands = make(map[string]CommandHandler)
	}
	s.Commands[command] = handler
}

func (s *State) RegisterInteractionRule(rule InteractionRule) {
	if s.InteractionRules == nil {
		s.InteractionRules = make([]InteractionRule, 0)
	}
	s.InteractionRules = append(s.InteractionRules, rule)
}

func (s *State) CheckInteraction(source, target *entity.Item) (bool, InteractionRule) {
	for _, rule := range s.InteractionRules {
		matches := true

		for trait, expectedValue := range rule.SourceTraits {
			actualValue := source.GetTrait(trait)
			if actualValue != expectedValue {
				matches = false
				break
			}
		}

		if !matches {
			continue
		}

		for trait, expectedValue := range rule.TargetTraits {
			actualValue := target.GetTrait(trait)
			if actualValue != expectedValue {
				matches = false
				break
			}
		}

		if matches {
			return true, rule
		}
	}

	return false, InteractionRule{}
}

func (s *State) ApplyInteraction(source, target *entity.Item) string {
	canInteract, rule := s.CheckInteraction(source, target)
	if !canInteract {
		return "нельзя применить"
	}

	if rule.StateModifier != nil {
		rule.StateModifier(s, source, target)
	}

	if rule.EventEmitter != nil {
		rule.EventEmitter(s, source, target)
	}

	if rule.ResultHandler != nil {
		return rule.ResultHandler(s, source, target)
	}

	return "применено"
}
