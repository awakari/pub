package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/pub/config"
	"github.com/awakari/pub/model"
	"github.com/awakari/pub/util"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/drankou/go-vader/vader"
	"github.com/pebbe/textcat"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Service interface {
	Preprocess(ctx context.Context, evt *pb.CloudEvent, groupId, userId string, internal bool) error
}

type svc struct {
	blacklist         model.Prefixes[model.BlacklistValue]
	cfgPreproc        config.PreprocessConfig
	cfgWriterInternal config.WriterInternalConfig
	sia               *vader.SentimentIntensityAnalyzer
	txtCat            *textcat.TextCat
}

const sentimentPrefix = "sentiment"
const sentimentFactor = 100
const langCodeLength = 2

var ErrRejected = errors.New("event rejected by prefix")

func New(
	blacklist model.Prefixes[model.BlacklistValue],
	cfgPreproc config.PreprocessConfig,
	cfgWriterInternal config.WriterInternalConfig,
	sia *vader.SentimentIntensityAnalyzer,
	txtCat *textcat.TextCat,
) Service {
	return svc{
		blacklist:         blacklist,
		cfgPreproc:        cfgPreproc,
		cfgWriterInternal: cfgWriterInternal,
		sia:               sia,
		txtCat:            txtCat,
	}
}

func (s svc) Preprocess(ctx context.Context, evt *pb.CloudEvent, groupId, userId string, internal bool) (err error) {
	if evt.Attributes == nil {
		evt.Attributes = make(map[string]*pb.CloudEventAttributeValue)
	}
	switch internal {
	case true:
		s.setInternal(evt)
	default:
		err = s.checkBlacklist(ctx, evt)
		if err == nil {
			if s.cfgPreproc.Snippets.Enabled {
				snippet := s.snippet(evt)
				s.setLanguageIfNotSet(evt, snippet)
				if s.cfgPreproc.Sentiments.Enabled {
					s.setSentiments(evt, snippet)
				}
			}
		}
	}
	if err == nil {
		setAuth(evt, groupId, userId)
		setTime(evt)
	}
	return
}

func (s svc) setInternal(evt *pb.CloudEvent) {
	evt.Attributes[s.cfgWriterInternal.Name] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeInteger{
			CeInteger: s.cfgWriterInternal.Value,
		},
	}
	return
}

func (s svc) checkBlacklist(ctx context.Context, evt *pb.CloudEvent) (err error) {
	var prefix string
	var attrName string
	var attrValue string
	prefix, _, _ = s.blacklist.FindOnePrefix(ctx, "source:"+evt.Source)
	switch prefix {
	case "":
		prefix, _, _ = s.blacklist.FindOnePrefix(ctx, "type:"+evt.Source)
	default:
		attrName = "source"
		attrValue = evt.Source
	}
	switch prefix {
	case "":
		for k, v := range evt.Attributes {
			switch vt := v.Attr.(type) {
			case *pb.CloudEventAttributeValue_CeString:
				attrValue = vt.CeString
			case *pb.CloudEventAttributeValue_CeUri:
				attrValue = vt.CeUri
			case *pb.CloudEventAttributeValue_CeUriRef:
				attrValue = vt.CeUriRef
			}
			if attrValue != "" {
				prefix, _, _ = s.blacklist.FindOnePrefix(ctx, k+":"+attrValue)
				if prefix != "" {
					attrName = k
					break
				}
			}
		}
	default:
		attrName = "type"
		attrValue = evt.Type
	}
	if prefix != "" {
		err = fmt.Errorf("%s: %s, id: %s, attribute: %s=%s\n", ErrRejected, prefix, evt.Id, attrName, attrValue)
	}
	return
}

func setAuth(evt *pb.CloudEvent, groupId, userId string) {
	evt.Attributes[model.KeyCeGroupId] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: groupId,
		},
	}
	evt.Attributes[model.KeyCeUserId] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: userId,
		},
	}
}

func setTime(evt *pb.CloudEvent) {
	evt.Attributes[model.KeyCePubTime] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeTimestamp{
			CeTimestamp: timestamppb.New(time.Now().UTC()),
		},
	}
}

func (s svc) snippet(evt *pb.CloudEvent) (result string) {
	result = util.ExtractTextSnippet(evt, s.cfgPreproc.Snippets.Length.Max)
	evt.Attributes[model.KeyCeSnippet] = &pb.CloudEventAttributeValue{
		Attr: &pb.CloudEventAttributeValue_CeString{
			CeString: result,
		},
	}
	return
}

func (s svc) setLanguageIfNotSet(evt *pb.CloudEvent, snippet string) {
	var lang string
	langAttr, langAttrPresent := evt.Attributes[model.KeyCeLanguage]
	if langAttrPresent {
		lang = langAttr.GetCeString()
	}
	switch {
	case lang == "":
		langsDetected, _ := s.txtCat.Classify(snippet)
		for _, langDetected := range langsDetected {
			if len(langDetected) >= langCodeLength {
				evt.Attributes[model.KeyCeLanguage] = &pb.CloudEventAttributeValue{
					Attr: &pb.CloudEventAttributeValue_CeString{
						CeString: langDetected[:langCodeLength],
					},
				}
				break
			}
		}
	case len(lang) > langCodeLength:
		// fix
		evt.Attributes[model.KeyCeLanguage] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: lang[:langCodeLength],
			},
		}
	}
	return
}

func (s svc) setSentiments(evt *pb.CloudEvent, snippet string) {
	scores := s.sia.PolarityScores(snippet)
	for k, v := range scores {
		attrName := sentimentPrefix
		switch k {
		case "pos":
			attrName += "positive"
		case "neg":
			attrName += "negative"
		case "neu":
			attrName += "neutral"
		case "compound":
		default:
			panic("unexpected sentiment score key: " + k)
		}
		evt.Attributes[attrName] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeInteger{
				CeInteger: int32(sentimentFactor * v),
			},
		}
	}
}
