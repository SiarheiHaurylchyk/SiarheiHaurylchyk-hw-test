package internalgrpc

import (
	"time"

	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/SiarheiHaurylchyk/SiarheiHaurylchyk-hw-test/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func eventToPB(event storage.Event) *pb.Event {
	return &pb.Event{
		Id:           event.ID,
		Title:        event.Title,
		StartTime:    timestamppb.New(event.StartTime),
		EndTime:      timestamppb.New(event.EndTime),
		Description:  event.Description,
		UserId:       event.UserID,
		NotifyBefore: durationpb.New(event.NotifyBefore),
	}
}

func eventsToPB(events []storage.Event) []*pb.Event {
	result := make([]*pb.Event, 0, len(events))
	for _, event := range events {
		result = append(result, eventToPB(event))
	}
	return result
}

func eventFromPB(event *pb.Event) storage.Event {
	if event == nil {
		return storage.Event{}
	}
	return storage.Event{
		ID:           event.GetId(),
		Title:        event.GetTitle(),
		StartTime:    timestampToTime(event.GetStartTime()),
		EndTime:      timestampToTime(event.GetEndTime()),
		Description:  event.GetDescription(),
		UserID:       event.GetUserId(),
		NotifyBefore: durationToTime(event.GetNotifyBefore()),
	}
}

func timestampToTime(value *timestamppb.Timestamp) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.AsTime()
}

func durationToTime(value *durationpb.Duration) time.Duration {
	if value == nil {
		return 0
	}
	return value.AsDuration()
}
