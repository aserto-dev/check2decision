package resource

import (
	"google.golang.org/protobuf/types/known/structpb"
)

type Context struct {
	ObjectType string `json:"object_type"`
	ObjectID   string `json:"object_id"`
	Relation   string `json:"relation"`
}

func (r Context) Struct() *structpb.Struct {
	return &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"object_type": structpb.NewStringValue(r.ObjectType),
			"object_id":   structpb.NewStringValue(r.ObjectID),
			"relation":    structpb.NewStringValue(r.Relation),
		},
	}
}
