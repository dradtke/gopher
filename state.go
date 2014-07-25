package allegory

import (
	"reflect"
	"runtime"
)

// GameState is an interface to the game's current state. Only one game
// state is active at any point in time, and states can be changed
// by using either NewState() or NewStateNow().
type GameState interface {
	// Perform initialization; this method is called once, when the
	// state becomes the game state.
	InitState()

	// Called once per frame to perform any necessary updates.
	UpdateState()

	// Render; this is called (ideally) once per frame, with a delta
	// value calculated based on lag.
	RenderState(delta float32)

	// Perform cleanup; this method is called once, when the state
	// has been replaced by another one.
	CleanupState()
}

// NewState() waits for all processes to finish without
// blocking the current goroutine, then changes the game state.
func NewState(state GameState, views ...View) {
	go func() {
		for _processes.Len() > 0 {
			runtime.Gosched()
		}
		setState(state, views...)
	}()
}

// NewStateNow() tells all processes to quit,
// waits for them to finish, then changes the game state.
func NewStateNow(state GameState, views ...View) {
	NotifyAllProcesses(quit{})
	for _processes.Len() > 0 {
		runtime.Gosched()
	}
	setState(state, views...)
}

type BaseState struct{}

func (s *BaseState) InitState()                {}
func (s *BaseState) UpdateState()              {}
func (s *BaseState) RenderState(delta float32) {}
func (s *BaseState) CleanupState()             {}

var _ GameState = (*BaseState)(nil)

func setState(state GameState, views ...View) {
	if _state != nil {
		_state.CleanupState()
	}
	for e := _views.Front(); e != nil; e = e.Next() {
		e.Value.(View).CleanupView()
	}

	_state = state
	_state.InitState()
	_views.Init()

	if views != nil {
		stateVal := reflect.ValueOf(state)
		for _, v := range views {
			assignStateField(stateVal, reflect.ValueOf(v))
			v.InitView()
			_views.PushBack(v)
		}
	}
}

var stateType = reflect.TypeOf((*GameState)(nil)).Elem()

// If the view embeds BaseView, set its State field.
func assignStateField(stateVal, viewVal reflect.Value) {
	for viewVal.Kind() == reflect.Interface || viewVal.Kind() == reflect.Ptr {
		viewVal = viewVal.Elem()
	}
	if _, ok := viewVal.Type().FieldByName("BaseView"); ok {
		base := viewVal.FieldByName("BaseView")
		if _, ok := base.Type().FieldByName("State"); ok {
			s := base.FieldByName("State")
			if s.Type().Implements(stateType) {
				s.Set(stateVal)
			}
		}
	}
}
