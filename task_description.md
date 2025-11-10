# Coding Challenge: Deluge Scenario and Virtual User Implementation

## Overview

Implement a load testing framework that orchestrates concurrent virtual users executing DSL scripts. The framework must manage user lifecycle, track execution statistics, and enforce time-based constraints.

## Core Components

### 1. SimUser - Virtual User

A `SimUser` represents a single virtual user that executes a DSL script.

**Required Type:**
```go
type SimUserStatus int

const (
    Virgin SimUserStatus = iota
    InProgress
    DoneSuccess
    DoneError
)

type SimUser struct {
    Name          string
    Status        SimUserStatus
    SleepDuration time.Duration
    // ... (private fields as needed)
}
```

**Required Functions:**
```go
func NewSimUser(name string, script ast.Node) *SimUser
func (su *SimUser) Run()
func (su *SimUser) ExecHTTPRequest(node ast.Node, args ...object.Object) object.Object
```

**Behavior:**
- Initialize with `Virgin` status
- `Run()` executes the script once using the DSL evaluator
  - Set status to `InProgress` when starting
  - Set status to `DoneSuccess` if script completes without errors
  - Set status to `DoneError` if an error occurs during evaluation
- Each `SimUser` should be reusable (can call `Run()` multiple times)
- Each `SimUser` should be able to perform HTTP requests by calling built-in function `http`. `http` args might contain `url` and `method` fields.

### 2. Scenario - Concurrent Load Orchestrator

A `Scenario` manages multiple `SimUser` instances running concurrently.

**Required Type:**
```go
type Scenario struct {
    Name               string
    TotalSimUsersCalls int64
    // ... (private fields as needed)
}
```

**Required Functions:**
```go
func NewScenario(name string, concurrent int, delay time.Duration, totalDuration time.Duration, script ast.Node) *Scenario
func (sc *Scenario) Run(duration time.Duration)
```

**Behavior:**
- `NewScenario` creates `concurrent` number of SimUsers, each with a unique name
- `Run(duration)` orchestrates concurrent execution of virtual users. They terminates if `duration` has elapsed or SimUser status is `DoneError`
- Scenarios are run through the deluge's Run method.

**Timing Requirements:**
- Enforce the `duration` parameter strictly - scenarios must terminate after approximately the specified duration
- Each SimUser should execute at regular `delay` intervals
