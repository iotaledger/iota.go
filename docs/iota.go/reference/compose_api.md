# ComposeAPI

ComposeAPI composes a new API from the given settings and provider.
If no provider function is supplied, then the default http provider is used.
Settings must not be nil.

> **Important note:** This API is currently in Beta and is subject to change. Use of these APIs in production applications is not supported.

## Input

| Parameter       | Type | Required or Optional | Description |
|:---------------|:--------|:--------| :--------|
| settings  | Settings| Required | Settings for provider creation.  |
| createProvider  | CreateProviderFunc | Optional | Function which creates a new Provider given the settings.  |

## Output

| Return type     | Description |
|:---------------|:--------|
| *API  | Composed API object. |
| error  |  |

| Exceptions     | Description |
|:---------------|:--------|
| Exception  | Description of the exception |

## (Optional) Related APIs (link to other product documentation)

| API     | Description |
|:---------------|:--------|
| [API name link]()  | Description of the API  |

## (Optional) Permissions and authentication

|Permission type      | Permissions (from least to most privileged)              |
|:--------------------|:---------------------------------------------------------|
| permission type | List of permissions    |
| permission type | List of permissions    |
| permission type | List of permissions    |

## Example

Example call formatted in the correct language block.

```java

```

### Result

Result as returned by the API. Formatted in the correct language block.

```java

```