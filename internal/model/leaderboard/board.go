package leaderboard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Board struct {
	ETag       string   `json:"-"`
	Year       string   `json:"year"`
	Week       string   `json:"week"`
	Algorithms []string `json:"algorithms"`

	Entries BoardEntry `json:"entries"`
}

type BoardEntry struct {
	UserId   string  `json:"user_id"`
	Username string  `json:"username"`
	Rating   float64 `json:"rating"`
}

func (m *Model) GetBoard(ctx context.Context, date time.Time) (*Board, bool, error) {
	year, week := date.ISOWeek()
	key := fmt.Sprintf("%d-%d.json", year, week)
	result, err := m.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(fmt.Sprint(m.prefix, key)),
	})
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}
	defer result.Body.Close()

	rawBoard, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}

	board := &Board{ETag: *result.ETag}
	err = json.Unmarshal(rawBoard, board)
	if err != nil {
		return nil, false, connect.NewError(connect.CodeInternal, err)
	}
	return board, true, nil
}

func (m *Model) PutBoard(ctx context.Context, date time.Time, board *Board) error {
	rawBoard, err := json.Marshal(board)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	year, week := date.ISOWeek()
	key := fmt.Sprintf("%d-%d.json", year, week)
	_, err = m.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:  aws.String(m.bucket),
		Key:     aws.String(fmt.Sprint(m.prefix, key)),
		Body:    bytes.NewReader(rawBoard),
		IfMatch: aws.String(board.ETag),
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
