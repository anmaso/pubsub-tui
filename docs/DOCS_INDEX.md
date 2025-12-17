# Documentation Index

Welcome! This guide will help you find the right documentation for your needs.

## 🎯 Quick Navigation

### I want to...

| Goal | Document | Time |
|------|----------|------|
| **Get started using the app** | [README.md](./README.md) | 5 min |
| **Understand TUI architecture basics** | [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md) | 15 min |
| **See visual diagrams** | [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md) | 10 min |
| **Deep dive into architecture** | [ARCHITECTURE.md](./ARCHITECTURE.md) | 45 min |
| **Understand requirements** | [REQUIREMENTS.md](./REQUIREMENTS.md) | 30 min |

## 📚 Documentation Overview

### [README.md](./README.md)
**Purpose**: Getting started guide and project overview

**Contents**:
- Installation instructions
- Setup and authentication
- Basic usage and keyboard shortcuts
- Message templates
- Quick troubleshooting
- Links to architecture docs

**Read this if**: You want to use the application or get it running.

---

### [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md)
**Purpose**: Introduction to TUI architecture for developers new to TUIs

**Contents**:
- The 10-second overview
- Visual application structure
- MVU pattern explained with complete example
- Key concepts (messages, commands, components)
- Common patterns
- Quick tips and common mistakes

**Read this if**: You're new to TUI development and want to understand how it works with a hands-on example.

**Highlights**:
- Complete walkthrough of selecting a topic (shows full message flow)
- Clear "DO" and "DON'T" examples
- Practical debugging tips

---

### [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md)
**Purpose**: Visual reference for architecture and data flow

**Contents**:
- System architecture diagrams
- MVU pattern visualized
- Component hierarchy
- Message flow diagrams (sync & async)
- Key workflows (startup, publishing, subscribing, filtering)
- State management diagrams
- GCP integration architecture

**Read this if**: You prefer visual learning or need to see how data flows through the system.

**Highlights**:
- 15+ ASCII diagrams
- Complete workflow visualizations
- State ownership and synchronization diagrams

---

### [ARCHITECTURE.md](./ARCHITECTURE.md)
**Purpose**: Complete architecture documentation and design decisions

**Contents**:
1. Overview
2. The Elm Architecture (MVU) - detailed explanation
3. Technology stack (BubbleTea, Lipgloss, Bubbles)
4. Project structure
5. Component architecture
6. Key design patterns (7 patterns explained)
7. Message passing & communication
8. State management
9. Data flow
10. Key design decisions (8 decisions with rationale)
11. Common TUI concepts
12. GCP integration

**Read this if**: You want comprehensive understanding of the architecture, design patterns, and key decisions.

**Highlights**:
- Detailed explanation of MVU pattern
- 7 design patterns with examples
- 8 key design decisions with alternatives considered
- Message passing patterns
- State management strategies

---

### [REQUIREMENTS.md](./REQUIREMENTS.md)
**Purpose**: Complete requirements specification (reverse-engineered)

**Contents**:
1. Executive Summary
2. System Architecture
3. Functional Requirements (45+ requirements)
4. Non-Functional Requirements (performance, usability, reliability)
5. Data Requirements
6. External Interface Requirements (GCP API)
7. Constraints and Assumptions
8. Future Enhancements
9. Testing Requirements
10. Acceptance Criteria

**Read this if**: You need to understand what the application does and why, or you're working on QA/testing.

**Highlights**:
- Complete feature list
- UI layout specification
- All keyboard shortcuts
- GCP API integration details
- IAM permissions required

---

## 🎓 Learning Paths

### Path 1: User → Developer (Quick Start)
1. [README.md](./README.md) - Get it running (5 min)
2. [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md) - Learn MVU basics (15 min)
3. [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md) - See the flows (10 min)
4. Pick a component and read its code
5. Trace one feature end-to-end

**Total Time**: ~45 minutes to productive

---

### Path 2: Visual Learner
1. [README.md](./README.md) - Context (5 min)
2. [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md) - All the diagrams (15 min)
3. [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md) - Concrete example (15 min)
4. [ARCHITECTURE.md](./ARCHITECTURE.md) - Deep dive on interesting sections (30 min)

**Total Time**: ~1 hour

---

### Path 3: Deep Dive (Experienced Developer)
1. [README.md](./README.md) - Quick context (5 min)
2. [ARCHITECTURE.md](./ARCHITECTURE.md) - Read thoroughly (45 min)
3. [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md) - Reference (10 min)
4. Study `internal/app/update.go` - See pattern in practice
5. Read one complete component (e.g., `internal/components/topics/`)

**Total Time**: ~1.5 hours for deep understanding

---

### Path 4: Product/QA Focus
1. [README.md](./README.md) - Basic usage (10 min)
2. [REQUIREMENTS.md](./REQUIREMENTS.md) - Complete spec (30 min)
3. [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md) - How it works (15 min)
4. Manual testing with the application

**Total Time**: ~1 hour

---

## 🔍 Find Specific Topics

### Architecture Patterns
- **MVU Pattern**: [ARCHITECTURE.md](./ARCHITECTURE.md#architectural-pattern-the-elm-architecture-mvu), [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md#the-mvu-pattern-explained-with-example)
- **Message Passing**: [ARCHITECTURE.md](./ARCHITECTURE.md#message-passing--communication), [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#message-flow-diagrams)
- **State Management**: [ARCHITECTURE.md](./ARCHITECTURE.md#state-management), [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#state-management)
- **Component Pattern**: [ARCHITECTURE.md](./ARCHITECTURE.md#component-architecture), [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md#key-concepts)

### Design Decisions
- **Why MVU?**: [ARCHITECTURE.md](./ARCHITECTURE.md#decision-1-why-the-elm-architecture)
- **Why Message Passing?**: [ARCHITECTURE.md](./ARCHITECTURE.md#decision-3-message-passing-vs-direct-calls)
- **Why Single Subscription?**: [ARCHITECTURE.md](./ARCHITECTURE.md#decision-4-single-subscription-vs-multiple)
- **All Decisions**: [ARCHITECTURE.md](./ARCHITECTURE.md#key-design-decisions)

### Workflows
- **Publishing**: [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#workflow-2-publishing-a-message)
- **Subscribing**: [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#workflow-3-receiving-messages)
- **Filtering**: [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#workflow-4-filtering)
- **Startup**: [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#workflow-1-application-startup)

### Integration
- **GCP Setup**: [README.md](./README.md#setup)
- **GCP API**: [REQUIREMENTS.md](./REQUIREMENTS.md#62-gcp-pubsub-api)
- **Authentication**: [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md#authentication-flow)
- **Client Wrapper**: [ARCHITECTURE.md](./ARCHITECTURE.md#integration-with-gcp)

### Implementation
- **Project Structure**: [ARCHITECTURE.md](./ARCHITECTURE.md#project-structure), [README.md](./README.md#project-structure)
- **Component Lifecycle**: [ARCHITECTURE.md](./ARCHITECTURE.md#component-pattern-all-panels-follow-this)
- **Adding Components**: [README.md](./README.md#adding-a-new-component), [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md#adding-a-new-component)
- **Debugging**: [README.md](./README.md#debugging), [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md#debugging)

---

## 💡 Quick Answers

### "I want to understand how selecting a topic works"
→ [ARCHITECTURE_QUICK_START.md - A Complete Example: Selecting a Topic](./ARCHITECTURE_QUICK_START.md#a-complete-example-selecting-a-topic)

### "I want to see all the diagrams"
→ [ARCHITECTURE_DIAGRAMS.md](./ARCHITECTURE_DIAGRAMS.md)

### "I want to understand the MVU pattern"
→ [ARCHITECTURE.md - The Elm Architecture](./ARCHITECTURE.md#architectural-pattern-the-elm-architecture-mvu)

### "I want to know what messages exist"
→ [ARCHITECTURE.md - Message Types](./ARCHITECTURE.md#message-types)

### "I want to add a new feature"
1. [ARCHITECTURE.md - Component Architecture](./ARCHITECTURE.md#component-architecture)
2. [ARCHITECTURE_QUICK_START.md - Adding a New Component](./ARCHITECTURE_QUICK_START.md#adding-a-new-component)
3. Study existing component: `internal/components/topics/`

### "I want to understand the technology stack"
→ [ARCHITECTURE.md - Technology Stack](./ARCHITECTURE.md#technology-stack)

### "I need to test this"
→ [REQUIREMENTS.md - Testing Requirements](./REQUIREMENTS.md#9-testing-requirements)

### "What are the keyboard shortcuts?"
→ [README.md - Usage](./README.md#usage) or press `?` in the app

---

## 📖 Reading Order Recommendations

### Scenario 1: "I'm completely new to this codebase"
```
1. README.md
   └─ Get it running, try it out

2. ARCHITECTURE_QUICK_START.md
   └─ Understand the basics with examples

3. ARCHITECTURE_DIAGRAMS.md
   └─ See the visual overview

4. Pick a feature to modify
   └─ Use ARCHITECTURE.md as reference
```

### Scenario 2: "I know TUIs, want to understand this architecture"
```
1. README.md (skim)
   └─ Quick context

2. ARCHITECTURE.md
   └─ Read sections 1-7 thoroughly

3. Read internal/app/update.go
   └─ See pattern in practice

4. Read internal/components/topics/
   └─ See component pattern
```

### Scenario 3: "I need to fix a bug"
```
1. Reproduce the bug

2. ARCHITECTURE_DIAGRAMS.md
   └─ Find the relevant workflow diagram

3. ARCHITECTURE_QUICK_START.md
   └─ Understand message flow

4. Trace the code
   └─ Use diagrams as map

5. ARCHITECTURE.md
   └─ Reference for details
```

### Scenario 4: "I need to add a feature"
```
1. REQUIREMENTS.md
   └─ See if similar feature exists

2. ARCHITECTURE.md - Component Architecture
   └─ Understand component pattern

3. ARCHITECTURE_QUICK_START.md
   └─ See how to add messages/components

4. Study similar component
   └─ e.g., internal/components/topics/

5. ARCHITECTURE_DIAGRAMS.md
   └─ Reference for message flow
```

---

## 🚀 Next Steps

After reading the docs:

1. **Run the application** - Best way to understand it
2. **Enable debug logging** - See messages flow in real-time
3. **Modify something small** - Change a color, add a log message
4. **Trace a feature** - Pick one and follow it from input to render
5. **Add a new message type** - Follow the pattern
6. **Create a new panel** - Ultimate test of understanding

---

## 📝 Document Maintenance

These docs were created on 2024-12-14. They document the current architecture and design decisions.

**Keeping docs updated**:
- Update REQUIREMENTS.md when features change
- Update ARCHITECTURE.md when patterns change
- Keep ARCHITECTURE_QUICK_START.md in sync with examples
- Update diagrams when flows change
- Update README.md for new setup steps

---

## ❓ Still Have Questions?

1. **Check the code**: Sometimes the code is clearer than docs
   - `internal/app/` - Start here
   - `internal/components/topics/` - Simple component
   - `internal/components/common/messages.go` - All messages

2. **Trace a feature**: Follow one feature from input to output
   - Add logging at each step
   - See which messages are sent
   - See how state changes

3. **Check BubbleTea docs**: https://github.com/charmbracelet/bubbletea
   - Official examples
   - API documentation
   - Community tutorials

4. **The Elm Guide**: https://guide.elm-lang.org/architecture/
   - Original MVU pattern
   - Clear explanations
   - Interactive examples

---

Happy learning! 🎉


