package federation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tubblelive/star/internal/outbound"
)

func doJSON(request *http.Request, out any) error {
	request.Header.Set("Accept", "application/json")

	response, err := outbound.Public.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
	if err != nil {
		return err
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%s %s responded with %d: %s", request.Method, request.URL.Host, response.StatusCode, body[:min(len(body), 256)])
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	return decoder.Decode(out)
}
