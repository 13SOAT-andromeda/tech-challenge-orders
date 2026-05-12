package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/13SOAT-andromeda/tech-challenge-s1/internal/adapter/config"
)

func SetupTest() (*config.Config, error) {

	err := initProjectRoot()

	if err != nil {
		return nil, err
	}

	cfg, err := config.Init()

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func initProjectRoot() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil
	}

	dir := filepath.Dir(filename)
	projectRoot := filepath.Join(dir, "..", "..")

	if err := os.Chdir(projectRoot); err != nil {
		return err
	}

	return nil
}

func NewUnauthenticatedReq(method, url string, body io.Reader) (*http.Response, error) {
	client := &http.Client{}

	if body == nil {
		body = http.NoBody
	}

	request, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")

	return client.Do(request)
}

// NewIdentifiedReq creates a request with Lambda Authorizer identity headers.
// userID, email and role are the values the Lambda would inject after JWT validation.
func NewIdentifiedReq(method, url string, body io.Reader, userID, email, role string) (*http.Response, error) {
	client := &http.Client{}

	if body == nil {
		body = http.NoBody
	}

	request, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-User-ID", userID)
	request.Header.Add("X-User-Email", email)
	request.Header.Add("X-User-Role", role)

	return client.Do(request)
}

func GetApiUrl(cfg config.Config) string {
	return cfg.Http.Url + ":" + cfg.Http.Port + "/api"
}

func ParseBody[T any](resp *http.Response, data *T) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyString := string(bodyBytes)
	return json.Unmarshal([]byte(bodyString), &data)
}

func BuildBody[T any](data T) (io.Reader, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonData), nil
}

func GenerateValidCPF(seed int64) string {
	base := seed % 1000000000
	if base < 100000000 {
		base += 100000000
	}

	baseStr := strconv.FormatInt(base, 10)

	firstChar := baseStr[0]
	allSame := true
	for i := 1; i < len(baseStr); i++ {
		if baseStr[i] != firstChar {
			allSame = false
			break
		}
	}

	if allSame {
		lastDigit, _ := strconv.Atoi(string(baseStr[8]))
		lastDigit = (lastDigit + 1) % 10
		baseStr = baseStr[:8] + strconv.Itoa(lastDigit)
	}

	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(baseStr[i]))
		sum += digit * (10 - i)
	}

	firstDigit := (sum * 10) % 11
	if firstDigit == 10 {
		firstDigit = 0
	}

	baseStr += strconv.Itoa(firstDigit)
	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(baseStr[i]))
		sum += digit * (11 - i)
	}

	secondDigit := (sum * 10) % 11
	if secondDigit == 10 {
		secondDigit = 0
	}

	cpf := baseStr + strconv.Itoa(secondDigit)

	firstChar = cpf[0]
	allSame = true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != firstChar {
			allSame = false
			break
		}
	}

	if allSame {
		return GenerateValidCPF(seed + 1)
	}

	return cpf
}

func GenerateValidPlate(seed int64) string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	letter1 := letters[seed%26]
	letter2 := letters[(seed/26)%26]
	letter3 := letters[(seed/676)%26]

	numbers := seed % 10000
	if numbers < 0 {
		numbers = -numbers
	}

	numbersStr := strconv.FormatInt(numbers, 10)
	for len(numbersStr) < 4 {
		numbersStr = "0" + numbersStr
	}

	return string(letter1) + string(letter2) + string(letter3) + numbersStr
}
