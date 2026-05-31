# Delta for Error Handling

## ADDED Requirements

### Requirement: ErrUnprocessable generic sentinel

The system MUST define `ErrUnprocessable` as an exported sentinel error in `platform/apperror` mapped to HTTP 422 Unprocessable Entity. The sentinel MUST NOT carry a domain-specific message — it serves as a generic 422 category for modules that need unprocessable semantics without owning a dedicated sentinel.

#### Scenario: ErrUnprocessable returns 422

- GIVEN `ErrUnprocessable` is returned from any layer
- WHEN `apperror.Status(err)` is called
- THEN it returns 422

#### Scenario: ErrUnprocessable has no domain message

- GIVEN `ErrUnprocessable` sentinel definition
- WHEN its `Message` field is inspected
- THEN it is empty or generic — never references roles, users, or other domain concepts
