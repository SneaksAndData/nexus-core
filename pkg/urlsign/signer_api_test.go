package urlsign

import (
	"fmt"
	"math/rand"
	"net/url"
	"testing"
	"time"
)

func generateUrls(numUrls int) []string {
	rnd := rand.New(rand.NewSource(1))

	urls := make([]string, numUrls)
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	bucket := randomString(rnd, 8, charset)

	for i := 0; i < numUrls; i++ {
		prefix1 := randomString(rnd, 5, charset)

		path := fmt.Sprintf("https://s3.svc.local", bucket, prefix1)
		// add one more prefix randomly
		if rnd.Intn(2) == 1 {
			path += "/" + randomString(rnd, 5, charset)
		}

		// add query params
		result, _ := url.Parse(path)
		query := result.Query()

		// 1 = one param, 2 = two params
		numParams := rnd.Intn(2)
		for paramIdx := 0; paramIdx < numParams; paramIdx++ {
			paramKey := fmt.Sprintf("param%d", paramIdx)
			paramValue := randomString(rnd, 5, charset)
			query.Add(paramKey, paramValue)
		}

		result.RawQuery = query.Encode()
		urls[i] = result.String()
	}

	return urls
}

func randomString(random *rand.Rand, desiredLength int, charset string) string {
	result := make([]uint8, desiredLength)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}

func TestSignVerify(t *testing.T) {
	secret := []byte("secret")
	for _, unsigned := range generateUrls(5) {
		parsed, err := url.Parse(unsigned)
		if err != nil {
			t.Fatalf("failed to parse url: %s", err)
		}
		signed, err := Sign(*parsed, "test", time.Hour, "test", secret)
		Verify(*signed.Url, secret)
	}
}
