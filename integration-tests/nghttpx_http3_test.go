//go:build quic

package nghttp2

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"testing"

	"golang.org/x/net/http2/hpack"
)

// TestH3H1PlainGET tests whether simple HTTP/3 GET request works.
func TestH3H1PlainGET(t *testing.T) {
	st := newServerTester(t, options{
		quic: true,
	})
	defer st.Close()

	res, err := st.http3(requestParam{
		name: "TestH3H1PlainGET",
	})
	if err != nil {
		t.Fatalf("Error st.http3() = %v", err)
	}

	want := 200
	if got := res.status; got != want {
		t.Errorf("status = %v; want %v", got, want)
	}
}

// TestH3H1RequestBody tests HTTP/3 request with body works.
func TestH3H1RequestBody(t *testing.T) {
	body := make([]byte, 3333)
	_, err := rand.Read(body)
	if err != nil {
		t.Fatalf("Unable to create request body: %v", err)
	}

	opts := options{
		handler: func(w http.ResponseWriter, r *http.Request) {
			buf := make([]byte, 4096)
			buflen := 0
			p := buf

			for {
				if len(p) == 0 {
					t.Fatal("Request body is too large")
				}

				n, err := r.Body.Read(p)

				p = p[n:]
				buflen += n

				if err != nil {
					if err == io.EOF {
						break
					}

					t.Fatalf("r.Body.Read() = %v", err)
				}
			}

			buf = buf[:buflen]

			if got, want := buf, body; !bytes.Equal(got, want) {
				t.Fatalf("buf = %v; want %v", got, want)
			}
		},
		quic: true,
	}
	st := newServerTester(t, opts)
	defer st.Close()

	res, err := st.http3(requestParam{
		name: "TestH3H1RequestBody",
		body: body,
	})
	if err != nil {
		t.Fatalf("Error st.http3() = %v", err)
	}
	if got, want := res.status, 200; got != want {
		t.Errorf("res.status: %v; want %v", got, want)
	}
}

// TestH3H1GenerateVia tests that server generates Via header field to
// and from backend server.
func TestH3H1GenerateVia(t *testing.T) {
	opts := options{
		handler: func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Header.Get("Via"), "3 nghttpx"; got != want {
				t.Errorf("Via: %v; want %v", got, want)
			}
		},
		quic: true,
	}
	st := newServerTester(t, opts)
	defer st.Close()

	res, err := st.http3(requestParam{
		name: "TestH3H1GenerateVia",
	})
	if err != nil {
		t.Fatalf("Error st.http3() = %v", err)
	}
	if got, want := res.header.Get("Via"), "1.1 nghttpx"; got != want {
		t.Errorf("Via: %v; want %v", got, want)
	}
}

// TestH3H1AppendVia tests that server adds value to existing Via
// header field to and from backend server.
func TestH3H1AppendVia(t *testing.T) {
	opts := options{
		handler: func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Header.Get("Via"), "foo, 3 nghttpx"; got != want {
				t.Errorf("Via: %v; want %v", got, want)
			}
			w.Header().Add("Via", "bar")
		},
		quic: true,
	}
	st := newServerTester(t, opts)
	defer st.Close()

	res, err := st.http3(requestParam{
		name: "TestH3H1AppendVia",
		header: []hpack.HeaderField{
			pair("via", "foo"),
		},
	})
	if err != nil {
		t.Fatalf("Error st.http3() = %v", err)
	}
	if got, want := res.header.Get("Via"), "bar, 1.1 nghttpx"; got != want {
		t.Errorf("Via: %v; want %v", got, want)
	}
}

// TestH3H1NoVia tests that server does not add value to existing Via
// header field to and from backend server.
func TestH3H1NoVia(t *testing.T) {
	opts := options{
		args: []string{"--no-via"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			if got, want := r.Header.Get("Via"), "foo"; got != want {
				t.Errorf("Via: %v; want %v", got, want)
			}
			w.Header().Add("Via", "bar")
		},
		quic: true,
	}
	st := newServerTester(t, opts)
	defer st.Close()

	res, err := st.http3(requestParam{
		name: "TestH3H1NoVia",
		header: []hpack.HeaderField{
			pair("via", "foo"),
		},
	})
	if err != nil {
		t.Fatalf("Error st.http3() = %v", err)
	}
	if got, want := res.header.Get("Via"), "bar"; got != want {
		t.Errorf("Via: %v; want %v", got, want)
	}
}
