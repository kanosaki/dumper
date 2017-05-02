package tumblr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/kurrik/oauth1a"
)

//Make queries to the Tumblr API through TumblrRequest.
type TumblrRequest struct {
	service    *oauth1a.Service
	userConfig *oauth1a.UserConfig
	host       string
	apiKey     string
}

func NewTumblrRequest(consumerKey, consumerSecret, oauthToken, oauthSecret, callbackUrl, host string) *TumblrRequest {
	service := &oauth1a.Service{
		RequestURL:   "http://www.tumblr.com/oauth/request_token",
		AuthorizeURL: "http://www.tumblr.com/oauth/authorize",
		AccessURL:    "http://www.tumblr.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    consumerKey,
			ConsumerSecret: consumerSecret,
			CallbackURL:    callbackUrl,
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}
	userConfig := oauth1a.NewAuthorizedConfig(oauthToken, oauthSecret)
	return &TumblrRequest{service, userConfig, host, consumerKey}
}

//Make a GET request to the API with properly formatted parameters.
//requestUrl: the url you are making the request to.
//params: the parameters needed for the request.
func (tr *TumblrRequest) Get(ctx context.Context, requestUrl string, params map[string]string, ret *BaseResponse) error {
	fullUrl := tr.host + requestUrl
	if len(params) != 0 {
		values := url.Values{}
		for key, value := range params {
			values.Set(key, value)
		}
		fullUrl = fullUrl + "?" + values.Encode()
	}
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	tr.service.Sign(req, tr.userConfig)
	httpClient := new(http.Client)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(ret); err != nil {
		return err
	}
	if err := ret.Meta.Err(); err != nil {
		return err
	}
	return nil
}

func (tr *TumblrRequest) Post(ctx context.Context, requestUrl string, params map[string]string, ret *BaseResponse) error {
	fullUrl := tr.host + requestUrl
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	req, err := http.NewRequest("POST", fullUrl, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tr.service.Sign(req, tr.userConfig)
	httpClient := new(http.Client)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(ret); err != nil {
		return err
	}
	if err := ret.Meta.Err(); err != nil {
		return err
	}
	return nil
}
