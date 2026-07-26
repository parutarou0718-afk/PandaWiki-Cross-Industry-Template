# OpenAI-compatible Chat API

`POST /share/v1/chat/completions` accepts the standard OpenAI chat-completions
request fields used by PandaWiki: `model`, `messages`, and `stream`.

## Basic call: one knowledge base per token

Enable **问答机器人 API** in the target knowledge base's administrator settings
and create its API Token. A token that is authorized for exactly one knowledge
base needs no PandaWiki-specific request header.

```bash
curl http://<server-ip>:2444/share/v1/chat/completions \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "knowledge-base",
    "messages": [
      {"role": "user", "content": "What does this knowledge base say about the contract?"}
    ],
    "stream": false
  }'
```

For streaming output, set `"stream": true`. The response is Server-Sent
Events (`text/event-stream`) with OpenAI-compatible chunks and a final
`data: [DONE]` event.

## Advanced call: a token authorized for multiple knowledge bases

When one API Token is intentionally enabled for more than one knowledge base,
the request must select an authorized knowledge base explicitly:

```bash
curl http://<server-ip>:2444/share/v1/chat/completions \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "X-KB-ID: <AUTHORIZED_KNOWLEDGE_BASE_ID>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "knowledge-base",
    "messages": [{"role": "user", "content": "Generate a summary."}],
    "stream": true
  }'
```

`X-KB-ID` is optional for a single-knowledge-base token. If it is supplied,
PandaWiki verifies that the token is authorized for that knowledge base before
retrieval starts; it cannot be used to cross knowledge-base boundaries.

## Errors

| HTTP status | Error type | Meaning |
| --- | --- | --- |
| 401 | `unauthorized` | The Bearer token is missing, malformed, unknown, or disabled. |
| 403 | `forbidden` | `X-KB-ID` names a knowledge base not authorized for the token. |
| 400 | `invalid_request_error` | The request is malformed, has no user message, or a multi-knowledge-base token needs `X-KB-ID`. |

Keep API Tokens secret. They are credentials for access to the knowledge base
and must not be embedded in browser-delivered applications, repository files,
or ordinary application logs.
