package urlsign

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
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

		path := fmt.Sprintf("https://s3.svc.local/%s/%s", bucket, prefix1)
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
		if err != nil {
			t.Fatalf("failed to sign: %s", err)
		}

		if signed == nil {
			t.Fatalf("nil signature generated")
		}

		err = Verify(*signed.Url, secret)

		if err != nil {
			t.Fatalf("invalid signature: %s", err)
		}
	}
}

func TestSignVerifyInvalidIfNoPath(t *testing.T) {
	parsed, _ := url.Parse("https://s3.svc.local/")

	_, err := Sign(*parsed, "test", time.Hour, "test", []byte("secret"))

	if err == nil {
		t.Fatalf("successfully signed an invalid url")
	}
}

func TestVerifyTampered(t *testing.T) {
	secret := []byte("secret")
	for _, unsigned := range generateUrls(5) {
		parsed, _ := url.Parse(unsigned)
		signed, _ := Sign(*parsed, "test", time.Hour, "test", secret)
		sigQuery := signed.Url.Query()
		sig := signed.Url.Query().Get("sig")
		sigQuery.Set("sig", fmt.Sprintf("abc123%s", sig))

		signed.Url.RawQuery = sigQuery.Encode()

		err := Verify(*signed.Url, secret)

		if err == nil {
			t.Fatalf("a tampered signature verification should fail")
		}
	}
}

func TestVerifyEmptySignature(t *testing.T) {
	parsed, _ := url.Parse("https://s3.svc.local/bucket/prefix?exp=1111&tid=123&chk=3123qdawqe")

	err := Verify(*parsed, []byte("test"))

	if err == nil {
		t.Fatalf("successfully signed an invalid url")
	}

	if !strings.Contains(err.Error(), "invalid signature") {
		t.Fatalf("incorrect error: %s", err)
	}
}

func TestVerifyExpiry(t *testing.T) {
	parsed, _ := url.Parse("https://s3.svc.local/bucket/prefix1/prefix2/file.json")

	signed, err := Sign(*parsed, "test", time.Second, "test", []byte("secret"))

	time.Sleep(time.Second * 2)

	err = Verify(*signed.Url, []byte("test"))

	if err == nil {
		t.Fatalf("successfully signed an expired signature")
	}

	if !strings.Contains(err.Error(), "url is no longer valid") {
		t.Fatalf("incorrect error: %s", err)
	}
}
