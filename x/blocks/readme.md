### Interactive UI in Cadence Workflows 

#### Status

September 29, 2025

this experimental and the API may change in future. The support for it is landed in the V4 web UI.

#### Background: 

The idea is that sometimes workflows require user input, and it'd be nice to provide a quick and easy way to accomplish this with the existing workflow state. 

Therefore, the query response rendered in the Cadence UI can render a query response with some buttons, giving the user a choice for how to interact.

Getting started

### 

Use the helpers `New` with `NewMarkdownSection` and similar to compose what UI you'd like to see returned from your workflow.

See this example:

```

	import (
		go.uber.org/cadence/x/blocks
	)

	func (w workflow) LunchSelectionWorkflow(ctx workflow.Ctx) error {

		votes := []map[string]string{}

		votesChan := workflow.GetSignalChannel(ctx, "lunch_order")
		workflow.Go(ctx, func(ctx workflow.Context) {
			for {
				var vote map[string]string
				votesChan.Receive(ctx, &vote)
				votes = append(votes, vote)
			}
		})
		defer func() {
			votesChan.Close()
		}()


		workflow.SetQueryHandler(ctx, "options", func() (LunchQueryResponse, error) {
			return blocks.New(
				NewMarkdownSection("## Lunch options\nWe're voting on where to order lunch today. Select the option you want to vote for."),
				NewDivider(),
				NewMarkdownSection("## Votes\n"),
				NewMarkdownSection(renderVotesTable(votes)),
				NewMarkdownSection("## Menu\n ... menu options"),
				NewSignalActions(
					NewSignalButton("Farmhouse", "lunch_order", map[string]string{"location": "farmhouse - red thai curry", "requests": "spicy"}),
					NewSignalButtonWithExternalWorkflow("Ethiopian", "no_lunch_order_walk_in_person", nil, "in-person-order-workflow", ""),
					NewSignalButton("Ler Ros", "lunch_order", map[string]string{"location": "Ler Ros", "meal": "tofo Bahn Mi"}),
				),
			)
		}
	}

	// This creates a markdown table and returns it as a string
	func renderVotesTable(votes []map[string]string) string {
		if len(votes) == 0 {
			return "| lunch order vote | meal | requests |\n|-------|-------|-------|\n| No votes yet |\n"
		}
		table := "| lunch order vote | meal | requests |\n|-------|-------|-------|\n"
		for _, vote := range votes {

			loc := vote["location"]
			meal := vote["meal"]
			requests := vote["requests"]

			table += "| " + loc + " | " + meal + " | " + requests + " |\n"
		}

		return table
	}
```

#### Specification of 'blocks' API 

### Example:

The buttons would signal the workflow, allowing the backend to then continue doing whatever user-specified behaviour. 

From the point of view of the workflow author, they would be looking to create a query with a response like this:

```json
{
   "cadenceResponseType": "formattedData",
   "format": "blocks",
   "blocks": [
		{
			"type": "section",
                    "format": "text/markdown",
                    "componentOptions": {
				"text": "### Lunch options\n\n|   |    |\n|---|----|\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Red_roast_duck_curry.jpg/200px-Red_roast_duck_curry.jpg) | <b>Farmhouse - Red Thai Curry</h5> </b>: The base Thai red curry paste ... |\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png/200px-B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png) | <b>Ler Ros: Lemongrass Tofu Bahn Mi</b> In Vietnamese cuisine, bánh mì, bánh mỳ or banh mi ... |\n\n\n\n\n"
      }
			}
		},
		{
"type": "divider"
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
					    "signal_value": { "location": "farmhouse - red thai curry", "requests": "spicy" }
                                 }
					
				},
				{
					"type": "button",
					"componentOptions": {
						"type": "plain_text",
						"text": "Kin Khao"
					},
                                 "action": {
						"type": "signal",
"signal_name": "no_lunch_order_walk_in_person",
                               	       "workflow_id": "some-other-workflow",
                                        "run_id": "49ea9236-0903-420a-8c26-09bbeeb6c50f"
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
					     "signal_value": { "location": "Ler Ros", "meal": "tofo Bahn Mi"}
                                 }
				}
			]
		}
	]
}

```

## Implementation and details

**Top level**
The top level has a section called ‘blocks’ which must always be an array/list of arbitrary length. Ie: 

```json
{
   "cadenceResponseType": "formattedData",
   "format": "blocks",
   "blocks": [...]
}
```

Where possible values for the type blocks are: `section`, `divider` and `actions`. They are CSS ‘block’ elements so they take up the entire div. Contents within typically are inline. 

**Section**

This is identical to the earlier proposal in that it is intended to be a formatted component with the intent of it being markdown, csv, svg etc (whatever existing components are supported). 

```c
{
  "type": "section",
  "format": "text/markdown",
  "componentOptions": {
 	 "text": "### Lunch options\n\n|   |    |\n|---|----|\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Red_roast_duck_curry.jpg/200px-Red_roast_duck_curry.jpg) | <b>Farmhouse - Red Thai Curry</h5> </b>: The base Thai red curry paste ... |\n| ![food](https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png/200px-B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png) | <b>Ler Ros: Lemongrass Tofu Bahn Mi</b> In Vietnamese cuisine, bánh mì, bánh mỳ or banh mi ... |\n\n\n\n\n"
  }
}
```

**Divider**  
This inserts a `<br />` or similar horizontal line for the aesthetic purpose of dividing content

```json
{
  "type": "divider"
}
```

**Actions**   
The actions section is an array of inputs, for now only buttons are proposed: 

```json
{
			"type": "actions",
			"elements": [ ... ]
}
```

And each button is proposed with the following set of fields

```json
{
  "type": "button",
  "componentOptions": {
    "type": "<TBA>",              // Baseweb button states such as highlighted or greyed out (design input needed here)
    "text": "Ler Ros"  
  },
  "action": {
     "type": "signal"                     // an identifyier of the type of action (other types might be added later)
     "signal_name": "lunch_order",        // the signal name
     "signal_value": {                    // the signal payload
        "location": "Ler Ros",
        "meal": "tofo Bahn Mi"
  },
  "workflow_id": "some-other-workflow",
  "run_id": "49ea9236-0903-420a-8c26-09bbeeb6c50f"
  }
}
```

For each button, in addition to the basic html fields (text etc) they would use additional data fields `signal` and `signal_value`, stored as props to the component which correspond to their signal API components. The user may optionally wish to specify the workflow/runID for the signal (TBA).

Clicking the button triggers a HTTP POST to the web backend using the existing signal API:

##### Security considerations: 

While the implementation in the UI attempts to mitigate some risks by only allowing a subset of full HTML functionality, this should still be used with care, particularly anywhere near untrusted input. An arbitrary input from the server-side into the query response might expose the user to XSS attacks where their cookies or security tokens risk being stolen.

This feature intentionally does not expose full html for this reason, but nevertheless, care should be taken to sanitize input before rendering it.
