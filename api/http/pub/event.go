package pub

import (
	"encoding/base64"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type event struct {
	Id           string               `json:"id"`
	SpecVersion1 *string              `json:"specVersion,omitempty"`
	SpecVersion2 *string              `json:"spec_version,omitempty"`
	Source       string               `json:"source"`
	Type         string               `json:"type"`
	Attributes   map[string]attribute `json:"attributes"`
	TextData1    *string              `json:"textData,omitempty"`
	TextData2    *string              `json:"text_data,omitempty"`
}

type attribute struct {
	CeBoolean1   *bool      `json:"ceBoolean,omitempty"`
	CeBoolean2   *bool      `json:"ce_boolean,omitempty"`
	CeBytes1     *string    `json:"ceBytes,omitempty"`
	CeBytes2     *string    `json:"ce_bytes,omitempty"`
	CeInteger1   *int32     `json:"ceInteger,omitempty"`
	CeInteger2   *int32     `json:"ce_integer,omitempty"`
	CeString1    *string    `json:"ceString,omitempty"`
	CeString2    *string    `json:"ce_string,omitempty"`
	CeTimestamp1 *time.Time `json:"ceTimestamp,omitempty"`
	CeTimestamp2 *time.Time `json:"ce_timestamp,omitempty"`
	CeUri1       *string    `json:"ceUri,omitempty"`
	CeUri2       *string    `json:"ce_uri,omitempty"`
	CeUriRef1    *string    `json:"ceUriRef,omitempty"`
	CeUriRef2    *string    `json:"ce_uri_ref,omitempty"`
}

type eventBatch struct {
	Events []event `json:"events"`
}

func Unmarshal(src []byte, dst *pb.CloudEvent) (err error) {
	var raw event
	err = sonic.Unmarshal(src, &raw)
	if err == nil {
		err = convert(raw, dst)
	}
	return
}

func UnmarshalBatch(src []byte) (dstBatch []*pb.CloudEvent, err error) {
	var rawBatch eventBatch
	err = sonic.Unmarshal(src, &rawBatch)
	if err == nil {
		for _, raw := range rawBatch.Events {
			var dst pb.CloudEvent
			err = convert(raw, &dst)
			if err != nil {
				break
			}
			dstBatch = append(dstBatch, &dst)
		}
	}
	return
}

func convert(raw event, dst *pb.CloudEvent) (err error) {

	dst.Id = raw.Id
	dst.Source = raw.Source
	dst.Type = raw.Type

	if raw.SpecVersion1 != nil {
		dst.SpecVersion = *raw.SpecVersion1
	}
	if raw.SpecVersion2 != nil {
		dst.SpecVersion = *raw.SpecVersion2
	}

	if raw.TextData1 != nil {
		dst.Data = &pb.CloudEvent_TextData{
			TextData: *raw.TextData1,
		}
	}
	if raw.TextData2 != nil {
		dst.Data = &pb.CloudEvent_TextData{
			TextData: *raw.TextData2,
		}
	}

	dst.Attributes = make(map[string]*pb.CloudEventAttributeValue)
	for name, srcAttr := range raw.Attributes {

		var dstAttr pb.CloudEventAttributeValue

		switch {
		case srcAttr.CeBoolean1 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeBoolean{
				CeBoolean: *srcAttr.CeBoolean1,
			}
		case srcAttr.CeBoolean2 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeBoolean{
				CeBoolean: *srcAttr.CeBoolean2,
			}
		case srcAttr.CeBytes1 != nil:
			var bytes []byte
			bytes, err = base64.StdEncoding.DecodeString(*srcAttr.CeBytes1)
			if err == nil {
				dstAttr.Attr = &pb.CloudEventAttributeValue_CeBytes{
					CeBytes: bytes,
				}
			}
		case srcAttr.CeBytes2 != nil:
			var bytes []byte
			bytes, err = base64.StdEncoding.DecodeString(*srcAttr.CeBytes2)
			if err == nil {
				dstAttr.Attr = &pb.CloudEventAttributeValue_CeBytes{
					CeBytes: bytes,
				}
			}
		case srcAttr.CeInteger1 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: *srcAttr.CeInteger1,
			}
		case srcAttr.CeInteger2 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: *srcAttr.CeInteger2,
			}
		case srcAttr.CeString1 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeString{
				CeString: *srcAttr.CeString1,
			}
		case srcAttr.CeString2 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeString{
				CeString: *srcAttr.CeString2,
			}
		case srcAttr.CeTimestamp1 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(*srcAttr.CeTimestamp1),
			}
		case srcAttr.CeTimestamp2 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeTimestamp{
				CeTimestamp: timestamppb.New(*srcAttr.CeTimestamp2),
			}
		case srcAttr.CeUri1 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeUri{
				CeUri: *srcAttr.CeUri1,
			}
		case srcAttr.CeUri2 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeUri{
				CeUri: *srcAttr.CeUri2,
			}
		case srcAttr.CeUriRef1 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeUriRef{
				CeUriRef: *srcAttr.CeUriRef1,
			}
		case srcAttr.CeUriRef2 != nil:
			dstAttr.Attr = &pb.CloudEventAttributeValue_CeUriRef{
				CeUriRef: *srcAttr.CeUriRef2,
			}
		default:
			err = fmt.Errorf("failed to convert event %s, unknown attribute type: %+v", raw.Id, srcAttr)
		}

		if err != nil {
			break
		}
		dst.Attributes[name] = &dstAttr
	}

	return
}
