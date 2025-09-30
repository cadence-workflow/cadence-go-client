package blocks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func make() string {

	r := New(
		NewMarkdownSection("## Lunch options\nWe're voting on where to order lunch today. Select the option you want to vote for."),
		NewDivider(),
		NewMarkdownSection("## Votes\n ... vote table"),
		NewMarkdownSection("## Menu\n ... menu options"),
		NewSignalActions(
			NewSignalButton("Farmhouse", "lunch_order", map[string]string{"location": "farmhouse - red thai curry", "requests": "spicy"}),
			NewSignalButtonWithExternalWorkflow("Ethiopian", "no_lunch_order_walk_in_person", nil, "in-person-order-workflow", ""),
			NewSignalButton("Ler Ros", "lunch_order", map[string]string{"location": "Ler Ros", "meal": "tofo Bahn Mi"}),
		),
	)

	d, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	return string(d)
}

func TestExample(t *testing.T) {
	expectedJSON := `
{
  "cadenceResponseType": "formattedData",
  "format": "blocks",
  "blocks": [
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "## Lunch options\nWe're voting on where to order lunch today. Select the option you want to vote for."
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "## Votes\n ... vote table"
      }
    },
    {
      "type": "section",
      "format": "text/markdown",
      "componentOptions": {
        "text": "## Menu\n ... menu options"
      }
    },
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Farmhouse"
          },
          "action": {
            "type": "signal",
            "signal_name": "lunch_order",
            "signal_value": {
              "location": "farmhouse - red thai curry",
              "requests": "spicy"
            }
          }
        },
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Ethiopian"
          },
          "action": {
            "type": "signal",
            "signal_name": "no_lunch_order_walk_in_person",
            "workflow_id": "in-person-order-workflow"
          }
        },
        {
          "type": "button",
          "componentOptions": {
            "type": "plain_text",
            "text": "Ler Ros"
          },
          "action": {
            "type": "signal",
            "signal_name": "lunch_order",
            "signal_value": {
              "location": "Ler Ros",
              "meal": "tofo Bahn Mi"
            }
          }
        }
      ]
    }
  ]
}
`

	var expected interface{}
	_ = json.Unmarshal([]byte(expectedJSON), &expected)

	var actual interface{}

	example := make()
	_ = json.Unmarshal([]byte(example), &actual)

	assert.Equal(t, expected, actual)
}
