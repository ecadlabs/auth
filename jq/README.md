# JSON filters

Expression is represented by JSON object with exactly one property:

```
{"type": ...}
```

Expression types:

| Type | Description               | Syntax                     |
| ---- | ------------------------- | -------------------------- |
| eq   | Equal                     | `{"property": value}`      |
| ne   | Not equal                 | `{"property": value}`      |
| lt   | Less than                 | `{"property": value}`      |
| gt   | Greater than              | `{"property": value}`      |
| le   | Less or equal             | `{"property": value}`      |
| ge   | Greater or equal          | `{"property": value}`      |
| re   | POSIX Regex               | `{"property": value}`      |
| l    | SQL LIKE                  | `{"property": value}`      |
| p    | Has prefix                | `{"property": value}`      |
| s    | Has suffix                | `{"property": value}`      |
| sub  | Has substring             | `{"property": value}`      |
| has  | Collection contains value | `{"property": value}`      |
| not  | Negation                  | *Expression* (see above)   |
| and  | Logical conjunction       | [*Expression*] (see above) |
| or   | Logical disjunction       | [*Expression*] (see above) |

Example:

```json
{
	"and": [
		{
			"eq": {
				"propertyA": "valueA"
			}
		},
		{
			"gt": {
				"propertyB": 10
			}
		},        
		{
			"or": [
				{
					"eq": {
						"propertyC": "valueC"
					}
				},
				{
					"ne": {
						"propertyD": 0
					}
				}
			]
		},
        {
            "not": {
                "eq": {
    				"propertyE": "valueE"
    			}
            }
        }
	]
}
```
