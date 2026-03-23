package automation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"superhue/internal/domain"
)

type RuleExecutor interface {
	ExecuteActions(ctx context.Context, actions []domain.Action) error
}

type Logger interface {
	Info(ctx context.Context, source, message string)
	Error(ctx context.Context, source, message string)
}

type Engine struct {
	rulesRepo domain.RuleRepository
	executor  RuleExecutor
	logger    Logger
	events    chan domain.RuleEvent
	once      sync.Once
}

func NewEngine(rulesRepo domain.RuleRepository, executor RuleExecutor, logger Logger) *Engine {
	return &Engine{rulesRepo: rulesRepo, executor: executor, logger: logger, events: make(chan domain.RuleEvent, 64)}
}

func (e *Engine) Events() chan<- domain.RuleEvent { return e.events }

func (e *Engine) Start(ctx context.Context) {
	e.once.Do(func() {
		go e.run(ctx)
		go e.scheduleLoop(ctx)
	})
}

func (e *Engine) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-e.events:
			e.handleEvent(ctx, event)
		}
	}
}

func (e *Engine) scheduleLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	lastMinute := ""
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			minute := now.Format("15:04")
			if minute == lastMinute {
				continue
			}
			lastMinute = minute
			e.handleEvent(ctx, domain.RuleEvent{Trigger: domain.TriggerTimeSchedule, Name: "time", Value: minute})
		}
	}
}

func (e *Engine) handleEvent(ctx context.Context, event domain.RuleEvent) {
	rules, err := e.rulesRepo.List(ctx)
	if err != nil {
		e.logger.Error(ctx, "automation", fmt.Sprintf("no se pudieron cargar reglas: %v", err))
		return
	}
	for _, rule := range rules {
		if !rule.Enabled || rule.Trigger != event.Trigger {
			continue
		}
		if !matches(rule, event) {
			continue
		}
		if err := e.executor.ExecuteActions(ctx, rule.Actions); err != nil {
			e.logger.Error(ctx, "automation", fmt.Sprintf("regla %s falló: %v", rule.Name, err))
			continue
		}
		e.logger.Info(ctx, "automation", fmt.Sprintf("Regla ejecutada: %s", rule.Name))
	}
}

func matches(rule domain.Rule, event domain.RuleEvent) bool {
	for _, cond := range rule.Conditions {
		matched := false
		switch cond.Type {
		case domain.ConditionProcessName:
			matched = strings.EqualFold(cond.Value, event.Value)
		case domain.ConditionScheduleAt:
			matched = cond.Value == event.Value
		case domain.ConditionDeviceState:
			matched = strings.EqualFold(cond.Key, event.Name) && strings.EqualFold(cond.Value, event.Value)
		}
		if cond.Negate {
			matched = !matched
		}
		if !matched {
			return false
		}
	}
	return true
}
