## Database documentation 

 Exracted in 86.162ms 

| Table | Columns | Records | First record | Row size | File size | Modified |
|---|---|---|---|---|---|---|
| [employees](#employees) | 16 | 3 | 808 | 523 | 2.4 kB | 2022-10-15 00:00:00 |
| [expense_categories](#expense_categories) | 3 | 5 | 392 | 59 | 687 B | 2022-10-15 00:00:00 |
| [expense_details](#expense_details) | 6 | 6 | 488 | 79 | 962 B | 2022-10-15 00:00:00 |
| [expense_reports](#expense_reports) | 9 | 3 | 584 | 140 | 1.0 kB | 2022-10-15 00:00:00 |


## Columns


### employees

| Name | Type | Golang type | Length | Comment |
|---|---|---|---|---|
| EMPLOYEEID | I | int32 | 4 |  |
| DEPARTMENT | C | string | 50 |  |
| SOCIALSECU | C | string | 30 |  |
| EMPLOYEENU | C | string | 30 |  |
| FIRSTNAME | C | string | 50 |  |
| LASTNAME | C | string | 50 |  |
| TITLE | C | string | 50 |  |
| EMAILNAME | C | string | 50 |  |
| EXTENSION | C | string | 30 |  |
| ADDRESS | M | string | 4 |  |
| CITY | C | string | 50 |  |
| STATEORPRO | C | string | 20 |  |
| POSTALCODE | C | string | 20 |  |
| COUNTRY | C | string | 50 |  |
| WORKPHONE | C | string | 30 |  |
| NOTES | M | string | 4 |  |

### expense_categories

| Name | Type | Golang type | Length | Comment |
|---|---|---|---|---|
| EXPENSECAT | I | int32 | 4 |  |
| EXPENSECA2 | C | string | 50 |  |
| EXPENSECA3 | I | int32 | 4 |  |

### expense_details

| Name | Type | Golang type | Length | Comment |
|---|---|---|---|---|
| EXPENSEDET | I | int32 | 4 |  |
| EXPENSEREP | I | int32 | 4 |  |
| EXPENSECAT | I | int32 | 4 |  |
| EXPENSEITE | Y | float64 | 8 |  |
| EXPENSEIT2 | C | string | 50 |  |
| EXPENSEDAT | T | time.Time | 8 |  |

### expense_reports

| Name | Type | Golang type | Length | Comment |
|---|---|---|---|---|
| EXPENSEREP | I | int32 | 4 |  |
| EMPLOYEEID | I | int32 | 4 |  |
| EXPENSETYP | C | string | 50 |  |
| EXPENSERPT | C | string | 30 |  |
| EXPENSERP2 | M | string | 4 |  |
| DATESUBMIT | T | time.Time | 8 |  |
| ADVANCEAMO | Y | float64 | 8 |  |
| DEPARTMENT | C | string | 30 |  |
| PAID | L | bool | 1 |  |
