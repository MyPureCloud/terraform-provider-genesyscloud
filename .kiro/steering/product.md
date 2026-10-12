# Product

This project is the **Genesys Cloud Terraform Provider** (also called **CX as Code**). It is a Terraform provider, written in Go, that wraps the Genesys Cloud Public REST APIs so that Genesys Cloud configuration (users, queues, skills, flows, routing, telephony, outbound campaigns, etc.) can be managed declaratively as infrastructure-as-code.

## What it does

- Exposes Genesys Cloud objects as Terraform **resources** (managed) and **data sources** (read-only references to existing objects).
- Performs all CRUD via Genesys Cloud Public API calls, which require an authorized OAuth client with the right permissions and scopes.
- Supports **export** (`genesyscloud_tf_export`) to generate Terraform config/state from an existing org.

## Users

Genesys Cloud administrators and platform engineers who want to version-control and automate their org configuration instead of clicking through the admin UI.

## Key facts

- Provider source (published): `mypurecloud/genesyscloud`. Local/dev source: `genesys.com/mypurecloud/genesyscloud`.
- Auth via `GENESYSCLOUD_OAUTHCLIENT_ID` / `_SECRET` / `_REGION` env vars (or an access token).
- The public API surface is reached through the `platformclientv2` SDK (`github.com/mypurecloud/platform-client-sdk-go`).

See `README.md` for full usage, and `DICTIONARY.md` for domain vocabulary.
