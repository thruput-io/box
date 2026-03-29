# Box Project ´

These guidelines apply to all work in this repository.

## Tooling
- Makefile contains target for developing and testing.
- box command is a wrapper around make but is tied to the project and can be issued from anywhere
- box self-test is the command you start and end tasks with.

## Overall Objectives
- Provide an emulated environment where code can be tested exactly as it would run in production.
- Initially allowing for configuration changes, but ultimately treat code and configuration as a single immutable unit (container).
- Making emulation so well so that new features can be tested without having to deploy the project. 
- Making it possible to run and test compositions of services / frontends / backends.
- Local first development. Run and debug code using a localhost development environment.
- Only dockerize when necessary.
- Identifying weaknesses or issues before they go to production.

## Defeats The Purpose
- Use localhost/127.0.0.1 urls instead of DNS names.
- Weaken ssl security by ignoring or disabling ssl errors
- Solve issue by using something that would not work in production.
- Using 'hardening' or 'robust' configuration or features that can hide weaknesses or issues.

## Remember!
- There is a network inside docker and a network outside docker. DNS names resolve differently.
- To verify behavior inside docker use:
  - browser: https://browser.web.com
  - mcp: https://tools.web.internal/sse
- This is a development, mock and testing environment:
  - Do not apply security measures for other reasons than emulating production. Make it supereasy to bypass.
  - Do not use persistence, restarts should always start from same clean state.
- FAIL FAST AND LOUD; it is the only way to find weaknesses and issues.
