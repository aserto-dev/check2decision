# check2decision

Converts directory check assertions into authorizer check_decision assertions

## Usage

```
Usage: check2decision [flags]

converts directory check assertions into authorizer check_decision assertions

Flags:
  -h, --help                          Show context-sensitive help.
  -i, --input=STRING                  assertions file path
  -o, --output=STRING                 decisions file path
      --policy-name="policy-rebac"    policy name
      --policy-path="rebac.check"     policy package path
      --policy-rule="allowed"         policy rule name
      --identity-type="sub"           identity type (sub|jwt|manual|none)
      --stdin                         read input from StdIn
      --version                       version info
```

## Example

Install the gdrive template using:

```
topaz templates install gdrive --force
```

Execute directory check assertions

```
topaz ds test exec $(topaz config info config.topaz_tmpl_dir -r)/gdrive/assertions/gdrive_assertions.json
```

Convert directory check assertions into authorizer decisions

```
GDRIVE_ASSERTIONS_DIR=$(topaz config info config.topaz_tmpl_dir -r)/gdrive/assertions
check2decision -i ${GDRIVE_ASSERTIONS_DIR}/gdrive_assertions.json -o ${GDRIVE_ASSERTIONS_DIR}/gdrive_decisions.json
```

Execute authorizer decision assertions

```
topaz az test exec ${GDRIVE_ASSERTIONS_DIR}/gdrive_decisions.json
```

## Installation

