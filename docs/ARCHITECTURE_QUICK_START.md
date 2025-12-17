# Pub/Sub TUI - Quick Start Guide for Developers

> **New to TUIs?** This guide will get you oriented quickly. Read [ARCHITECTURE.md](./ARCHITECTURE.md) for the full deep dive.

## What is This?

A Terminal User Interface (TUI) for Google Cloud Pub/Sub, built with Go and the BubbleTea framework.

**Think of it as**: A text-based GUI that runs entirely in your terminal, similar to `vim`, `htop`, or `k9s`.

## The 10-Second Overview

```
┌────────────────────────────────────────────────────────┐
│  This TUI uses The Elm Architecture (MVU Pattern)     │
│                                                        │
│     User Input → Message → Update → Model → View      │
│          ↑                                      ↓      │
│          └──────────────────────────────────────┘      │
│                                                        │
│  Everything is a message. State changes only in       │
│  Update. View is a pure function. Commands handle     │
│  side effects.                                         │
└────────────────────────────────────────────────────────┘
```

## Visual Architecture

### Application Structure

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                             │
│  • Verify GCP credentials                                   │
│  • Create Pub/Sub client                                    │
│  • Start BubbleTea program                                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                   internal/app/                             │
│                   (Root Coordinator)                        │
│                                                             │
│  • Owns the main Model                                      │
│  • Coordinates all child components                         │
│  • Handles global events (quit, focus, window size)        │
│  • Routes messages between components                       │
│  • Manages GCP client and subscriptions                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
┌───────▼──────┐  ┌───▼──────┐  ┌──▼──────────┐
│  Components  │  │  pubsub  │  │    utils    │
│  (UI Panels) │  │(GCP API) │  │  (Helpers)  │
└──────────────┘  └──────────┘  └─────────────┘
```

### UI Layout

```
Terminal Window (Full Screen)
┌──────────────────────────────────────────────────────────┐
│                                                          │
│  ┌──────────────┐  ┌──────────────────────────────────┐ │
│  │   Topics     │  │         Publisher                 │ │
│  │              │  │  • Select JSON file               │ │
│  │  • list      │  │  • Set variables                  │ │
│  │  • create    │  │  • Publish to topic               │ │
│  │  • delete    │  └──────────────────────────────────┘ │
│  │  • filter    │                                        │
│  └──────────────┘  ┌──────────────────────────────────┐ │
│                    │         Subscriber                │ │
│  ┌──────────────┐  │  • Receive messages              │ │
│  │Subscriptions │  │  • View details                  │ │
│  │              │  │  • Acknowledge                   │ │
│  │  • list      │  │  • Filter                        │ │
│  │  • create    │  └──────────────────────────────────┘ │
│  │  • delete    │                                        │
│  │  • filter    │                                        │
│  └──────────────┘                                        │
│                                                          │
│  ┌──────────────┐                                        │
│  │ Activity Log │                                        │
│  │  (read-only) │                                        │
│  └──────────────┘                                        │
│                                                          │
├──────────────────────────────────────────────────────────┤
│ Tab: cycle | 1-4: jump to panel | q: quit | /: filter   │
└──────────────────────────────────────────────────────────┘
```

## The MVU Pattern Explained (With Example)

### What is MVU?

MVU = **Model** + **View** + **Update**

It's a functional programming pattern where:
- State lives in a **Model** (struct)
- UI is rendered by a **View** function (Model → String)
- State changes happen in **Update** function (Model + Message → New Model)

### A Complete Example: Selecting a Topic

Let's trace what happens when you press Enter on a topic:

#### 1. **Input Event**

```
User presses Enter
       ↓
BubbleTea creates a KeyMsg
```

#### 2. **Update (Child Component)**

```go
// File: internal/components/topics/update.go

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    
    case tea.KeyMsg:
        if key.Matches(msg, key.Enter) {
            topic := m.SelectedTopic()
            
            // Return a message to the parent
            return m, func() tea.Msg {
                return common.TopicSelectedMsg{
                    TopicName: topic.Name,
                    TopicFull: topic.FullName,
                }
            }
        }
    }
    return m, nil
}
```

**What happened?**
- Topics component received Enter key
- It didn't directly call other components
- It returned a `TopicSelectedMsg` message
- This message goes to the parent (app)

#### 3. **Update (Parent Coordinator)**

```go
// File: internal/app/update.go

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    
    case common.TopicSelectedMsg:
        // Update our own state
        m.selectedTopic = msg.TopicName
        
        // Coordinate child components
        m.topics.SetSelectedTopic(msg.TopicName)
        m.subscriptions.SetTopicFilter(msg.TopicName)
        m.publisher.SetTargetTopic(msg.TopicName)
        
        // Log the action
        return m, func() tea.Msg {
            return common.Info("Selected topic: " + msg.TopicName)
        }
    }
    return m, nil
}
```

**What happened?**
- Parent received `TopicSelectedMsg`
- Updated its own state (`m.selectedTopic`)
- Synchronized three child components
- Returned a log message
- Log message will go to activity panel

#### 4. **View (Rendering)**

```go
// File: internal/app/view.go

func (m Model) View() string {
    // Build left panel
    leftPanel := lipgloss.JoinVertical(
        lipgloss.Left,
        m.topics.View(),        // ← Shows selected topic
        m.subscriptions.View(),  // ← Shows filtered subs
        m.activity.View(),       // ← Shows log message
    )
    
    // Build right panel
    rightPanel := lipgloss.JoinVertical(
        lipgloss.Left,
        m.publisher.View(),      // ← Shows target topic
        m.subscriber.View(),
    )
    
    // Combine and render
    return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}
```

**What happened?**
- View called on each component
- Each component renders its current state
- Results combined into final layout
- Rendered to terminal

#### 5. **The Full Flow**

```
┌─────────────────────────────────────────────────────────┐
│ 1. User presses Enter                                   │
└─────────────────┬───────────────────────────────────────┘
                  ↓
┌─────────────────▼───────────────────────────────────────┐
│ 2. topics.Update(KeyMsg)                                │
│    • Detects Enter key                                  │
│    • Returns TopicSelectedMsg                           │
└─────────────────┬───────────────────────────────────────┘
                  ↓
┌─────────────────▼───────────────────────────────────────┐
│ 3. app.Update(TopicSelectedMsg)                         │
│    • Updates m.selectedTopic                            │
│    • Calls m.topics.SetSelectedTopic()                  │
│    • Calls m.subscriptions.SetTopicFilter()             │
│    • Calls m.publisher.SetTargetTopic()                 │
│    • Returns LogMsg                                     │
└─────────────────┬───────────────────────────────────────┘
                  ↓
┌─────────────────▼───────────────────────────────────────┐
│ 4. app.Update(LogMsg)                                   │
│    • Forwards to m.activity.Update()                    │
│    • Activity adds log entry                            │
└─────────────────┬───────────────────────────────────────┘
                  ↓
┌─────────────────▼───────────────────────────────────────┐
│ 5. View renders                                         │
│    • Topics shows selected indicator                    │
│    • Subscriptions shows filtered list                  │
│    • Publisher shows target topic                       │
│    • Activity shows log message                         │
└─────────────────────────────────────────────────────────┘
```

## Key Concepts

### 1. Messages Are Everything

Messages are how everything communicates:

```go
// User input
type KeyMsg struct { ... }

// Data loaded from API
type TopicsLoadedMsg struct {
    Topics []TopicData
    Err    error
}

// User action
type TopicSelectedMsg struct {
    TopicName string
}

// Logging
type LogMsg struct {
    Level   LogLevel
    Message string
}
```

**Rule**: If something needs to happen, send a message.

### 2. Commands Handle Side Effects

Commands are functions that do async work and return messages:

```go
// Command to load topics from GCP
func (m Model) loadTopics() tea.Cmd {
    return func() tea.Msg {
        // This runs in a goroutine
        ctx := context.Background()
        topics, err := m.client.ListTopics(ctx)
        
        // Return result as message
        return TopicsLoadedMsg{
            Topics: topics,
            Err:    err,
        }
    }
}
```

**Usage**:
```go
case NeedTopicsMsg:
    // Issue the command
    return m, m.loadTopics()
    
case TopicsLoadedMsg:
    // Handle the result
    m.topics = msg.Topics
    return m, nil
```

### 3. Components Are Composable

Each component follows the same pattern:

```go
type Model struct {
    // Bubbles components (from library)
    list        list.Model
    filterInput textinput.Model
    
    // Our own state
    allTopics   []TopicData
    focused     bool
    mode        Mode
}

func New() Model { ... }
func (m Model) Init() tea.Cmd { ... }
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) { ... }
func (m Model) View() string { ... }
```

**Composing**:
```go
type AppModel struct {
    topics    topics.Model     // Compose other components
    publisher publisher.Model
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    // Delegate to child
    m.topics, cmd = m.topics.Update(msg)
    return m, cmd
}

func (m Model) View() string {
    // Render child
    return m.topics.View()
}
```

### 4. State Flows Downward, Messages Flow Upward

```
        ┌──────────┐
        │   App    │  ← Coordinates
        └─────┬────┘
              │ State flows down
      ┌───────┼───────┐
      │       │       │
  ┌───▼──┐ ┌──▼──┐ ┌─▼───┐
  │Topics│ │Subs │ │Pub  │  ← Render
  └───┬──┘ └──┬──┘ └─┬───┘
      │       │      │
      └───────┼──────┘
              │ Messages flow up
        ┌─────▼────┐
        │   App    │  ← Handles
        └──────────┘
```

## Common Patterns

### Pattern 1: Loading Data

```go
// Step 1: Issue command
case StartupMsg:
    return m, m.loadTopics()

// Step 2: Command executes async
func (m Model) loadTopics() tea.Cmd {
    return func() tea.Msg {
        topics, err := fetchTopics()
        return TopicsLoadedMsg{Topics: topics, Err: err}
    }
}

// Step 3: Handle result
case TopicsLoadedMsg:
    if msg.Err != nil {
        m.error = msg.Err
    } else {
        m.topics = msg.Topics
    }
    return m, nil
```

### Pattern 2: User Action

```go
// Detect action
case tea.KeyMsg:
    if key.Matches(msg, key.Enter) {
        // Return message for parent
        return m, func() tea.Msg {
            return TopicSelectedMsg{TopicName: m.selected}
        }
    }

// Parent handles
case TopicSelectedMsg:
    m.updateComponents(msg.TopicName)
    return m, nil
```

### Pattern 3: Continuous Stream

```go
// Start stream
return m, m.pollMessages()

// Polling command
func (m Model) pollMessages() tea.Cmd {
    return func() tea.Msg {
        msg := <-m.stream  // Block until message
        return MessageReceivedMsg{Message: msg}
    }
}

// Handle and continue
case MessageReceivedMsg:
    m.messages = append(m.messages, msg)
    return m, m.pollMessages()  // Continue polling
```

## Where to Start Reading

### For Understanding the Architecture

1. **Start**: `main.go` - See how it all starts
2. **Then**: `internal/app/app.go` - See the root Model
3. **Then**: `internal/app/update.go` - See how messages are handled
4. **Then**: `internal/components/topics/` - See a complete component
5. **Then**: `internal/components/common/messages.go` - See all messages

### For Understanding a Feature

Pick a feature and trace it end-to-end:

**Example: Topic Creation**

1. `topics/update.go` - User presses 'n', enters name, presses Enter
2. `CreateTopicMsg` sent to parent
3. `app/update.go` - Parent calls `m.createTopic(name)`
4. `app/update.go` - Command calls GCP API
5. `TopicCreatedMsg` returned
6. `app/update.go` - Forwards to `topics.Update()`
7. `topics/update.go` - Refreshes list
8. `topics/view.go` - Renders updated list

## Quick Tips

### Debugging

1. **Add logging**:
   ```go
   return m, func() tea.Msg {
       return common.Info("Debug: " + value)
   }
   ```

2. **Print messages** (to file, stdout will break TUI):
   ```go
   f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
   fmt.Fprintf(f, "Received: %#v\n", msg)
   ```

3. **Trace state changes**: Look at activity log in UI

### Adding a New Message Type

1. Define in `internal/components/common/messages.go`:
   ```go
   type MyNewMsg struct {
       Data string
   }
   ```

2. Send it:
   ```go
   return m, func() tea.Msg {
       return common.MyNewMsg{Data: "test"}
   }
   ```

3. Handle in `internal/app/update.go`:
   ```go
   case common.MyNewMsg:
       // Do something
       return m, nil
   ```

### Adding a New Component

1. Create `internal/components/mycomponent/`
2. Add `model.go`, `update.go`, `view.go`
3. Follow the pattern from `topics/`
4. Add to `app.Model`:
   ```go
   type Model struct {
       myComponent mycomponent.Model
   }
   ```
5. Initialize in `app.New()`
6. Update in `app.Update()`
7. Render in `app.View()`

## Common Mistakes

### ❌ DON'T: Call other components directly

```go
// ❌ Wrong
m.publisher.SetTopic(topic)
m.subscriptions.FilterByTopic(topic)
```

### ✅ DO: Send messages

```go
// ✅ Correct
return m, func() tea.Msg {
    return TopicSelectedMsg{TopicName: topic}
}

// Parent handles coordination
case TopicSelectedMsg:
    m.publisher.SetTopic(msg.TopicName)
    m.subscriptions.FilterByTopic(msg.TopicName)
```

### ❌ DON'T: Do I/O in Update or View

```go
// ❌ Wrong - blocks UI
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    topics, _ := m.client.ListTopics()  // Blocks!
    m.topics = topics
    return m, nil
}
```

### ✅ DO: Use commands

```go
// ✅ Correct - async
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    return m, m.loadTopics()  // Returns command
}

func (m Model) loadTopics() tea.Cmd {
    return func() tea.Msg {
        topics, _ := m.client.ListTopics()  // Runs async
        return TopicsLoadedMsg{Topics: topics}
    }
}
```

### ❌ DON'T: Mutate state outside Update

```go
// ❌ Wrong
func (m *Model) DoSomething() {
    m.state = "changed"  // State change outside Update!
}
```

### ✅ DO: Change state only in Update

```go
// ✅ Correct
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    case SomeMsg:
        m.state = "changed"  // State changes here
        return m, nil
}
```

## Architecture Principles Summary

1. **Single Source of Truth**: Each piece of state has one owner
2. **Unidirectional Data Flow**: Always User → Message → Update → Model → View
3. **Pure Functions**: View has no side effects
4. **Immutable Updates**: Update returns new state, doesn't mutate
5. **Commands for Side Effects**: All I/O wrapped in commands
6. **Message Passing**: Components communicate via messages
7. **Composition**: Build complex UIs from simple components

## Next Steps

1. **Read the full docs**: [ARCHITECTURE.md](./ARCHITECTURE.md)
2. **Study BubbleTea examples**: https://github.com/charmbracelet/bubbletea/tree/master/examples
3. **Trace a feature**: Pick one and follow it from input to render
4. **Modify something**: Change a message handler, add a log
5. **Build something**: Add a new panel or feature

## Questions?

Common questions answered in [ARCHITECTURE.md](./ARCHITECTURE.md):

- Why use MVU instead of traditional UI patterns?
- How does the event loop work?
- How is state synchronized?
- What are commands and how do they work?
- How does message passing work?
- How is the GCP client integrated?

---

**Remember**: Everything is a message. State changes only in Update. View is pure. Commands handle side effects.

Happy coding! 🚀


