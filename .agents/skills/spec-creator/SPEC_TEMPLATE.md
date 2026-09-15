# Spec Name

Status: Proposed | Implemented

> Write every section in terms of intended behavior: what the system must do and why, never how. No implementation details and no references to source code.

## Overview

### Purpose

- Clearly articulate the problem being solved and the rationale for the feature.
- Jobs to be done and user needs addressed.

### Goals

- Specific, measurable outcomes the spec aims to achieve.

### Non-Goals

- Clarify what is out of scope to prevent scope creep.

### Scope

- Define the boundaries of the feature, including what is included and excluded.

## High-Level Design

### Components and responsibilities

- Describe the conceptual components, their responsibilities, and how they interact.
- Name components by role; never by file, package, class, or function names.
- Optionally include an ASCII diagram of how the components interact.

### Data flow summary

- Overview of how data moves through the system for key operations.

## Data model

### Core Entities

- Define primary entities with types and fields
- Define fields with logical, language-neutral types; do not include code snippets.
- Explain relationships between entities

### Relationships

- Explain how entities relate to others features/modules/contexts.

### Persistence Notes

- Persistence requirements as a logical model: what is stored, its fields, types, and constraints.
- State durability expectations (e.g. data survives restarts) without prescribing the storage technology or SQL dialect.

## Workflows

- Step-by-step flows for critical operations
- Include both happy path and error/merge cases
- Describe intended, observable behavior: inputs or triggers, expected outcomes, and error handling.

## APIs

- Base paths
- Endpoints (method, path, purpose)
- Auth requirements
- Request/response payloads

## Client SDK Design

- Initialization patterns
- Example usage
- Behavior expectations (batching, retry, persistence, etc)

## Configuration

- Environment variables, extra configurations and defaults

## Permissions

- Roles and access matrix

## Security Considerations

- Sensitive data handling
- Validation rules
- Key management

## Technical Constraints

- External systems, protocols, compatibility, and performance requirements the implementation must satisfy, with rationale.

## Open Questions / Risks

## Verifications

Section with 3-5 objective checks, each stating an observable behavior rather than a test suite or tool.
Examples:

- Requests without a valid key are rejected with 401.
- Data persists across restarts.
- Invalid input produces a descriptive validation error.
- The command completes within the stated performance constraint for a typical input.

## Appendices

- Compatibility notes
- Future considerations
