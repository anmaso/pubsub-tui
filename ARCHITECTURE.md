# Pub/Sub TUI - Architecture Documentation

## Table of Contents

1. [Overview](#overview)
2. [Architectural Pattern: The Elm Architecture (MVU)](#architectural-pattern-the-elm-architecture-mvu)
3. [Technology Stack](#technology-stack)
4. [Project Structure](#project-structure)
5. [Component Architecture](#component-architecture)
6. [Key Design Patterns](#key-design-patterns)
7. [Message Passing & Communication](#message-passing--communication)
8. [State Management](#state-management)
9. [Data Flow](#data-flow)
10. [Key Design Decisions](#key-design-decisions)
11. [Common TUI Concepts](#common-tui-concepts)
12. [Integration with GCP](#integration-with-gcp)

---

## Overview

This is a Terminal User Interface (TUI) application for managing Google Cloud Pub/Sub. It provides an interactive, keyboard-driven interface for:
- Browsing topics and subscriptions
- Publishing messages with template substitution
- Subscribing to and receiving messages in real-time
- Managing acknowledgments

**Core Philosophy**: The application follows functional programming principles with unidirectional data flow, making it predictable, testable, and maintainable.

---

## Architectural Pattern: The Elm Architecture (MVU)

The application is built using **The Elm Architecture** (also called Model-View-Update or MVU), popularized by the Elm programming language and implemented in Go via the [BubbleTea](https://github.com/charmbracelet/bubbletea) framework.

### The Three Core Components

```
┌─────────────────────────────────────────────────┐
│                                                 │
│           The Elm Architecture                  │
│                                                 │
│   ┌──────────┐         ┌──────────┐            │
│   │          │         │          │            │
│   │  Model   │────────▶│   View   │            │
│   │ (State)  │         │ (Render) │            │
│   │          │         │          │            │
│   └────▲─────┘         └──────────┘            │
│        │                                        │
│        │                                        │
│   ┌────┴─────┐                                 │
│   │          │                                  │
│   │  Update  │◀─── Messages (Events)           │
│   │ (Logic)  │                                  │
│   │          │                                  │
│   └──────────┘                                  │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 1. **Model** (State)

The Model holds all application state. It's a Go struct containing:
- Current UI state (focused panel, dimensions)
- Domain data (topics, subscriptions, messages)
- Child component models
- Active connections and contexts

```go
type Model struct {
    // Pub/Sub client
    client    *pubsub.Client
    projectID string
    
    // Child components (each has its own Model)
    topics        topics.Model
    subscriptions subscriptions.Model
    publisher     publisher.Model
    subscriber    subscriber.Model
    activity      activity.Model
    
    // Application state
    focus    FocusPanel
    width    int
    height   int
    
    // Domain state
    selectedTopic        string
    selectedSubscription string
    activeSubscription   *pubsub.Subscription
}
```

**Key Principle**: The model is immutable from the outside. All changes happen through the Update function.

### 2. **View** (Rendering)

The View is a pure function that transforms the Model into a string representation:

```go
func (m Model) View() string {
    // Pure function: Same input (model) = Same output (string)
    // No side effects, no state mutation
    
    if !m.ready {
        return "Initializing..."
    }
    
    // Build UI from model state
    leftPanel := lipgloss.JoinVertical(
        lipgloss.Left,
        m.topics.View(),
        m.subscriptions.View(),
        m.activity.View(),
    )
    
    rightPanel := lipgloss.JoinVertical(
        lipgloss.Left,
        m.publisher.View(),
        m.subscriber.View(),
    )
    
    mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
    footer := m.renderFooter()
    
    return lipgloss.JoinVertical(lipgloss.Left, mainContent, footer)
}
```

**Key Principle**: View functions are pure and side-effect-free. They simply render the current state.

### 3. **Update** (State Transitions)

The Update function is the only place where state changes occur. It receives the current model and a message, and returns a new model and optional commands:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    
    case tea.KeyMsg:
        // Handle keyboard input
        if key.Matches(msg, keys.Quit) {
            m.stopSubscription()
            return m, tea.Quit
        }
        
    case common.TopicSelectedMsg:
        // Update state based on event
        m.selectedTopic = msg.TopicName
        m.topics.SetSelectedTopic(msg.TopicName)
        m.publisher.SetTargetTopic(msg.TopicName)
        
        // Return command to log the action
        return m, func() tea.Msg {
            return common.Info("Selected topic: " + msg.TopicName)
        }
        
    case common.TopicsLoadedMsg:
        // Update model with loaded data
        m.topics, _ = m.topics.Update(msg)
        return m, nil
    }
    
    return m, nil
}
```

**Key Principle**: Update is the single source of truth for state transitions. It's deterministic and testable.

### Messages (Events)

Messages are the mechanism for triggering state changes. They can come from:
- **User input** (keyboard, mouse)
- **System events** (window resize, timer)
- **Async operations** (API responses, subscription messages)
- **Inter-component communication** (child to parent, parent to child)

```go
// Message types are just Go structs
type TopicSelectedMsg struct {
    TopicName string
    TopicFull string
}

type SubscriptionsLoadedMsg struct {
    Subscriptions []SubscriptionData
    Err           error
}
```

### Commands (Side Effects)

Commands are functions that perform side effects and return messages. They're how the architecture handles asynchronous operations:

```go
// Command to load topics from GCP
func (m Model) loadTopics() tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        topicsList, err := m.client.ListTopics(ctx)
        
        if err != nil {
            return common.TopicsLoadedMsg{Err: err}
        }
        
        return common.TopicsLoadedMsg{Topics: topicsList}
    }
}
```

**The Flow**:
1. Update returns a command
2. BubbleTea runtime executes the command in a goroutine
3. Command performs side effect (API call, file I/O, etc.)
4. Command returns a message
5. Message is fed back into Update
6. Cycle continues

---

## Technology Stack

### Core Framework: BubbleTea

[BubbleTea](https://github.com/charmbracelet/bubbletea) is a Go framework for building terminal UIs based on The Elm Architecture. It provides:
- Event loop and message dispatch
- Terminal rendering and input handling
- Alt-screen support for full-screen TUIs
- Mouse and keyboard event handling

### UI Components: Bubbles

[Bubbles](https://github.com/charmbracelet/bubbles) provides pre-built, composable components:
- **List**: Scrollable lists with filtering and selection
- **TextInput**: Single-line text input with validation
- **Viewport**: Scrollable content view
- **Spinner**: Loading indicators

### Styling: Lipgloss

[Lipgloss](https://github.com/charmbracelet/lipgloss) is a style framework for terminal UI:
- CSS-like styling (colors, borders, padding, margins)
- Layout primitives (vertical/horizontal joining)
- Responsive styling
- ANSI color support

Example:
```go
var TitleStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("62")).
    Bold(true).
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("62")).
    Padding(0, 1)
```

### Business Logic: Google Cloud Pub/Sub SDK

Official Google Cloud SDK for Go:
- Client initialization and authentication
- Topic/subscription management
- Message publishing
- Message streaming (pull subscriptions)

---

## Project Structure

```
pubsub-tui/
├── main.go                        # Application entry point
│
├── internal/
│   ├── app/                       # Main application coordinator
│   │   ├── app.go                 # Model definition and initialization
│   │   ├── update.go              # Central state update logic
│   │   └── view.go                # Layout and rendering
│   │
│   ├── components/                # UI components (each follows MVU)
│   │   ├── topics/                # Topics panel
│   │   │   ├── model.go           # State and initialization
│   │   │   ├── update.go          # Event handling
│   │   │   └── view.go            # Rendering
│   │   │
│   │   ├── subscriptions/         # Subscriptions panel
│   │   ├── publisher/             # Publisher panel
│   │   │   └── substitution.go    # Variable substitution logic
│   │   ├── subscriber/            # Subscriber panel
│   │   ├── activity/              # Activity log panel
│   │   │
│   │   └── common/                # Shared code
│   │       ├── messages.go        # Message types for communication
│   │       └── styles.go          # Shared UI styles
│   │
│   ├── pubsub/                    # GCP Pub/Sub wrapper
│   │   ├── client.go              # Client initialization
│   │   ├── auth.go                # Authentication verification
│   │   ├── topics.go              # Topic CRUD operations
│   │   ├── subscriptions.go       # Subscription CRUD operations
│   │   ├── publisher.go           # Publishing logic
│   │   └── subscriber.go          # Subscription streaming
│   │
│   └── utils/                     # Utility functions
│       ├── regex.go               # Regex filtering
│       ├── json.go                # JSON formatting
│       └── file.go                # File operations
│
└── testdata/                      # Sample message templates
    ├── sample-message.json
    └── order-event.json
```

### Architecture Layers

```
┌─────────────────────────────────────────────┐
│           main.go (Entry Point)             │
│  - Verify credentials                       │
│  - Initialize GCP client                    │
│  - Create BubbleTea program                 │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│         internal/app (Coordinator)          │
│  - Main application model                   │
│  - Coordinate child components              │
│  - Handle global events                     │
│  - Manage layout                            │
└─────────────────┬───────────────────────────┘
                  │
        ┌─────────┼─────────┐
        │         │         │
┌───────▼──┐  ┌───▼────┐  ┌▼─────────┐
│Components│  │pubsub  │  │utils     │
│UI panels │  │Business│  │Helpers   │
│MVU pattern│  │Logic   │  │          │
└──────────┘  └────────┘  └──────────┘
```

---

## Component Architecture

### Hierarchical Component Structure

The application uses a hierarchical component structure where:
- **Parent (App)** coordinates child components
- **Children (Panels)** manage their own state
- **Communication** happens via messages

```
                    ┌──────────────────┐
                    │   App Model      │
                    │  (Parent/Root)   │
                    └────────┬─────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────▼─────┐      ┌──────▼──────┐    ┌──────▼──────┐
    │  Topics  │      │  Publisher  │    │ Subscriber  │
    │  Model   │      │   Model     │    │   Model     │
    └──────────┘      └─────────────┘    └─────────────┘
         │
    ┌────▼──────────┐
    │ Subscriptions │
    │    Model      │
    └───────────────┘
         │
    ┌────▼─────┐
    │ Activity │
    │  Model   │
    └──────────┘
```

### Component Pattern (All panels follow this)

Each component follows the same structure:

```go
// 1. Model Definition (model.go)
type Model struct {
    // Bubbles components
    list        list.Model
    filterInput textinput.Model
    
    // Domain data
    allTopics   []common.TopicData
    
    // UI state
    width       int
    height      int
    focused     bool
    mode        Mode  // e.g., Normal, Filter, Create
    
    // Status
    loading     bool
    loadError   error
}

// 2. Constructor
func New() Model {
    return Model{
        list:    list.New(...),
        loading: true,
        mode:    ModeNormal,
    }
}

// 3. Update Function (update.go)
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle input
    case SomeDataMsg:
        // Update with new data
    }
    return m, nil
}

// 4. View Function (view.go)
func (m Model) View() string {
    // Render based on current state
    if m.loading {
        return "Loading..."
    }
    
    return m.renderContent()
}

// 5. Public API for parent coordination
func (m *Model) SetFocused(focused bool)
func (m *Model) SetSize(width, height int)
func (m Model) IsInputActive() bool
```

### Component Modes

Components use modes to manage different interaction states:

```go
type Mode int

const (
    ModeNormal        Mode = iota  // Normal navigation
    ModeFilter                     // Filtering/searching
    ModeCreate                     // Creating new resource
    ModeConfirmDelete              // Confirming deletion
)
```

**Example**: Topics Panel Modes
- **Normal**: Navigate list with arrow keys, select with Enter
- **Filter**: Type regex to filter topics
- **Create**: Enter name for new topic
- **ConfirmDelete**: Confirm topic deletion

Mode changes affect:
- Which input receives focus
- What keyboard shortcuts are active
- What's rendered on screen

---

## Key Design Patterns

### 1. **Unidirectional Data Flow**

Data always flows in one direction: User Action → Message → Update → Model → View

```
User presses Enter on a topic
          ↓
KeyMsg event
          ↓
Update function handles it
          ↓
Creates TopicSelectedMsg
          ↓
Update processes TopicSelectedMsg
          ↓
Updates model.selectedTopic
          ↓
View renders new state
          ↓
UI updates
```

**Benefits**:
- Predictable state changes
- Easy to debug (trace message flow)
- Testable (given model + message = new model)

### 2. **Message Passing for Communication**

Components don't directly call each other. They communicate via messages:

```go
// Topics panel wants to notify about selection
// It returns a message, not calling other components directly
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    case tea.KeyMsg:
        if key.Matches(msg, key.Enter) {
            // Don't call m.publisher.SetTopic() directly
            // Instead, return a message
            return m, func() tea.Msg {
                return common.TopicSelectedMsg{
                    TopicName: selectedTopic,
                }
            }
        }
}

// Parent app handles the message and coordinates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case common.TopicSelectedMsg:
        // Coordinate multiple components
        m.topics.SetSelectedTopic(msg.TopicName)
        m.subscriptions.SetTopicFilter(msg.TopicName)
        m.publisher.SetTargetTopic(msg.TopicName)
        return m, nil
}
```

**Benefits**:
- Loose coupling between components
- Parent has full control over coordination
- Easy to add new behaviors (just handle the message differently)

### 3. **Command Pattern for Side Effects**

All side effects (API calls, file I/O, timers) are wrapped in commands:

```go
// Command to publish a message
func (m *Model) publishMessage(topic string, content []byte) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        result := m.client.Publish(ctx, topic, content, nil)
        
        // Return result as message
        return publisher.PublishResultMsg{
            MessageID: result.MessageID,
            Err:       result.Error,
        }
    }
}

// In Update:
case publisher.PublishRequestMsg:
    // Execute publish command
    cmd := m.publishMessage(msg.Topic, msg.Content)
    return m, cmd
    
case publisher.PublishResultMsg:
    // Handle result when it comes back
    if msg.Err == nil {
        return m, showSuccess("Published: " + msg.MessageID)
    } else {
        return m, showError("Failed: " + msg.Err.Error())
    }
```

**Benefits**:
- Async operations handled cleanly
- No blocking the UI thread
- Easy to test (mock commands)

### 4. **Composition over Inheritance**

Components are composed, not inherited:

```go
type Model struct {
    // Compose Bubbles components
    list        list.Model      // Not extending, but containing
    filterInput textinput.Model
    
    // Compose our own components
    topics      topics.Model
    publisher   publisher.Model
}

// Delegate to composed components
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Delegate to child component
    m.topics, cmd = m.topics.Update(msg)
    return m, cmd
}
```

### 5. **Separation of Concerns**

Clear separation between:
- **UI Layer** (components/): Only UI concerns, no business logic
- **Business Layer** (pubsub/): GCP API interactions, no UI concerns
- **Utility Layer** (utils/): Pure functions, no side effects

```go
// ❌ Bad: Business logic in UI component
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    // Don't do GCP API calls directly in components
    topics, err := client.Topics(ctx).All()
}

// ✅ Good: UI coordinates, business layer handles logic
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    // Component returns a command
    return m, m.loadTopics()
}

func (m Model) loadTopics() tea.Cmd {
    return func() tea.Msg {
        // Business layer handles the actual call
        topics, err := m.client.ListTopics(ctx)
        return TopicsLoadedMsg{Topics: topics, Err: err}
    }
}
```

### 6. **State Machine Pattern**

Components are essentially state machines:

```
Topics Panel States:

    ┌─────────────────┐
    │   ModeNormal    │ ◄─── Default state
    │ (Navigate list) │
    └────┬───┬───┬────┘
         │   │   │
    "/"  │   │   │  "n"
         │   │   │
    ┌────▼───┐   └────▼────────┐
    │ModeFilter│  │ModeCreate   │
    │(Type regex) │(Enter name) │
    └────┬───────┘ └─────┬──────┘
         │               │
    ESC  │          ESC  │
         │               │
    ┌────▼───────────────▼─────┐
    │      ModeNormal           │
    └──────────────────────────┘
```

### 7. **Pub/Sub Pattern for Messages**

The message system is essentially a pub/sub pattern:
- Components "publish" messages
- Parent "subscribes" by handling them in Update
- Multiple handlers can respond to same message

```go
// Publisher component publishes a message
return m, func() tea.Msg {
    return publisher.PublishRequestMsg{...}
}

// Multiple components can "subscribe" (handle) it
case publisher.PublishRequestMsg:
    // 1. Activity log subscribes to log it
    m.activity.Log("Publishing message...")
    
    // 2. Publisher subscribes to show status
    m.publisher.SetStatus("Publishing...")
    
    // 3. App subscribes to execute it
    return m, m.publishMessage(msg.Topic, msg.Content)
```

---

## Message Passing & Communication

### Message Types

Messages are defined in `internal/components/common/messages.go`:

```go
// Selection Messages
type TopicSelectedMsg struct {
    TopicName string
    TopicFull string
}

type SubscriptionSelectedMsg struct {
    SubscriptionName string
    TopicName        string
}

// Data Loading Messages
type TopicsLoadedMsg struct {
    Topics []TopicData
    Err    error
}

type SubscriptionsLoadedMsg struct {
    Subscriptions []SubscriptionData
    Err           error
}

// CRUD Messages
type TopicCreatedMsg struct {
    TopicName string
    Err       error
}

type TopicDeletedMsg struct {
    TopicName string
    Err       error
}

// Logging Messages
type LogMsg struct {
    Level   LogLevel
    Message string
    Time    time.Time
}
```

### Communication Patterns

#### Pattern 1: Child → Parent (Request)

Child component wants parent to do something:

```go
// In topics/update.go
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
    case tea.KeyMsg:
        if key.Matches(msg, key.Enter) {
            topic := m.SelectedTopic()
            
            // Return message to parent
            return m, func() tea.Msg {
                return common.TopicSelectedMsg{
                    TopicName: topic.Name,
                }
            }
        }
}
```

#### Pattern 2: Parent → Children (Broadcast)

Parent receives message and coordinates children:

```go
// In app/update.go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case common.TopicSelectedMsg:
        // Update multiple children
        m.selectedTopic = msg.TopicName
        m.topics.SetSelectedTopic(msg.TopicName)          // Child 1
        m.subscriptions.SetTopicFilter(msg.TopicName)     // Child 2
        m.publisher.SetTargetTopic(msg.TopicName)         // Child 3
        
        // Log the action
        return m, func() tea.Msg {
            return common.Info("Selected topic: " + msg.TopicName)
        }
}
```

#### Pattern 3: Side Effect → Update (Async Result)

Command executes asynchronously and returns result:

```go
// Command is issued
case subscriptions.CreateSubscriptionMsg:
    cmd := m.createSubscription(msg.SubscriptionName, msg.TopicName)
    return m, cmd

// Command implementation
func (m *Model) createSubscription(name, topic string) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        err := m.client.CreateSubscription(ctx, name, topic)
        
        // Result comes back as message
        return common.SubscriptionCreatedMsg{
            SubscriptionName: name,
            Err:              err,
        }
    }
}

// Result is handled
case common.SubscriptionCreatedMsg:
    if msg.Err == nil {
        // Success: refresh list
        return m, m.loadSubscriptions()
    } else {
        // Error: show error
        return m, showError(msg.Err)
    }
```

#### Pattern 4: Continuous Stream (Polling)

For ongoing message subscription:

```go
// Start subscription
func (m *Model) startSubscription(subName string) tea.Cmd {
    m.activeSubscription = m.client.Subscribe(subName)
    m.activeSubscription.Start(ctx)
    
    // Return polling command
    return m.pollMessages()
}

// Polling command
func (m *Model) pollMessages() tea.Cmd {
    return func() tea.Msg {
        select {
        case msg := <-m.activeSubscription.Messages():
            return subscriber.MessageReceivedMsg{Message: msg}
        case err := <-m.activeSubscription.Errors():
            return subscriber.SubscriptionErrorMsg{Error: err}
        }
    }
}

// Handle message and continue polling
case subscriber.MessageReceivedMsg:
    m.subscriber, cmd = m.subscriber.Update(msg)
    
    // Continue polling
    if m.activeSubscription != nil {
        return m, tea.Batch(cmd, m.pollMessages())
    }
    return m, cmd
```

### Message Flow Example: Publishing a Message

```
1. User presses Enter in Publisher panel
   ↓
2. Publisher.Update receives KeyMsg
   ↓
3. Publisher returns PublishRequestMsg
   ↓
4. App.Update receives PublishRequestMsg
   ↓
5. App creates publishMessage command
   ↓
6. Command executes in goroutine (API call to GCP)
   ↓
7. Command returns PublishResultMsg
   ↓
8. App.Update receives PublishResultMsg
   ↓
9. App forwards to Publisher.Update
   ↓
10. Publisher updates its status
    ↓
11. App creates LogMsg command
    ↓
12. Activity.Update receives LogMsg
    ↓
13. Activity adds log entry
    ↓
14. View renders new state
```

---

## State Management

### State Ownership

**Single Source of Truth**: Each piece of state has one owner.

```go
// App owns:
- selectedTopic        // Which topic is selected
- selectedSubscription // Which subscription is active
- activeSubscription   // The subscription connection
- focus                // Which panel is focused

// Topics component owns:
- allTopics            // List of all topics
- filterText           // Current filter
- mode                 // Current interaction mode
- loading              // Loading state

// Publisher component owns:
- files                // Available JSON files
- selectedFile         // Currently selected file
- variables            // Variable substitutions
- publishStatus        // Result of last publish
```

### State Synchronization

**Parent Coordinates**: When state needs to be shared, parent coordinates:

```go
// When topic is selected:
case common.TopicSelectedMsg:
    // Parent updates its own state
    m.selectedTopic = msg.TopicName
    
    // Parent synchronizes children
    m.topics.SetSelectedTopic(msg.TopicName)
    m.subscriptions.SetTopicFilter(msg.TopicName)
    m.publisher.SetTargetTopic(msg.TopicName)
```

**No Direct Communication**: Children never directly modify each other's state.

### Derived State

Some state is derived from other state:

```go
// In topics panel
func (m *Model) applyFilter() {
    var items []list.Item
    
    for _, topic := range m.allTopics {  // Source state
        if matchesFilter(topic.Name, m.filterText) {
            items = append(items, TopicItem{...})
        }
    }
    
    m.list.SetItems(items)  // Derived state
}
```

### Transient State

Some state is temporary and reset:

```go
func (m *Model) SetFocused(focused bool) {
    m.focused = focused
    
    if !focused {
        // Reset transient state when losing focus
        m.mode = ModeNormal
        m.filterInput.Blur()
        m.createInput.Blur()
    }
}
```

---

## Data Flow

### Startup Flow

```
1. main.go
   ├─ Verify GCP credentials
   ├─ Get project ID
   ├─ Create Pub/Sub client
   └─ Create BubbleTea program
           ↓
2. app.New()
   ├─ Initialize all child components
   └─ Set initial focus
           ↓
3. app.Init()
   ├─ Load topics (async command)
   ├─ Load subscriptions (async command)
   └─ Load JSON files (async command)
           ↓
4. Commands execute
   ├─ TopicsLoadedMsg returned
   ├─ SubscriptionsLoadedMsg returned
   └─ FilesLoadedMsg returned
           ↓
5. Update processes messages
   ├─ Forward to relevant components
   └─ Log to activity panel
           ↓
6. View renders initial state
```

### Topic Selection Flow

```
User navigates topics list and presses Enter
                    ↓
        topics.Update(KeyMsg)
                    ↓
        Returns TopicSelectedMsg
                    ↓
        app.Update(TopicSelectedMsg)
                    ↓
    ┌───────────────┼───────────────┐
    │               │               │
    ▼               ▼               ▼
topics.Set    subscriptions    publisher.Set
SelectedTopic .SetTopicFilter TargetTopic
    │               │               │
    └───────────────┼───────────────┘
                    ↓
            Returns LogMsg
                    ↓
        activity.Update(LogMsg)
                    ↓
              View renders
```

### Message Publishing Flow

```
1. User selects file and enters variables
                    ↓
2. User presses Enter
                    ↓
3. publisher.Update(KeyMsg)
        ↓
4. Reads file content
        ↓
5. Applies variable substitution
        ↓
6. Returns PublishRequestMsg
        ↓
7. app.Update(PublishRequestMsg)
        ↓
8. Creates publishMessage command
        ↓
9. Command calls GCP API (async)
        ↓
10. Returns PublishResultMsg
        ↓
11. app.Update(PublishResultMsg)
        ↓
12. Updates publisher status
        ↓
13. Logs to activity
        ↓
14. View renders result
```

### Message Subscription Flow

```
1. User selects subscription and presses Enter
                    ↓
2. Returns SubscriptionSelectedMsg
                    ↓
3. app.Update stops any previous subscription
                    ↓
4. app.startSubscription()
        ↓
5. Creates subscription stream
        ↓
6. Returns pollMessages command
        ↓
        ┌──────────────┐
        │              │
7. ◄───┘pollMessages()│
        │              │
        ▼              │
8. Waits for message  │
        │              │
        ▼              │
9. MessageReceivedMsg │
        │              │
        ▼              │
10. subscriber.Update │
        │              │
        ▼              │
11. Displays message  │
        │              │
        ▼              │
12. View renders      │
        │              │
        └──────────────┘
   (Loop continues until subscription stops)
```

---

## Key Design Decisions

### Decision 1: Why The Elm Architecture?

**Alternatives Considered**:
- Imperative UI (like traditional GUI frameworks)
- Object-oriented component model
- React-style virtual DOM

**Chosen**: The Elm Architecture (MVU)

**Rationale**:
- ✅ **Predictable**: Unidirectional data flow makes state changes obvious
- ✅ **Testable**: Pure functions are easy to unit test
- ✅ **Debuggable**: Can trace every state change through messages
- ✅ **Maintainable**: Clear separation of concerns
- ✅ **No race conditions**: Single-threaded update loop
- ✅ **Composable**: Components naturally nest

**Trade-offs**:
- ❌ More boilerplate (message types, command functions)
- ❌ Steeper learning curve for those unfamiliar with functional patterns
- ✅ But: Worth it for long-term maintainability

### Decision 2: Hierarchical Components vs Flat Structure

**Chosen**: Hierarchical (Parent app coordinates children)

**Rationale**:
- ✅ **Clear ownership**: Each component owns its state
- ✅ **Easier to reason about**: Top-down flow
- ✅ **Reusable**: Components can be used independently
- ✅ **Scalable**: Easy to add new panels

**Alternative** (Flat/Global State):
- All components read from global state store
- More like Redux pattern
- ❌ Rejected: Overkill for this app size

### Decision 3: Message Passing vs Direct Calls

**Chosen**: Message passing

**Rationale**:
- ✅ **Loose coupling**: Components don't know about each other
- ✅ **Flexible**: Easy to change how messages are handled
- ✅ **Loggable**: Can log all messages for debugging
- ✅ **Fits MVU**: Natural with command/message pattern

**Alternative** (Direct calls):
```go
// ❌ Rejected approach
m.publisher.SetTopic(topic)
m.subscriptions.FilterByTopic(topic)
```
- ❌ Tight coupling
- ❌ Harder to test
- ❌ Doesn't fit MVU pattern

### Decision 4: Single Subscription vs Multiple

**Chosen**: Only one active subscription at a time

**Rationale**:
- ✅ **Simpler state management**: No need to track multiple streams
- ✅ **Better UX**: Focus on one message stream
- ✅ **Resource efficient**: Don't keep unused connections
- ✅ **Matches UI**: Only one subscriber panel

**Alternative** (Multiple subscriptions):
- Support tabs or splits
- ❌ More complex UI
- ❌ More complex state management
- ❌ Overkill for target use case

### Decision 5: In-Memory vs Persistent Storage

**Chosen**: All state in-memory, no persistence

**Rationale**:
- ✅ **Simpler**: No database, no file I/O for state
- ✅ **Faster**: No disk access
- ✅ **Secure**: No sensitive data on disk
- ✅ **Stateless sessions**: Fresh start each time
- ✅ **Matches use case**: Quick debugging/testing tool

**Trade-offs**:
- ❌ Messages lost on exit
- ❌ No session history
- ✅ But: Not needed for primary use case

### Decision 6: Blocking vs Async Operations

**Chosen**: All I/O is async via commands

**Rationale**:
- ✅ **Responsive UI**: Never blocks the render loop
- ✅ **Better UX**: Can show loading states
- ✅ **Fits MVU**: Command pattern naturally handles async
- ✅ **Cancellable**: Can cancel operations (e.g., stop subscription)

**Implementation**:
- Commands run in goroutines
- Results return as messages
- UI updates when results arrive

### Decision 7: Regex for Filtering

**Chosen**: Use regex for all filtering

**Rationale**:
- ✅ **Powerful**: Can do complex filters
- ✅ **Familiar**: Developers know regex
- ✅ **Consistent**: Same pattern across all panels
- ✅ **Flexible**: Simple strings work too (as literal regex)

**Alternative** (Simple string matching):
- ❌ Less powerful
- ✅ But: Easier for non-technical users

**Compromise**: Show regex errors and fallback to showing all items

### Decision 8: Variable Substitution Format

**Chosen**: `${variableName}` syntax with `key=value` input

**Rationale**:
- ✅ **Familiar**: Same as shell/template languages
- ✅ **Simple**: Easy to parse
- ✅ **Visual**: Clear what will be substituted
- ✅ **Flexible**: Can have multiple variables

**Alternative** (JSON templates):
```json
{
  "variables": ["user", "env"],
  "template": { ... }
}
```
- ❌ More complex files
- ❌ Harder to edit

---

## Common TUI Concepts

### The Event Loop

BubbleTea runs an event loop:

```
┌─────────────────────────────────────────┐
│          BubbleTea Runtime              │
│                                         │
│  ┌──────────────────────────────────┐  │
│  │ 1. Read input (keyboard, mouse,  │  │
│  │    terminal events)               │  │
│  └─────────────┬────────────────────┘  │
│                ↓                        │
│  ┌─────────────▼────────────────────┐  │
│  │ 2. Convert to Msg                │  │
│  └─────────────┬────────────────────┘  │
│                ↓                        │
│  ┌─────────────▼────────────────────┐  │
│  │ 3. Call Update(model, msg)       │  │
│  └─────────────┬────────────────────┘  │
│                ↓                        │
│  ┌─────────────▼────────────────────┐  │
│  │ 4. Execute returned Cmd (async)  │  │
│  └─────────────┬────────────────────┘  │
│                ↓                        │
│  ┌─────────────▼────────────────────┐  │
│  │ 5. Call View(model)              │  │
│  └─────────────┬────────────────────┘  │
│                ↓                        │
│  ┌─────────────▼────────────────────┐  │
│  │ 6. Render to terminal            │  │
│  └─────────────┬────────────────────┘  │
│                ↓                        │
│  └────────────────────────────────────┘
│               │
│         (Repeat)
│               │
└───────────────┼─────────────────────────┘
```

### Alt Screen Mode

The app runs in alt-screen mode:
- Takes over the entire terminal
- Original terminal content is saved
- Restored when app exits
- Like `vim`, `less`, `htop`

```go
tea.NewProgram(
    app.New(client, projectID),
    tea.WithAltScreen(),       // Enable alt-screen
    tea.WithMouseCellMotion(), // Enable mouse
)
```

### Rendering Cycle

View is called after every Update:

```
User Input → Update → View → Screen
      ↑                        │
      └────────────────────────┘
         (Loop continues)
```

**Important**: View is called frequently, so it must be fast:
- No I/O in View
- No expensive computations
- Just format the current state

### Terminal Sizing

The app is responsive to terminal size:

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.updateComponentSizes()
```

Components calculate their dimensions:
- Left panel: 1/3 width
- Right panel: 2/3 width
- Heights distributed proportionally

### Focus Management

Only one component receives input:

```go
func (m *Model) updateFocus() {
    m.topics.SetFocused(m.focus == FocusTopics)
    m.subscriptions.SetFocused(m.focus == FocusSubscriptions)
    m.publisher.SetFocused(m.focus == FocusPublisher)
    m.subscriber.SetFocused(m.focus == FocusSubscriber)
}
```

Focused component:
- Has colored border
- Receives keyboard input
- May have active cursor

---

## Integration with GCP

### Authentication

Uses Application Default Credentials (ADC):

```go
// 1. Check environment
projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

// 2. Fallback to gcloud config
if projectID == "" {
    cmd := exec.Command("gcloud", "config", "get-value", "project")
    output, _ := cmd.Output()
    projectID = string(output)
}

// 3. Create client
client, err := pubsub.NewClient(ctx, projectID)
```

### Client Wrapper

`internal/pubsub/client.go` wraps the GCP client:

```go
type Client struct {
    client    *pubsub.Client
    projectID string
}

func (c *Client) ListTopics(ctx context.Context) ([]Topic, error)
func (c *Client) CreateTopic(ctx context.Context, name string) error
func (c *Client) DeleteTopic(ctx context.Context, name string) error
func (c *Client) Publish(ctx context.Context, topic string, data []byte) error
func (c *Client) Subscribe(name string) *Subscription
```

**Benefits**:
- Abstracts GCP SDK details
- Provides app-specific API
- Easier to mock for testing

### Subscription Streaming

Real-time message subscription:

```go
type Subscription struct {
    sub      *pubsub.Subscription
    messages chan *ReceivedMessage
    errors   chan error
    cancel   context.CancelFunc
}

func (s *Subscription) Start(ctx context.Context) {
    go func() {
        s.sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
            s.messages <- &ReceivedMessage{
                ID:          msg.ID,
                Data:        msg.Data,
                PublishTime: msg.PublishTime,
                AckFunc:     msg.Ack,
            }
        })
    }()
}

func (s *Subscription) Messages() <-chan *ReceivedMessage {
    return s.messages
}
```

The app polls this channel:

```go
func (m *Model) pollMessages() tea.Cmd {
    return func() tea.Msg {
        select {
        case msg := <-m.activeSubscription.Messages():
            return subscriber.MessageReceivedMsg{Message: msg}
        }
    }
}
```

---

## Summary

### Core Principles

1. **Unidirectional Data Flow**: User Action → Message → Update → Model → View
2. **Immutable State**: Model is never mutated directly, only through Update
3. **Pure Functions**: View and utility functions have no side effects
4. **Message Passing**: Components communicate via messages, not direct calls
5. **Commands for Side Effects**: All async/I/O wrapped in commands
6. **Composition**: Components compose other components
7. **Single Source of Truth**: Each piece of state has one owner

### Why These Patterns?

- **Predictability**: Easy to understand what will happen
- **Testability**: Pure functions are easy to test
- **Debuggability**: Can trace every state change
- **Maintainability**: Clear structure and responsibilities
- **Scalability**: Easy to add new features

### Learning Path for New TUI Developers

1. **Start with MVU**: Understand Model-View-Update pattern
2. **Study Message Flow**: Trace how messages propagate
3. **Read app.Update**: See how parent coordinates children
4. **Read one component**: Understand the component pattern
5. **Trace one feature end-to-end**: e.g., topic selection
6. **Experiment**: Add a new message type or command

### Further Reading

- [BubbleTea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [The Elm Architecture](https://guide.elm-lang.org/architecture/)
- [Lipgloss Examples](https://github.com/charmbracelet/lipgloss/tree/master/examples)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)

---

## Appendix: Quick Reference

### Key Files

| File | Purpose |
|------|---------|
| `main.go` | Entry point, client setup |
| `internal/app/app.go` | Root model definition |
| `internal/app/update.go` | Central state transitions |
| `internal/app/view.go` | Layout and rendering |
| `internal/components/common/messages.go` | All message types |
| `internal/pubsub/client.go` | GCP client wrapper |

### Message Types by Purpose

| Purpose | Message Type |
|---------|--------------|
| Selection | `TopicSelectedMsg`, `SubscriptionSelectedMsg` |
| Data Loading | `TopicsLoadedMsg`, `SubscriptionsLoadedMsg` |
| CRUD Operations | `TopicCreatedMsg`, `TopicDeletedMsg`, etc. |
| Publishing | `PublishRequestMsg`, `PublishResultMsg` |
| Subscribing | `MessageReceivedMsg`, `SubscriptionErrorMsg` |
| Logging | `LogMsg` |
| UI Control | `WindowSizeMsg`, `FocusMsg` |

### Component Lifecycle

```go
// 1. Creation
model := component.New()

// 2. Initialization (if needed)
cmd := model.Init()

// 3. Event loop
for {
    // Input
    msg := <-events
    
    // Update
    model, cmd = model.Update(msg)
    
    // Render
    view := model.View()
    render(view)
}

// 4. Cleanup (if needed)
model.Cleanup()
```


