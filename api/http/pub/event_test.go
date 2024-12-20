package pub

import (
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEvent_Unmarshal(t *testing.T) {
	in := `{
  "id": "2qRgPkuvdypNEnrA7HdsWpASaco",
  "specVersion": "1.0",
  "source": "https://awakari.com/pub-msg.html",
  "type": "com_awakari_webapp",
  "attributes": {
    "time": {
      "ce_timestamp": "2024-12-19T17:52:59.216Z"
    },
    "string": {
      "ce_string": "v1"
    },
    "boolean": {
      "ce_boolean": true
    },
    "bytes": {
      "ce_bytes": "TWFueSBoYW5kcyBtYWtlIGxpZ2h0IHdvcmsu"
    },
    "integer": {
      "ce_integer": -42
    },
    "uri": {
      "ce_uri": "https://awakari.com/pub-msg.html"
    },
    "uriref": {
      "ce_uri_ref": "https://awakari.com/pub-msg.html"
    }
  },
  "text_data": "test"
}
`
	var out pb.CloudEvent
	err := Unmarshal([]byte(in), &out)
	require.NoError(t, err)
	assert.Equal(t, "2qRgPkuvdypNEnrA7HdsWpASaco", out.Id)
	assert.Equal(t, "com_awakari_webapp", out.Type)
	assert.Equal(t, "https://awakari.com/pub-msg.html", out.Source)
	assert.Equal(t, "1.0", out.SpecVersion)
	assert.Equal(t, "test", out.GetTextData())
	assert.Equal(t, true, out.Attributes["boolean"].GetCeBoolean())
	assert.Equal(t, []byte("Many hands make light work."), out.Attributes["bytes"].GetCeBytes())
	assert.Equal(t, int32(-42), out.Attributes["integer"].GetCeInteger())
	assert.Equal(t, "v1", out.Attributes["string"].GetCeString())
	assert.Equal(t, time.Date(2024, 12, 19, 17, 52, 59, 216_000_000, time.UTC), out.Attributes["time"].GetCeTimestamp().AsTime())
}
