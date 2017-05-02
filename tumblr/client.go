package tumblr

import (
	"context"
	"encoding/json"
)

type TumblrClient struct {
	request *TumblrRequest
}

func NewTumblrClient(consumerKey, consumerSecret, oauthToken, oauthSecret, callbackUrl, host string) *TumblrClient {
	return &TumblrClient{
		request: NewTumblrRequest(consumerKey, consumerSecret, oauthToken, oauthSecret, callbackUrl, host),
	}
}

func (tc *TumblrClient) Dashboard(ctx context.Context, dp *DashboardParams) (*DashboardResponse, error) {
	var base BaseResponse
	if err := tc.request.Get(ctx, "/v2/user/dashboard", dp.Params(), &base); err != nil {
		return nil, err
	}
	var dr DashboardResponse
	if err := json.Unmarshal(base.Response, &dr); err != nil {
		return nil, err
	}
	return &dr, nil
}

func (tc *TumblrClient) Info(ctx context.Context) (*Info, error) {
	var base BaseResponse
	if err := tc.request.Get(ctx, "/v2/user/info", nil, &base); err != nil {
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(base.Response, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
