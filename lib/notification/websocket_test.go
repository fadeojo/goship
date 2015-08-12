package notification

import (
	"io"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/websocket"
)

type stubServer struct {
	location url.URL
}

func (s *stubServer) Dial(t *testing.T) *websocket.Conn {
	const origin = "http://origin.example"
	conn, err := websocket.Dial(s.location.String(), "", origin)
	if err != nil {
		t.Fatalf("websocket.Dial(%q, %q, %q) failed with %v; want success", s.location.String(), "", origin, err)
	}
	return conn
}

func withStubServer(t *testing.T, h *Hub, f func(s *stubServer)) {
	s := httptest.NewServer(websocket.Handler(h.AcceptConnection))
	defer s.Close()
	u, err := url.Parse(s.URL)
	if err != nil {
		t.Fatalf("url.Parse(%q) failed with %v; want success", s.URL)
	}
	u.Scheme = "ws"

	f(&stubServer{location: *u})
}

func TestBroadcast(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	h := NewHub(ctx)
	withStubServer(t, h, func(s *stubServer) {
		ws := s.Dial(t)
		defer ws.Close()

		msgs := []string{
			"example 1",
			"example 2",
			"example 3",
		}
		done := make(chan struct{})
		go func() {
			defer close(done)
			for _, want := range msgs {
				var msg string
				if err := websocket.Message.Receive(ws, &msg); err != nil {
					t.Errorf("websocket.Message.Recieve(ws, &got) failed with %v; want success", err)
				}
				if got := msg; got != want {
					t.Errorf("msg = %q; want %q", got, want)
				}
			}
		}()

		for _, msg := range msgs {
			h.Broadcast(msg)
		}

		select {
		case <-time.NewTimer(100 * time.Millisecond).C:
			t.Errorf("done timed out; want closed")
		case <-done:
		}
	})
}

func TestBroadcastWithMultipleConnections(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	h := NewHub(ctx)
	withStubServer(t, h, func(s *stubServer) {
		msgs := []string{
			"example 1",
			"example 2",
			"example 3",
		}
		var done []<-chan struct{}
		for i := 0; i < 5; i++ {
			ws := s.Dial(t)
			if i == 2 {
				// one client terminates earlier
				ws.Close()
				continue
			}
			defer ws.Close()

			ch := make(chan struct{})
			go func(i int) {
				defer close(ch)
				for _, want := range msgs {
					var msg string
					if err := websocket.Message.Receive(ws, &msg); err != nil {
						t.Errorf("websocket.Message.Recieve(ws, &got) failed with %v; want success; i=%d", err, i)
					}
					if got := msg; got != want {
						t.Errorf("msg = %q; want %q; i=%d", got, want, i)
					}
				}
			}(i)
			done = append(done, ch)
		}

		for _, msg := range msgs {
			h.Broadcast(msg)
		}

		for _, ch := range done {
			select {
			case <-time.NewTimer(100 * time.Millisecond).C:
				t.Errorf("ch timed out; want closed")
			case <-ch:
			}
		}
	})
}

func TestBroadcastWithoutActiveConnection(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	h := NewHub(ctx)
	h.Broadcast("example message")
}

func TestContextCancel(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	h := NewHub(ctx)
	withStubServer(t, h, func(s *stubServer) {
		ws := s.Dial(t)
		defer ws.Close()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				var msg string
				switch err := websocket.Message.Receive(ws, &msg); err {
				case nil:
					continue
				case io.EOF:
					return
				default:
					t.Errorf("websocket.Message.Recieve(ws, &got) failed; want nil or io.EOF", err)
					return
				}
			}
		}()

		go func() {
			for {
				select {
				case <-done:
					return
				default:
					h.Broadcast("example message")
				}
			}
		}()

		cancel()
		select {
		case <-time.NewTimer(100 * time.Millisecond).C:
			t.Errorf("done timed out; want closed")
		case <-done:
		}
	})
}

func TestBroadCastWithStuckClient(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	h := NewHub(ctx)
	withStubServer(t, h, func(s *stubServer) {
		ws := s.Dial(t)
		defer ws.Close()

		for i := 0; i < 100000; i++ {
			h.Broadcast("example message")
		}
	})
}
