package music

import (
	"context"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"log"
	"os"
)

var ytService *youtube.Service

func getYouTubeService() *youtube.Service {
	if ytService != nil {
		return ytService
	}

	token := os.Getenv("YOUTUBE_TOKEN")
	if token == "" {
		log.Fatalf("YOUTUBE_TOKEN not set")
	}

	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(token))

	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	ytService = youtubeService
	return youtubeService
}

func SearchYoutube(query string) (*Song, error) {
	yt := getYouTubeService()

	res, err := yt.Search.List([]string{"id", "snippet"}).Q(query).MaxResults(3).RegionCode("FI").Do()
	if err != nil {
		log.Fatalf("Error searching: %v", err)

		return nil, err
	}

	song := &Song{
		Title:     res.Items[0].Snippet.Title,
		Link:      "https://www.youtube.com/watch?v=" + res.Items[0].Id.VideoId,
		Thumbnail: res.Items[0].Snippet.Thumbnails.Default.Url,
	}

	return song, nil
}
