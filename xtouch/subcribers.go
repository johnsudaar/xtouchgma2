package xtouch

import "context"

type FaderChangedListener func(context.Context, FaderChangedEvent)
type ButtonChangedListener func(context.Context, ButtonChangedEvent)

func (s *Server) SubscribeToFaderChanges(l FaderChangedListener) {
	s.listenerLock.Lock()
	defer s.listenerLock.Unlock()
	s.faderChangedListeners = append(s.faderChangedListeners, l)
}

func (s *Server) sendFaderChange(ctx context.Context, e FaderChangedEvent) {
	s.listenerLock.RLock()
	defer s.listenerLock.RUnlock()
	for _, l := range s.faderChangedListeners {
		go l(ctx, e)
	}
}

func (s *Server) SubscribeButtonChange(l ButtonChangedListener) {
	s.listenerLock.Lock()
	defer s.listenerLock.Unlock()
	s.buttonChangedListeners = append(s.buttonChangedListeners, l)
}

func (s *Server) sendButtonChange(ctx context.Context, e ButtonChangedEvent) {
	s.listenerLock.RLock()
	defer s.listenerLock.RUnlock()
	for _, l := range s.buttonChangedListeners {
		go l(ctx, e)
	}
}
