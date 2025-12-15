# Pub/Sub TUI - Architecture Diagrams

> Visual reference for understanding data flow, component interactions, and key workflows.

## Table of Contents

1. [System Architecture Overview](#system-architecture-overview)
2. [The MVU Pattern Visualized](#the-mvu-pattern-visualized)
3. [Component Hierarchy](#component-hierarchy)
4. [Message Flow Diagrams](#message-flow-diagrams)
5. [Key Workflows](#key-workflows)
6. [State Management](#state-management)
7. [Integration Architecture](#integration-architecture)

---

## System Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User's Terminal                          │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                   BubbleTea Runtime                       │ │
│  │                                                           │ │
│  │  ┌─────────────────────────────────────────────────────┐ │ │
│  │  │              Application (MVU)                      │ │ │
│  │  │                                                     │ │ │
│  │  │  ┌─────────┐    ┌─────────┐    ┌─────────┐       │ │ │
│  │  │  │ Model   │───▶│  View   │───▶│ Render  │       │ │ │
│  │  │  │ (State) │    │ (String)│    │ (ANSI)  │       │ │ │
│  │  │  └────▲────┘    └─────────┘    └─────────┘       │ │ │
│  │  │       │                                           │ │ │
│  │  │  ┌────┴────────┐                                 │ │ │
│  │  │  │   Update    │◀──── Messages (Events)          │ │ │
│  │  │  │  (Logic)    │                                 │ │ │
│  │  │  └─────────────┘                                 │ │ │
│  │  │                                                     │ │ │
│  │  └─────────────────────────────────────────────────────┘ │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ API Calls
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Google Cloud Pub/Sub                          │
│  ┌──────────┐  ┌──────────┐  ┌────────────┐  ┌──────────────┐ │
│  │  Topics  │  │  Subs    │  │  Publish   │  │  Subscribe   │ │
│  └──────────┘  └──────────┘  └────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

### Layered Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Presentation Layer                        │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  internal/app/ (Coordinator)                           │  │
│  │  - Layout management                                   │  │
│  │  - Focus management                                    │  │
│  │  - Global event handling                               │  │
│  └────────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  internal/components/ (UI Panels)                      │  │
│  │  - Topics, Subscriptions, Publisher, Subscriber        │  │
│  │  - Each follows MVU pattern                            │  │
│  │  - Uses Bubbles components (list, textinput, etc.)    │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                            │
┌──────────────────────────▼────────────────────────────────────┐
│                     Business Layer                            │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  internal/pubsub/ (Domain Logic)                       │  │
│  │  - GCP client wrapper                                  │  │
│  │  - Topic/Subscription operations                       │  │
│  │  - Publisher/Subscriber logic                          │  │
│  │  - Message streaming                                   │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                            │
┌──────────────────────────▼────────────────────────────────────┐
│                     Utility Layer                             │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  internal/utils/ (Helpers)                             │  │
│  │  - Regex filtering                                     │  │
│  │  - JSON formatting                                     │  │
│  │  - File operations                                     │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
                            │
┌──────────────────────────▼────────────────────────────────────┐
│                  External Dependencies                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │  BubbleTea   │  │  Lipgloss    │  │  GCP SDK     │       │
│  │  (Framework) │  │  (Styling)   │  │  (Pub/Sub)   │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
└──────────────────────────────────────────────────────────────┘
```

---

## The MVU Pattern Visualized

### The Update Loop

```
                   ┌─────────────────────────┐
                   │   BubbleTea Runtime     │
                   │    (Event Loop)         │
                   └───────────┬─────────────┘
                               │
                               ▼
            ┌──────────────────────────────────┐
            │    1. Input Event                │
            │    - Keyboard                    │
            │    - Mouse                       │
            │    - Window resize               │
            │    - Custom messages             │
            └──────────────┬───────────────────┘
                           │
                           ▼
            ┌──────────────────────────────────┐
            │    2. Update(model, msg)         │
            │    - Pattern match on message    │
            │    - Apply business logic        │
            │    - Return new model            │
            │    - Return command (optional)   │
            └──────────────┬───────────────────┘
                           │
                           ├────────────────────┐
                           ▼                    ▼
         ┌─────────────────────────┐   ┌────────────────┐
         │    3a. Execute Command  │   │  3b. New Model │
         │    - Async operation    │   └────────┬───────┘
         │    - API call           │            │
         │    - File I/O           │            │
         │    - Timer              │            │
         └──────────┬──────────────┘            │
                    │                           │
                    │ Returns Message           │
                    ▼                           │
         ┌─────────────────────────┐            │
         │  4. New Message         │            │
         │  (Async result)         │            │
         └──────────┬──────────────┘            │
                    │                           │
                    └────────────┬──────────────┘
                                 ▼
                   ┌──────────────────────────────────┐
                   │    5. View(model)                │
                   │    - Pure function               │
                   │    - Render to string            │
                   │    - No side effects             │
                   └──────────────┬───────────────────┘
                                  ▼
                   ┌──────────────────────────────────┐
                   │    6. Render to Terminal         │
                   │    - ANSI escape codes           │
                   │    - Cursor positioning          │
                   │    - Color codes                 │
                   └──────────────────────────────────┘
                                  │
                                  │ (Loop continues)
                                  ▼
```

### State Transitions

```
                Current State
                     │
                     │ Event occurs
                     ▼
              ┌──────────────┐
              │  KeyMsg or   │
              │  CustomMsg   │
              └──────┬───────┘
                     │
                     ▼
              ┌──────────────┐
              │   Update()   │
              │              │
              │  Match msg   │
              │  Apply logic │
              └──────┬───────┘
                     │
         ┌───────────┴───────────┐
         ▼                       ▼
   ┌──────────┐          ┌──────────┐
   │ New State│          │ Command  │
   └────┬─────┘          └────┬─────┘
        │                     │
        │                     │ Async execution
        │                     ▼
        │              ┌──────────────┐
        │              │ Result Msg   │
        │              └──────┬───────┘
        │                     │
        └──────────┬──────────┘
                   ▼
             Next State
                   │
                   ▼
               View()
                   │
                   ▼
               Render
```

---

## Component Hierarchy

### Parent-Child Relationships

```
                        ┌──────────────────────┐
                        │     app.Model        │
                        │   (Root/Parent)      │
                        │                      │
                        │ - Coordinates all    │
                        │ - Manages GCP client │
                        │ - Handles routing    │
                        └──────────┬───────────┘
                                   │
               ┌───────────────────┼───────────────────┐
               │                   │                   │
               │                   │                   │
   ┌───────────▼──────────┐ ┌──────▼──────┐ ┌─────────▼─────────┐
   │   topics.Model       │ │publisher.Model│ │subscriber.Model  │
   │                      │ │               │ │                  │
   │ - Topic list         │ │ - File list   │ │ - Message list   │
   │ - Filter state       │ │ - Variables   │ │ - Ack state      │
   │ - Create/Delete UI   │ │ - Publish UI  │ │ - Detail view    │
   └──────────────────────┘ └───────────────┘ └──────────────────┘
               │
               │
   ┌───────────▼──────────┐
   │subscriptions.Model   │
   │                      │
   │ - Subscription list  │
   │ - Filter state       │
   │ - Active indicator   │
   └──────────┬───────────┘
              │
              │
   ┌──────────▼───────────┐
   │  activity.Model      │
   │                      │
   │ - Log entries        │
   │ - Auto-scroll        │
   │ - Color coding       │
   └──────────────────────┘
```

### Component Communication

```
Components never talk directly. They send messages through parent:

   ┌──────────┐                              ┌──────────┐
   │  Topics  │                              │Publisher │
   └─────┬────┘                              └─────▲────┘
         │                                         │
         │ TopicSelectedMsg                       │
         ▼                                         │
   ┌─────────────────────────────────────────┐    │
   │              app.Model                  │    │
   │                                         │    │
   │  1. Receives TopicSelectedMsg           │    │
   │  2. Updates own state                   │    │
   │  3. Calls publisher.SetTargetTopic()    │────┘
   │  4. Calls subscriptions.SetTopicFilter()│────┐
   │  5. Returns LogMsg                      │    │
   └─────────────────────────────────────────┘    │
                                                  │
                                                  ▼
                                           ┌──────────────┐
                                           │Subscriptions │
                                           └──────────────┘
```

---

## Message Flow Diagrams

### Synchronous Message Flow (Selection)

```
User Action (Enter key)
        │
        ▼
┌───────────────────┐
│ Keyboard Input    │
│ (KeyMsg)          │
└────────┬──────────┘
         │
         ▼
┌────────────────────────────┐
│ topics.Update(KeyMsg)      │
│ - Detects Enter            │
│ - Gets selected topic      │
│ - Returns TopicSelectedMsg │
└────────┬───────────────────┘
         │
         ▼
┌────────────────────────────────────┐
│ app.Update(TopicSelectedMsg)       │
│ - Updates m.selectedTopic          │
│ - Synchronizes child components    │
│ - Returns LogMsg                   │
└────────┬───────────────────────────┘
         │
         ▼
┌────────────────────────────┐
│ app.Update(LogMsg)         │
│ - Forwards to activity     │
│ - Activity adds log        │
└────────┬───────────────────┘
         │
         ▼
┌────────────────────────────┐
│ View() called              │
│ - All components render    │
│ - Screen updates           │
└────────────────────────────┘
```

### Asynchronous Message Flow (API Call)

```
User Action (Create Topic)
        │
        ▼
┌───────────────────────────┐
│ topics.Update(KeyMsg)     │
│ - Returns CreateTopicMsg  │
└────────┬──────────────────┘
         │
         ▼
┌──────────────────────────────────────┐
│ app.Update(CreateTopicMsg)           │
│ - Returns createTopic() command      │
└────────┬─────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────┐
│ Command executes in goroutine        │
│ ┌────────────────────────────────┐   │
│ │ ctx := context.Background()    │   │
│ │ err := client.CreateTopic(ctx) │   │
│ │ return TopicCreatedMsg{Err}    │   │
│ └────────────────────────────────┘   │
└────────┬─────────────────────────────┘
         │ (async)
         │
         ▼
┌──────────────────────────────────────┐
│ app.Update(TopicCreatedMsg)          │
│ - If success: refresh topics list    │
│ - If error: show error message       │
│ - Returns LogMsg                     │
└────────┬─────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────┐
│ View() renders updated state         │
└──────────────────────────────────────┘
```

### Continuous Stream Flow (Subscription)

```
Start Subscription
        │
        ▼
┌───────────────────────────────────────┐
│ startSubscription()                   │
│ - Creates subscription stream         │
│ - Returns pollMessages() command      │
└────────┬──────────────────────────────┘
         │
         ▼
    ┌────────────────────────────────┐
    │    Polling Loop                │
    │                                │
    │ ┌──────────────────────────┐   │
    │ │ pollMessages()           │   │
    │ │ - Blocks on channel      │   │
    │ │ - Waits for message      │   │
    │ └───────┬──────────────────┘   │
    │         │                      │
    │         ▼                      │
    │ ┌──────────────────────────┐   │
    │ │ Message arrives          │   │
    │ │ Returns MessageReceivedMsg│  │
    │ └───────┬──────────────────┘   │
    │         │                      │
    └─────────┼──────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ app.Update(MessageReceivedMsg)     │
│ - Forwards to subscriber.Update()  │
│ - Returns pollMessages() again     │──┐
└─────────────────────────────────────┘  │
              │                          │
              ▼                          │
┌─────────────────────────────────────┐  │
│ subscriber.Update()                 │  │
│ - Adds message to list              │  │
│ - Updates UI                        │  │
└─────────────────────────────────────┘  │
              │                          │
              ▼                          │
┌─────────────────────────────────────┐  │
│ View() renders new message          │  │
└─────────────────────────────────────┘  │
              │                          │
              └──────────────────────────┘
                (Loop continues)
```

---

## Key Workflows

### Workflow 1: Application Startup

```
main.go
  │
  ├─ 1. Verify GCP credentials
  │    └─ Check GOOGLE_CLOUD_PROJECT
  │    └─ Check gcloud config
  │
  ├─ 2. Create Pub/Sub client
  │    └─ pubsub.NewClient(projectID)
  │
  └─ 3. Create BubbleTea program
       └─ tea.NewProgram(app.New(...))
            │
            ▼
       app.New()
         │
         ├─ Initialize child components
         │  ├─ topics.New()
         │  ├─ subscriptions.New()
         │  ├─ publisher.New()
         │  ├─ subscriber.New()
         │  └─ activity.New()
         │
         └─ Return initial model
              │
              ▼
         app.Init()
           │
           ├─ loadTopics() command
           ├─ loadSubscriptions() command
           └─ publisher.LoadFiles() command
                │
                ▼ (Async execution)
           Commands complete
                │
                ├─ TopicsLoadedMsg
                ├─ SubscriptionsLoadedMsg
                └─ FilesLoadedMsg
                     │
                     ▼
                All handled in Update()
                     │
                     ▼
                View() renders
                     │
                     ▼
                App ready for use
```

### Workflow 2: Publishing a Message

```
1. User selects topic
   └─ Topics panel sends TopicSelectedMsg
        └─ Publisher receives target topic

2. User selects file
   └─ Publisher loads file preview

3. User enters variables (optional)
   └─ "key1=value1 key2=value2"

4. User presses Enter
        │
        ▼
   publisher.Update(KeyMsg)
        │
        ├─ Read file content
        ├─ Apply variable substitution
        ├─ Validate JSON
        │
        └─ Return PublishRequestMsg
             │
             ▼
   app.Update(PublishRequestMsg)
        │
        └─ Return publishMessage() command
             │
             ▼
   Command executes (async)
        │
        ├─ Call GCP API
        │  ctx := context.Background()
        │  result := client.Publish(ctx, topic, data)
        │
        └─ Return PublishResultMsg{ID, Err}
             │
             ▼
   app.Update(PublishResultMsg)
        │
        ├─ Forward to publisher.Update()
        │  └─ Update status display
        │
        └─ Return LogMsg
             │
             ▼
   activity.Update(LogMsg)
        │
        └─ Add log entry
             │
             ▼
   View() renders success/error
```

### Workflow 3: Receiving Messages

```
1. User selects subscription
        │
        ▼
   subscriptions.Update(KeyMsg: Enter)
        │
        └─ Return SubscriptionSelectedMsg
             │
             ▼
   app.Update(SubscriptionSelectedMsg)
        │
        ├─ Stop previous subscription (if any)
        │
        ├─ Create new subscription stream
        │  └─ activeSubscription = client.Subscribe(name)
        │
        └─ Return startSubscription() command
             │
             ▼
   startSubscription()
        │
        ├─ Start stream
        │  └─ activeSubscription.Start(ctx)
        │
        └─ Return pollMessages() command
             │
             ▼
        ┌────────────────────┐
        │   Polling Loop     │
        │                    │
    ┌───▼────────────────────▼───┐
    │ pollMessages()             │
    │ - Wait on channel          │
    │ - Receive message          │
    │ - Return MessageReceivedMsg│
    └───┬────────────────────┬───┘
        │                    │
        ▼                    │
   subscriber.Update()       │
        │                    │
        ├─ Add to list       │
        ├─ Update detail     │
        └─ Render            │
             │               │
             └───────────────┘
             (Continue polling)

2. User views message
   └─ Press Enter
        └─ Show detail view

3. User acknowledges message
   └─ Press 'a'
        └─ Call msg.Ack()
        └─ Mark as acknowledged
        └─ Keep in list with checkmark
```

### Workflow 4: Filtering

```
User presses '/' in any panel
        │
        ▼
   component.Update(KeyMsg: '/')
        │
        ├─ Set mode = ModeFilter
        ├─ Focus filter input
        └─ Return updated model
             │
             ▼
   View() renders filter input
        │
User types regex pattern
        │
        ▼
   component.Update(KeyMsg: 'a')
        │
        ├─ Update filterInput
        ├─ Apply filter in real-time
        │  └─ Match each item against regex
        │  └─ Update filtered list
        └─ Return updated model
             │
             ▼
   View() renders filtered list
        │
User presses Esc or Enter
        │
        ▼
   component.Update(KeyMsg: 'Esc')
        │
        ├─ Set mode = ModeNormal
        ├─ Keep filter applied
        └─ Return updated model
             │
             ▼
   View() renders normal mode with filter
```

---

## State Management

### State Ownership

```
┌─────────────────────────────────────────────────────────┐
│                     app.Model                           │
│                                                         │
│  Owns:                                                  │
│  • selectedTopic         (which topic is selected)      │
│  • selectedSubscription  (which sub is active)          │
│  • activeSubscription    (subscription connection)      │
│  • focus                 (which panel is focused)       │
│  • width, height         (window dimensions)            │
│  • client                (GCP Pub/Sub client)           │
│                                                         │
│  Child Components:                                      │
│  ┌───────────────────────────────────────────────────┐ │
│  │ topics.Model                                      │ │
│  │ Owns: allTopics, filterText, mode, loading       │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │ subscriptions.Model                               │ │
│  │ Owns: allSubs, topicFilter, filterText, mode     │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │ publisher.Model                                   │ │
│  │ Owns: files, selectedFile, variables, status     │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │ subscriber.Model                                  │ │
│  │ Owns: messages, autoAck, filterText, detailView  │ │
│  └───────────────────────────────────────────────────┘ │
│  ┌───────────────────────────────────────────────────┐ │
│  │ activity.Model                                    │ │
│  │ Owns: logEntries                                 │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### State Synchronization

```
When a topic is selected:

┌──────────────────────────────────────────────────────────┐
│              TopicSelectedMsg received                   │
└────────────────────┬─────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │   app.Update()          │
        │                         │
        │  m.selectedTopic = name │
        └────────────┬────────────┘
                     │
     ┌───────────────┼───────────────┐
     │               │               │
     ▼               ▼               ▼
┌──────────┐   ┌──────────┐   ┌──────────┐
│ topics   │   │   subs   │   │publisher │
│.SetSel   │   │.SetFilt  │   │.SetTgt   │
│ected     │   │er        │   │          │
└──────────┘   └──────────┘   └──────────┘
     │               │               │
     └───────────────┼───────────────┘
                     │
                  Synced!
```

### State Flow Direction

```
                    ┌──────────────┐
                    │  app.Model   │
                    │   (Parent)   │
                    └───────┬──────┘
                            │
                    State flows DOWN
                    (via method calls)
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
   ┌─────────┐         ┌─────────┐        ┌─────────┐
   │ Topics  │         │  Subs   │        │  Pub    │
   └────┬────┘         └────┬────┘        └────┬────┘
        │                   │                  │
        │                   │                  │
    Messages flow UP (via return values)
        │                   │                  │
        └───────────────────┼──────────────────┘
                            │
                    ┌───────▼──────┐
                    │  app.Model   │
                    │   (Parent)   │
                    └──────────────┘
```

---

## Integration Architecture

### GCP Integration Flow

```
┌──────────────────────────────────────────────────────────┐
│                    Application                           │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │              internal/pubsub/                      │ │
│  │          (Abstraction Layer)                       │ │
│  │                                                    │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐       │ │
│  │  │  Client  │  │  Topics  │  │Publisher │       │ │
│  │  └─────┬────┘  └─────┬────┘  └─────┬────┘       │ │
│  │        │             │             │             │ │
│  └────────┼─────────────┼─────────────┼─────────────┘ │
│           │             │             │               │
└───────────┼─────────────┼─────────────┼───────────────┘
            │             │             │
            │  Wraps GCP SDK            │
            │             │             │
┌───────────▼─────────────▼─────────────▼───────────────┐
│          Google Cloud Pub/Sub Go SDK                  │
│                                                        │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐     │
│  │ pubsub.    │  │ pubsub.    │  │ pubsub.    │     │
│  │ Client     │  │ Topic      │  │ Subscription│     │
│  └────────────┘  └────────────┘  └────────────┘     │
└───────────────────────────┬────────────────────────────┘
                            │
                    HTTPS / gRPC
                            │
┌───────────────────────────▼────────────────────────────┐
│         Google Cloud Pub/Sub Service                   │
│                                                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────────┐       │
│  │  Topics  │  │  Subs    │  │  Messages    │       │
│  └──────────┘  └──────────┘  └──────────────┘       │
└────────────────────────────────────────────────────────┘
```

### Authentication Flow

```
Application Startup
        │
        ▼
┌─────────────────────────┐
│ GetProjectID()          │
│                         │
│ 1. Check env vars       │
│    • GOOGLE_CLOUD_PROJECT
│    • GCLOUD_PROJECT     │
│                         │
│ 2. Fallback to gcloud   │
│    • gcloud config get  │
│                         │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ VerifyCredentials()     │
│                         │
│ 1. Try to create client │
│ 2. Test API call        │
│                         │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ NewClient(projectID)    │
│                         │
│ Uses ADC:               │
│ 1. GOOGLE_APPLICATION_  │
│    CREDENTIALS env var  │
│ 2. gcloud auth          │
│ 3. GCE metadata         │
│                         │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│ Client ready            │
│ Start TUI               │
└─────────────────────────┘
```

### Data Flow: Publishing

```
┌──────────────┐
│ UI Component │
│  (Publisher) │
└──────┬───────┘
       │ PublishRequestMsg
       ▼
┌──────────────────┐
│  app.Update()    │
│  Creates command │
└──────┬───────────┘
       │
       ▼
┌────────────────────────────┐
│ publishMessage() command   │
│                            │
│  internal/pubsub/          │
│  Publisher.Publish()       │
└──────┬─────────────────────┘
       │
       ▼
┌────────────────────────────┐
│ GCP SDK                    │
│ topic.Publish(ctx, msg)    │
└──────┬─────────────────────┘
       │ HTTPS/gRPC
       ▼
┌────────────────────────────┐
│ Google Cloud Pub/Sub       │
│ Stores message             │
└────────────────────────────┘
```

### Data Flow: Subscribing

```
┌────────────────────────────┐
│ Google Cloud Pub/Sub       │
│ Has messages in topic      │
└──────┬─────────────────────┘
       │ gRPC Stream
       ▼
┌────────────────────────────┐
│ GCP SDK                    │
│ sub.Receive(ctx, handler)  │
└──────┬─────────────────────┘
       │
       ▼
┌────────────────────────────┐
│ internal/pubsub/           │
│ Subscription wrapper       │
│ - Receives messages        │
│ - Sends to channel         │
└──────┬─────────────────────┘
       │ Go channel
       ▼
┌────────────────────────────┐
│ pollMessages() command     │
│ - Blocks on channel        │
│ - Returns MessageReceivedMsg
└──────┬─────────────────────┘
       │
       ▼
┌────────────────────────────┐
│ subscriber.Update()        │
│ - Adds to UI list          │
│ - Renders message          │
└────────────────────────────┘
```

---

## Component Interaction Matrix

```
┌────────┬───────┬──────┬────────┬──────────┬─────────┐
│ From→  │Topics │ Subs │  Pub   │   Sub    │Activity │
│   To↓  │       │      │        │          │         │
├────────┼───────┼──────┼────────┼──────────┼─────────┤
│ Topics │   -   │  ✓   │   ✓    │    -     │    ✓    │
│        │       │filter│ target │          │   log   │
├────────┼───────┼──────┼────────┼──────────┼─────────┤
│  Subs  │   -   │  -   │   -    │    ✓     │    ✓    │
│        │       │      │        │  start   │   log   │
├────────┼───────┼──────┼────────┼──────────┼─────────┤
│  Pub   │   -   │  -   │   -    │    -     │    ✓    │
│        │       │      │        │          │   log   │
├────────┼───────┼──────┼────────┼──────────┼─────────┤
│  Sub   │   -   │  -   │   -    │    -     │    ✓    │
│        │       │      │        │          │   log   │
├────────┼───────┼──────┼────────┼──────────┼─────────┤
│Activity│   -   │  -   │   -    │    -     │    -    │
│        │       │      │        │          │         │
└────────┴───────┴──────┴────────┴──────────┴─────────┘

Legend:
  ✓ = Communication via messages through parent
  - = No direct communication

Note: ALL communication goes through app.Model (parent)
```

---

## Summary: Key Architectural Properties

### Unidirectional Data Flow
```
User Input → Message → Update → Model → View → Render
     ↑                                            │
     └────────────────────────────────────────────┘
```

### Single Source of Truth
```
Each piece of state has ONE owner
Changes propagate top-down
Events bubble bottom-up
```

### Functional Composition
```
Complex UI = Composition of Simple Components
Each component: Independent, Testable, Reusable
```

### Asynchronous by Default
```
All I/O wrapped in commands
UI never blocks
Results return as messages
```

### Message-Driven Architecture
```
Everything is a message
Components communicate via messages
Parent coordinates all interactions
```

---

For more details, see:
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Full architecture documentation
- [ARCHITECTURE_QUICK_START.md](./ARCHITECTURE_QUICK_START.md) - Quick start guide

