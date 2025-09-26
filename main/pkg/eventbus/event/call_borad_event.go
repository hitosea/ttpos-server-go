package event

import (
	"ttpos-server-go/pkg/eventbus"
)

type CallBoardChangeEvent struct {
	CompanyUuid uint64
}

func NewRefreshCallBoardEvent(companyUuid uint64) CallBoardChangeEvent {
	return CallBoardChangeEvent{
		CompanyUuid: companyUuid,
	}
}

type CallBoardChangeHandler func(msg CallBoardChangeEvent)

func (e *SystemEventBus) SubscribeCallBoardChangeEvent(handler CallBoardChangeHandler) {
	e.bus.Subscribe(string(EventCallBoardChange), func(event eventbus.Event) {
		msg, ok := event.Payload.(CallBoardChangeEvent)
		if !ok {
			return
		}
		handler(msg)
	})
}

func (e *SystemEventBus) PublishCallBoardChangeEvent(msg CallBoardChangeEvent) {
	e.bus.Publish(eventbus.Event{
		Name:    string(EventCallBoardChange),
		Payload: msg,
	})
}

type CallBoardLanguageChangeEvent struct {
	CompanyUuid uint64
}
type CallBoardLanguageChangeHandler func(msg CallBoardLanguageChangeEvent)

func (e *SystemEventBus) SubscribeCallBoardLanguageChangeEvent(handler CallBoardLanguageChangeHandler) {
	e.bus.Subscribe(string(EventCallBoardLanguageChange), func(event eventbus.Event) {
		msg, ok := event.Payload.(CallBoardLanguageChangeEvent)
		if !ok {
			return
		}
		handler(msg)
	})
}

func (e *SystemEventBus) PublishCallBoardLanguageChangeEvent(msg CallBoardLanguageChangeEvent) {
	e.bus.Publish(eventbus.Event{
		Name:    string(EventCallBoardLanguageChange),
		Payload: msg,
	})
}
