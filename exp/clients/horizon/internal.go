package horizonclient

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/stellar/go/support/errors"
)

func decodeResponse(resp *http.Response, object interface{}) (err error) {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	// resp.Request should not be nil for Client requests
	if resp.Request != nil {
		setCurrentServerTime(resp.Request.Host, resp.Header["Date"])
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		horizonError := &Error{
			Response: resp,
		}
		decodeError := decoder.Decode(&horizonError.Problem)
		if decodeError != nil {
			return errors.Wrap(decodeError, "error decoding horizon.Problem")
		}
		return horizonError
	}

	err = decoder.Decode(&object)
	if err != nil {
		return
	}
	return
}

func countParams(params ...interface{}) int {
	counter := 0
	for _, param := range params {
		switch param := param.(type) {
		case string:
			if param != "" {
				counter++
			}
		case int:
			if param > 0 {
				counter++
			}
		case uint:
			if param > 0 {
				counter++
			}
		case bool:
			counter++
		default:
			panic("Unknown parameter type")
		}

	}
	return counter
}

func addQueryParams(params ...interface{}) string {
	query := url.Values{}

	for _, param := range params {
		switch param := param.(type) {
		case cursor:
			if param != "" {
				query.Add("cursor", string(param))
			}
		case Order:
			if param != "" {
				query.Add("order", string(param))
			}
		case limit:
			if param != 0 {
				query.Add("limit", strconv.Itoa(int(param)))
			}
		case assetCode:
			if param != "" {
				query.Add("asset_code", string(param))
			}
		case assetIssuer:
			if param != "" {
				query.Add("asset_issuer", string(param))
			}
		case includeFailed:
			if param {
				query.Add("include_failed", "true")
			}
		case map[string]string:
			for key, value := range param {
				if value != "" {
					query.Add(key, value)
				}
			}
		default:
			panic("Unknown parameter type")
		}
	}

	return query.Encode()
}

func setCurrentServerTime(host string, serverDate []string) {
	st, err := time.Parse(time.RFC1123, serverDate[0])
	if err != nil {
		return
	}
	ServerTimeMap[host] = ServerTimeRecord{ServerTime: st.Unix(), LocalTimeRecorded: time.Now().Unix()}
}

func currentServerTime(host string) int64 {
	st := ServerTimeMap[host]
	if &st == nil {
		return 0
	}

	currentTime := time.Now().Unix()
	// if it's been more than 5 minutes from the last time, then return 0
	if currentTime-st.LocalTimeRecorded > 60*5 {
		return 0
	}

	return currentTime - st.LocalTimeRecorded + st.ServerTime
}
