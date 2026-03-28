# CoralHub Backend Handbook

This handbook is the public documentation surface for the repository.

It is written for engineers, reviewers, and operators who need to understand how the backend works, how to run it locally, and what has already been delivered.

CoralHub Backend is a multi-tenant choir management backend built as a pragmatic modular monolith in Go, with separate API and worker processes.

## Start Here

If you are new to the project, read in this order:

1. [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md)
2. [TECHNICAL_OVERVIEW.md](TECHNICAL_OVERVIEW.md)
3. [CONTRIBUTING.md](CONTRIBUTING.md)
4. [INFRASTRUCTURE.md](INFRASTRUCTURE.md)
5. [FEATURES.md](FEATURES.md)

## Documentation Map

### Local Development

[LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md)

Use this guide for local setup, runtime commands, environment variables, and local validation.

### Technical Overview

[TECHNICAL_OVERVIEW.md](TECHNICAL_OVERVIEW.md)

Use this guide for the system shape, module boundaries, tenant rules, and the main runtime flows.

### Contributing

[CONTRIBUTING.md](CONTRIBUTING.md)

Use this guide for the contribution workflow, coding expectations, validation steps, and change checklists.

### Architecture Decisions

[adr/README.md](adr/README.md)

Use this section for the key technical decisions and the reasoning behind them.

### Infrastructure

[INFRASTRUCTURE.md](INFRASTRUCTURE.md)

Use this guide for the production summary, CI baseline, and links to deeper operational detail.

For the deeper production operating model, continue with:

- [PRODUCTION_BLUEPRINT.md](PRODUCTION_BLUEPRINT.md)

### Features

[FEATURES.md](FEATURES.md)

Use this guide to review the delivered capabilities, detailed feature docs, and remaining follow-up work.

### Delivery History

[HISTORY.md](HISTORY.md)

Use this guide for the roadmap status, completed slices, and follow-up areas after Stage 11.
