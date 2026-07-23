# Report generation with a cloud LLM

PandaWiki report generation uses the currently configured chat model. This V1
implementation does not add a local-model provider or a separate report-model
configuration.

Before a report request is sent to the model, PandaWiki resolves the requesting
user's authorization groups and retrieves evidence only through the existing
RAG permission path. The report prompt can contain the selected report profile,
the user's permitted input values, and the authorized retrieval excerpts. Those
excerpts are then sent to the configured cloud-model provider to generate the
report.

Deployers are responsible for selecting a provider and account that are
appropriate for the sensitivity of their research, legal, financial, or other
knowledge-base material. In particular, do not configure a cloud provider for
data that policy, contract, or regulation prohibits from leaving the deployment
boundary.

Model credentials continue to use PandaWiki's existing model configuration.
They must not be placed in client-side configuration, API responses, or logs.

Report operational logs intentionally contain only non-content metadata:

- report, knowledge-base, and profile identifiers;
- terminal status and result code;
- duration, retrieval-chunk count, and citation count; and
- configured provider and model identifiers.

They do not include the report system prompt, input values, authorized excerpts,
report body, group membership, cloud-provider response body, or model API key.

The report request has a 90-second server-side generation deadline. Provider
rate limits, upstream failures, empty responses, invalid citations, and timeout
outcomes are recorded with a safe result code and message; raw provider errors
are not returned to clients or written to report logs. PandaWiki does not add a
report-specific automatic retry policy in this release. Existing provider-level
retry behavior, if configured by the current model client, remains unchanged.
