package web

import (
//	"fmt"
	"net/http"
	"testing"
	"github.com/GlenKelley/battleref/testing"
	"encoding/json"
)

func TestWebError(test * testing.T) {
	t := (*testutil.T)(test)
	err := NewError(http.StatusInternalServerError, "Validation errors")
	err.AddError(NewErrorItem("Missing field", fmt.Sprintf("Missing required field %v", "foobar"), "foobar", "formfield"));
	if _, e2 := json.Marshal(Json{nil, err}); e2 != nil {
		t.ErrorNow(e2)
	}
}
