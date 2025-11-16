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
	ETag       string           `json:"-"`
	Year       string           `json:"year"`
	Week       string           `json:"week"`
	Algorithms map[string]int64 `json:"algorithms"`

	Entries map[string]BoardEntry `json:"entries"`
}

type BoardEntry struct {
	UserId   string            `json:"user_id"`
	Username string            `json:"username"`
	Streak   int64             `json:"streak"`
	Rating   map[int64]float64 `json:"rating"`
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
	key := fmt.Sprintf("%s%d-%d.json", m.prefix, year, week)
	input := &s3.PutObjectInput{
		Bucket:  aws.String(m.bucket),
		Key:     aws.String(key),
		Body:    bytes.NewReader(rawBoard),
		IfMatch: aws.String(board.ETag),
	}
	if board.ETag == "" {
		input.IfNoneMatch = aws.String(key)
	} else {
		input.IfMatch = aws.String(board.ETag)
	}
	_, err = m.s3Client.PutObject(ctx, input)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	return nil
}
