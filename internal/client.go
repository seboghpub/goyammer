package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const YammerApiURL = "https://www.yammer.com/api/v1/"

// see: https://medium.com/@marcus.olsson/writing-a-go-client-for-your-restful-api-c193a2f4998c

type YammerMessageBody struct {
	Plain string `json:"plain"`
}

type YammerMessage struct {
	ID            int64             `json:"id"`
	SenderID      int64             `json:"sender_id"`
	RepliedToID   int64             `json:"replied_to_id"`
	CreatedAt     string            `json:"created_at"`
	SenderType    string            `json:"sender_type"`
	Body          YammerMessageBody `json:"body"`
	ThreadID      int64             `json:"thread_id"`
	ClientType    string            `json:"client_type"`
	ClientURL     string            `json:"client_url"`
	DirectMessage bool              `json:"direct_message"`
	Privacy       string            `json:"privacy"`
}

type YammerMessageResponse struct {
	Messages []YammerMessage `json:"messages"`
	Meta     struct {
		OlderAvailable        bool        `json:"older_available"`
		RequestedPollInterval int         `json:"requested_poll_interval"`
		LastSeenMessageID     interface{} `json:"last_seen_message_id"`
		UnseenThreadCount     int         `json:"unseen_thread_count"`
		CurrentUserID         int64       `json:"current_user_id"`
		FeedName              string      `json:"feed_name"`
		FeedDesc              string      `json:"feed_desc"`
	} `json:"meta"`
}

type YammerUserResponse struct {
	Type              string `json:"type"`
	ID                int64  `json:"id"`
	State             string `json:"state"`
	JobTitle          string `json:"job_title"`
	Location          string `json:"location"`
	FullName          string `json:"full_name"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	WebURL            string `json:"web_url"`
	Name              string `json:"name"`
	MugshotURL        string `json:"mugshot_url"`
	BirthDate         string `json:"birth_date"`
	BirthDateComplete string `json:"birth_date_complete"`
	Timezone          string `json:"timezone"`
	Email             string `json:"email"`
}

type YammerGroup struct {
	Type        string `json:"type"`
	ID          int64  `json:"id"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

type YammerGroupResponse []YammerGroup

type Client struct {
	httpClient *http.Client
	Token      string
	BaseURL    *url.URL
	UserAgent  string
}

func NewClient(token string) *Client {
	baseUrl, _ := url.Parse(YammerApiURL)
	return &Client{
		httpClient: http.DefaultClient,
		Token:      token,
		BaseURL:    baseUrl,
		UserAgent:  "goyammer",
	}
}

func (c *Client) newRequest(method, path string, query map[string]string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// add query parameters
	if query != nil {
		parameters := req.URL.Query()
		for key, value := range query {
			parameters.Add(key, value)
		}
		req.URL.RawQuery = parameters.Encode()

	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	log.Debug().Msg(fmt.Sprintf("url: %s", req.URL.String()))

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("response status %d", resp.StatusCode))
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

func (c *Client) GetImage(url string) ([]byte, error) {

	req, errReq := http.NewRequest(http.MethodGet, url, nil)
	if errReq != nil {
		return nil, fmt.Errorf("failed to construct mug shot request: %v", errReq)
	}
	resp, errDo := c.httpClient.Do(req)
	if errDo != nil {
		return nil, fmt.Errorf("failed to do mug shot request: %v", errDo)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("mug shot request response status %d", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return nil, fmt.Errorf("failed to read mug shot response body: %v", errRead)
	}

	return body, nil
}
