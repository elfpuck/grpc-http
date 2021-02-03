package elfHttp

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http/httptest"
	"testing"
)

func BenchmarkPingMethod(b *testing.B) {
	params := struct{}{}
	paramsByte, _ := json.Marshal(params)
	req := httptest.NewRequest("POST", "/abc/abc", bytes.NewBuffer(paramsByte))
	e := New()
	e.Use(responseFormatMv(), recoveryMv(), appendCtxHandlersMv())
	g := e.Service("abc")
	g.Method("abc", func(c *Ctx) {
		c.Result(nil, errors.New("abc"))
		return
	})
	w := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		e.ServeHTTP(w, req)
	}
}
