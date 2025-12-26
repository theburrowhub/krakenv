<!--
  SYNC IMPACT REPORT
  ==================
  Version change: N/A → 1.0.0 (initial constitution)
  
  Added Principles:
  - I. Code Quality First
  - II. Test-Driven Development (TDD)
  - III. User Experience Consistency
  - IV. Performance Requirements
  
  Added Sections:
  - Quality Standards
  - Development Workflow
  - Governance
  
  Templates Status:
  - .specify/templates/plan-template.md ✅ compatible (Constitution Check section exists)
  - .specify/templates/spec-template.md ✅ compatible (success criteria align with UX/performance)
  - .specify/templates/tasks-template.md ✅ compatible (test-first workflow supported)
  
  Follow-up TODOs: None
-->

# Krakenv Constitution

## Core Principles

### I. Code Quality First

Code quality is the foundation of sustainable software development. All code contributions MUST adhere to the following non-negotiable standards:

- **Single Responsibility**: Each function, module, and class MUST have one clear purpose
- **Functional Programming**: Prefer pure functions and immutable data structures; side effects MUST be isolated and explicit
- **Naming Conventions**: All identifiers MUST be self-documenting in US English; no abbreviations unless universally understood
- **Code Review**: All changes MUST be reviewed before merge; no self-approved PRs for production code
- **Technical Debt**: Debt MUST be documented with `TODO(category): explanation` comments and tracked in backlog
- **DRY Principle**: No code duplication; extract shared logic into reusable functions or modules

**Rationale**: High-quality code reduces long-term maintenance costs, improves team velocity, and minimizes production incidents.

### II. Test-Driven Development (TDD)

Testing is not optional. All features MUST follow the TDD workflow:

- **Red-Green-Refactor**: Write failing tests FIRST → implement minimal code to pass → refactor with confidence
- **Test Coverage**: Critical paths MUST have ≥80% coverage; edge cases MUST be explicitly tested
- **Test Types Required**:
  - **Unit tests**: For all pure functions and isolated logic
  - **Integration tests**: For inter-module communication and external dependencies
  - **Contract tests**: For API boundaries and data schemas
- **Test Independence**: Each test MUST be runnable in isolation; no shared mutable state between tests
- **CI/CD Gate**: All tests MUST pass before merge; failed tests block deployment

**Rationale**: TDD catches defects early, documents expected behavior, and provides safety net for refactoring.

### III. User Experience Consistency

The user interface MUST deliver a cohesive, predictable, and accessible experience:

- **Design System**: All UI components MUST follow the established design system; no one-off styling
- **Accessibility**: WCAG 2.1 AA compliance is REQUIRED; semantic HTML, ARIA labels, and keyboard navigation MUST be supported
- **Responsive Design**: All interfaces MUST work across defined breakpoints (mobile, tablet, desktop)
- **Error Handling**: User-facing errors MUST be clear, actionable, and never expose internal details
- **Loading States**: All async operations MUST display appropriate feedback (spinners, skeletons, progress)
- **Feedback Loops**: User actions MUST have immediate visual confirmation (within 100ms perceived response)

**Rationale**: Consistent UX builds user trust, reduces support burden, and improves adoption rates.

### IV. Performance Requirements

Performance is a feature. All code MUST meet defined performance budgets:

- **Response Time**: API endpoints MUST respond within p95 < 200ms for standard operations
- **Time to Interactive**: Frontend MUST be interactive within 3 seconds on 3G networks
- **Bundle Size**: JavaScript bundles MUST not exceed defined limits; lazy loading REQUIRED for non-critical paths
- **Memory Efficiency**: No memory leaks; long-running processes MUST have stable memory profiles
- **Database Queries**: N+1 queries are FORBIDDEN; all queries MUST be optimized and indexed
- **Monitoring**: Performance metrics MUST be collected in production; degradation triggers alerts

**Rationale**: Poor performance directly impacts user satisfaction, conversion rates, and operational costs.

## Quality Standards

This section defines cross-cutting quality requirements that complement the core principles:

### Code Standards

- **Linting**: All code MUST pass configured linters with zero warnings
- **Formatting**: Automatic formatting MUST be applied; manual formatting debates are eliminated
- **Type Safety**: Strongly-typed languages preferred; runtime type checks for dynamic boundaries
- **Documentation**: Public APIs MUST have documentation; complex algorithms MUST have inline comments explaining "why"

### Security Standards

- **Input Validation**: All external input MUST be validated and sanitized
- **Authentication**: Sensitive operations REQUIRE proper authentication
- **Authorization**: Access control MUST follow principle of least privilege
- **Secrets Management**: No hardcoded secrets; use environment variables or secret managers

### Observability Standards

- **Structured Logging**: All logs MUST be structured (JSON) with correlation IDs
- **Metrics**: Key business and technical metrics MUST be tracked
- **Tracing**: Distributed operations MUST be traceable across service boundaries
- **Alerting**: Critical failures MUST trigger automated alerts

## Development Workflow

This section defines the required development practices:

### Version Control

- **Branch Strategy**: Feature branches from main; short-lived branches preferred
- **Commit Messages**: Follow conventional commits format (feat:, fix:, docs:, etc.)
- **Pull Requests**: MUST include description, testing evidence, and checklist completion

### Code Review Process

- **Review Scope**: All changes require at least one approval
- **Review Criteria**: Functionality, tests, performance impact, security, and adherence to constitution
- **Feedback**: Constructive, specific, and actionable; focus on code, not person

### Deployment

- **Environment Parity**: Development, staging, and production MUST be as similar as possible
- **Rollback Ready**: All deployments MUST support quick rollback
- **Feature Flags**: New features SHOULD be deployable behind feature flags

## Governance

This constitution supersedes all other development practices. Amendments and compliance follow these rules:

### Amendment Process

1. **Proposal**: Submit documented proposal with rationale
2. **Review**: Allow minimum 48 hours for team review and feedback
3. **Approval**: Requires consensus or designated authority approval
4. **Migration**: Include migration plan for existing code if breaking change
5. **Documentation**: Update version, date, and sync all dependent templates

### Versioning Policy

- **MAJOR**: Backward-incompatible changes to principles or removal of requirements
- **MINOR**: New principles, sections, or material expansions
- **PATCH**: Clarifications, typo fixes, non-semantic refinements

### Compliance Review

- All PRs MUST verify compliance with constitution principles
- Quarterly review of constitution relevance and effectiveness
- Violations MUST be documented and addressed before merge

**Version**: 1.0.0 | **Ratified**: 2025-12-26 | **Last Amended**: 2025-12-26
