package fakedata

import (
	"github.com/bxcodec/faker/v3"
)

func Tweets(count int, cb func(t *Tweet) error) error {
	var tw Tweet
	if err := faker.SetRandomMapAndSliceSize(3); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		if err := faker.FakeData(&tw); err != nil {
			return err
		}
		if err := cb(&tw); err != nil {
			return err
		}
	}
	return nil
}

type Tweet struct {
	CreatedAt            string `json:"created_at"`
	ID                   int64  `json:"id"`
	IDStr                string `json:"id_str"`
	Text                 string `json:"text"`
	Source               string `json:"source"`
	Truncated            bool   `json:"truncated"`
	InReplyToStatusID    int64  `json:"in_reply_to_status_id"`
	InReplyToStatusIDStr string `json:"in_reply_to_status_id_str"`
	InReplyToUserID      int64  `json:"in_reply_to_user_id"`
	InReplyToUserIDStr   string `json:"in_reply_to_user_id_str"`
	InReplyToScreenName  string `json:"in_reply_to_screen_name"`
	User                 struct {
		ID                             int64  `json:"id"`
		IDStr                          string `json:"id_str"`
		Name                           string `json:"name"`
		ScreenName                     string `json:"screen_name"`
		Location                       string `json:"location"`
		URL                            string `json:"url"`
		Description                    string `json:"description"`
		TranslatorType                 string `json:"translator_type"`
		Protected                      bool   `json:"protected"`
		Verified                       bool   `json:"verified"`
		FollowersCount                 int    `json:"followers_count"`
		FriendsCount                   int    `json:"friends_count"`
		ListedCount                    int    `json:"listed_count"`
		FavouritesCount                int    `json:"favourites_count"`
		StatusesCount                  int    `json:"statuses_count"`
		CreatedAt                      string `json:"created_at"`
		UtcOffset                      string `json:"utc_offset"`
		TimeZone                       string `json:"time_zone"`
		GeoEnabled                     bool   `json:"geo_enabled"`
		Lang                           string `json:"lang"`
		ContributorsEnabled            bool   `json:"contributors_enabled"`
		IsTranslator                   bool   `json:"is_translator"`
		ProfileBackgroundColor         string `json:"profile_background_color"`
		ProfileBackgroundImageURL      string `json:"profile_background_image_url"`
		ProfileBackgroundImageURLHTTPS string `json:"profile_background_image_url_https"`
		ProfileBackgroundTile          bool   `json:"profile_background_tile"`
		ProfileLinkColor               string `json:"profile_link_color"`
		ProfileSidebarBorderColor      string `json:"profile_sidebar_border_color"`
		ProfileSidebarFillColor        string `json:"profile_sidebar_fill_color"`
		ProfileTextColor               string `json:"profile_text_color"`
		ProfileUseBackgroundImage      bool   `json:"profile_use_background_image"`
		ProfileImageURL                string `json:"profile_image_url"`
		ProfileImageURLHTTPS           string `json:"profile_image_url_https"`
		ProfileBannerURL               string `json:"profile_banner_url"`
		DefaultProfile                 bool   `json:"default_profile"`
		DefaultProfileImage            bool   `json:"default_profile_image"`
	} `json:"user"`
	Geo struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"geo"`
	Coordinates struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"`
	} `json:"coordinates"`
	Place struct {
		ID          string `json:"id"`
		URL         string `json:"url"`
		PlaceType   string `json:"place_type"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		CountryCode string `json:"country_code"`
		Country     string `json:"country"`
		BoundingBox struct {
			Type        string        `json:"type"`
			Coordinates [][][]float64 `json:"coordinates"`
		} `json:"bounding_box"`
		Attributes struct {
		} `json:"attributes"`
	} `json:"place"`
	Contributors  string `json:"contributors"`
	IsQuoteStatus bool   `json:"is_quote_status"`
	QuoteCount    int    `json:"quote_count"`
	ReplyCount    int    `json:"reply_count"`
	RetweetCount  int    `json:"retweet_count"`
	FavoriteCount int    `json:"favorite_count"`
	Entities      struct {
		Hashtags     []string `json:"hashtags"`
		Urls         []string `json:"urls"`
		UserMentions []struct {
			ScreenName string `json:"screen_name"`
			Name       string `json:"name"`
			ID         int64  `json:"id"`
			IDStr      string `json:"id_str"`
			Indices    []int  `json:"indices"`
		} `json:"user_mentions",faker:"len=1"`
		Symbols []string `json:"symbols"`
		Media   []Media  `json:"media"`
	} `json:"entities"`
	ExtendedEntities struct {
		Media []Media `json:"media"`
	} `json:"extended_entities"`
	Favorited         bool   `json:"favorited"`
	Retweeted         bool   `json:"retweeted"`
	PossiblySensitive bool   `json:"possibly_sensitive"`
	FilterLevel       string `json:"filter_level"`
	Lang              string `json:"lang"`
	TimestampMs       string `json:"timestamp_ms"`
}

type Media struct {
	ID            int64  `json:"id"`
	IDStr         string `json:"id_str"`
	Indices       []int  `json:"indices"`
	MediaURL      string `json:"media_url"`
	MediaURLHTTPS string `json:"media_url_https"`
	URL           string `json:"url"`
	DisplayURL    string `json:"display_url"`
	ExpandedURL   string `json:"expanded_url"`
	Type          string `json:"type"`
	Sizes         struct {
		Thumb struct {
			W      int    `json:"w"`
			H      int    `json:"h"`
			Resize string `json:"resize"`
		} `json:"thumb"`
		Small struct {
			W      int    `json:"w"`
			H      int    `json:"h"`
			Resize string `json:"resize"`
		} `json:"small"`
		Medium struct {
			W      int    `json:"w"`
			H      int    `json:"h"`
			Resize string `json:"resize"`
		} `json:"medium"`
		Large struct {
			W      int    `json:"w"`
			H      int    `json:"h"`
			Resize string `json:"resize"`
		} `json:"large"`
	} `json:"sizes"`
	SourceStatusID    int64  `json:"source_status_id"`
	SourceStatusIDStr string `json:"source_status_id_str"`
	SourceUserID      int64  `json:"source_user_id"`
	SourceUserIDStr   string `json:"source_user_id_str"`
}
