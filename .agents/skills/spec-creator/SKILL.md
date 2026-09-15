---
name: spec-creator
description: Create, update, or review engineering-grade technical specifications for a codebase, following spec-driven development (SDD) conventions. Use whenever the user asks for a spec, requirements or design doc, feature specification, API contract, data model, or wants a feature planned on paper before implementation — including phrases like "write a spec for X", "draft the spec first", "update the specs", or any spec-first workflow (e.g. with Ralph). Produces implementation-ready, testable specs using the bundled SPEC_TEMPLATE.md.
---

# Spec Authoring

## Purpose

Produce detailed, engineering-grade specifications similar in depth and structure to a full platform spec. The output must be explicit, testable, and implementation-ready without needing back-and-forth clarification. Specs state intended behavior — what the system must do and why — and leave how to implementation.

## Core Principles

1. State intent clearly and early.
2. Define scope and non-scope to avoid ambiguity.
3. Make every requirement verifiable.
4. Prefer concrete examples over vague descriptions.
5. Use consistent naming and types across the document.
6. Include failure modes and edge cases.
7. Avoid assumptions; write them down.
8. Define canonical data models in a single spec and reference them elsewhere to prevent duplication.
9. Describe intended, observable behavior: what the system must do, not how it does it.
10. Keep implementation details out of the spec: no algorithms, file/package layouts, or internal design choices.
11. Never reference source code: no file paths, no function, class, or module names, no code snippets from the codebase.
12. Give the why: justify decisions and constraints so implementers understand the intent behind them.

## Required Structure

See the [SPEC_TEMPLATE.md](./SPEC_TEMPLATE.md) for a template with detailed section descriptions and writing guidelines.

Follow this structure exactly, but customize content to the project.

## Writing Rules

- Use precise, plain english language.
- Prefer tables for enums, options, or matrices.
- Describe requirements as intended behavior: given X, the system must Y — including error and edge cases.
- Write what and why, never how: justify decisions and constraints; leave algorithms, file/package organization, and library choices to implementation.
- Never reference source code artifacts: no file paths, no function, class, or module names, no code snippets.
- Include code blocks only to illustrate examples (inputs, outputs, commands, data formats); never source code.
- Keep naming consistent across entities, APIs, and SDKs.
- If a behavior depends on policy, state it explicitly.
- If a data model is shared across domains, treat one spec as the source of truth and reference it rather than re-defining it.

## Output Quality Checklist

Before finalizing, confirm:

- Scope is explicit and non-goals are listed.
- All core entities and relationships are defined.
- Every endpoint has auth and payloads.
- All critical workflows are described step-by-step.
- Security and permissions are not implicit.
- At least one example is given for each major section.
- Every requirement reads as intended, observable behavior (what and why).
- No implementation details and no source-code references remain.

## Related docs

When the new spec is complete, update specs/README.md to include it in the list of available specs.
