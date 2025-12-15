# Google Cloud Pub/Sub TUI - Requirements Document

## Document Information
- **Project Name**: Google Cloud Pub/Sub Terminal User Interface (TUI)
- **Version**: 1.0
- **Date**: 2025-12-14
- **Status**: Reverse-Engineered from Implementation

## 1. Executive Summary

### 1.1 Purpose
This document specifies the requirements for a Terminal User Interface (TUI) application that provides an interactive, keyboard-driven interface for managing Google Cloud Pub/Sub resources, publishing messages, and subscribing to message streams.

### 1.2 Scope
The application provides developers and operators with a fast, efficient terminal-based tool for:
- Browsing and managing Pub/Sub topics and subscriptions
- Publishing messages with template variable substitution
- Receiving and viewing messages in real-time
- Managing message acknowledgments
- Filtering resources and messages using regex patterns

### 1.3 Target Users
- Cloud developers working with GCP Pub/Sub
- DevOps engineers managing message queues
- QA engineers testing Pub/Sub workflows
- Anyone requiring quick terminal access to Pub/Sub operations

## 2. System Architecture

### 2.1 Technology Stack

**Programming Language**: Go 1.20+

**Core Dependencies**:
- **Bubbletea**: TUI framework implementing the Elm Architecture (Model-View-Update pattern)
- **Bubbles**: Pre-built TUI components (lists, text inputs, viewports)
- **Lipgloss**: Terminal styling and layout library
- **Google Cloud Pub/Sub SDK**: Official GCP client library

### 2.2 Architectural Pattern

**Model-View-Update (MVU) Pattern**:
- **Model**: Application state (topics, subscriptions, messages, UI state)
- **View**: Pure rendering functions that transform model to terminal output
- **Update**: State transitions based on user input and system events

**Component Architecture**:
- Modular components with independent state
- Parent app coordinates child components
- Message passing for inter-component communication

### 2.3 Project Structure

```
pubsub-tui/
├── main.go                           # Application entry point
├── internal/
│   ├── app/                          # Main application coordinator
│   │   ├── app.go                    # Application model and initialization
│   │   ├── update.go                 # Central state update logic
│   │   └── view.go                   # Layout and rendering
│   ├── components/                   # UI components
│   │   ├── topics/                   # Topics panel
│   │   │   ├── model.go              # State and initialization
│   │   │   ├── update.go             # Event handling
│   │   │   └── view.go               # Rendering
│   │   ├── subscriptions/            # Subscriptions panel
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── publisher/                # Publisher panel
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   ├── view.go
│   │   │   └── substitution.go       # Variable substitution logic
│   │   ├── subscriber/               # Subscriber panel
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   ├── activity/                 # Activity log panel
│   │   │   ├── model.go
│   │   │   ├── update.go
│   │   │   └── view.go
│   │   └── common/                   # Shared components
│   │       ├── messages.go           # Message types for inter-component communication
│   │       └── styles.go             # Shared UI styles
│   ├── pubsub/                       # GCP Pub/Sub wrapper
│   │   ├── client.go                 # Client initialization and management
│   │   ├── auth.go                   # Authentication verification
│   │   ├── topics.go                 # Topic operations
│   │   ├── subscriptions.go          # Subscription operations
│   │   ├── publisher.go              # Publishing operations
│   │   └── subscriber.go             # Subscription streaming
│   └── utils/                        # Utility functions
│       ├── regex.go                  # Regex filtering utilities
│       ├── json.go                   # JSON formatting and validation
│       ├── file.go                   # File operations
│       └── files.go                  # File listing utilities
└── testdata/                         # Sample message templates
```

## 3. Functional Requirements

### 3.1 Authentication and Project Management

#### FR-1.1: GCP Authentication
**Priority**: Critical
**Description**: The application must authenticate using Google Cloud SDK default credentials.

**Requirements**:
- Use Application Default Credentials (ADC)
- Verify authentication on startup
- Display clear error messages if authentication fails
- Provide instructions for setting up authentication

**Implementation Details**:
- Check for `GOOGLE_CLOUD_PROJECT` environment variable
- Fallback to `gcloud config get-value project`
- Verify credentials before initializing UI
- Exit gracefully with helpful error messages if auth fails

#### FR-1.2: Project Detection
**Priority**: Critical
**Description**: Automatically detect the active GCP project.

**Requirements**:
- Check `GOOGLE_CLOUD_PROJECT` environment variable first
- Fallback to gcloud configuration
- Display project ID in the UI footer
- Handle missing project configuration gracefully

### 3.2 User Interface Layout

#### FR-2.1: Multi-Panel Layout
**Priority**: Critical
**Description**: The application must display a multi-panel layout with proper sizing and focus management.

**Layout Specification**:

```
┌─────────────────────────────────────────────────────────────────┐
│ Left Panel (1/3 width)     │  Right Panel (2/3 width)           │
│                            │                                     │
│ ┌─ Topics (30% height) ─┐ │ ┌─ Publisher (33% height) ────────┐│
│ │ • topic-1              │ │ │ JSON Files:  [file-list]        ││
│ │ • topic-2              │ │ │ Preview:     [file preview]     ││
│ │ • topic-3              │ │ │ Variables:   [key=val input]    ││
│ │ Filter: [/regex/]      │ │ │ Status:      [success/error]    ││
│ └────────────────────────┘ │ └─────────────────────────────────┘│
│                            │                                     │
│ ┌─ Subscriptions (30%) ─┐ │ ┌─ Subscriber (67% height) ───────┐│
│ │ • subscription-1       │ │ │ Messages:    [msg list]         ││
│ │ • subscription-2       │ │ │              [msg detail]       ││
│ │ Filter: [/regex/]      │ │ │ Filter:      [/regex/]          ││
│ └────────────────────────┘ │ │ Auto-ack:    [x] enabled        ││
│                            │ └─────────────────────────────────┘│
│ ┌─ Activity Log (40%) ──┐ │                                     │
│ │ [12:00:01] Connected   │ │                                     │
│ │ [12:00:05] Loaded 5... │ │                                     │
│ │ [12:00:10] Published   │ │                                     │
│ └────────────────────────┘ │                                     │
└─────────────────────────────────────────────────────────────────┘
Tab: cycle panels | q/Ctrl+C: quit | Project: my-gcp-project
```

**Panel Dimensions**:
- Left panel: 1/3 of terminal width
  - Topics: 30% of available height
  - Subscriptions: 30% of available height
  - Activity Log: 40% of available height
- Right panel: 2/3 of terminal width
  - Publisher: 1/3 of available height
  - Subscriber: 2/3 of available height
- Footer: 2 lines for help text and project info

#### FR-2.2: Focus Management
**Priority**: Critical
**Description**: Only one panel can be focused at a time, indicated by visual styling.

**Requirements**:
- Focused panel has colored border
- Unfocused panels have muted border
- Tab key cycles focus: Topics → Subscriptions → Publisher → Subscriber → Topics
- Focused panel receives all keyboard input
- Activity log is non-focusable (display-only)

#### FR-2.3: Responsive Sizing
**Priority**: High
**Description**: UI must adapt to terminal window size changes.

**Requirements**:
- Handle `WindowSizeMsg` events
- Recalculate panel dimensions on resize
- Update all child component sizes
- Prevent rendering with invalid dimensions (0 width/height)
- Maintain proportional layout at any size

### 3.3 Topics Panel

#### FR-3.1: Topic Listing
**Priority**: Critical
**Description**: Display all Pub/Sub topics in the current GCP project.

**Requirements**:
- Load topics on application startup
- Display topics in a scrollable list
- Show loading state while fetching
- Handle empty topic list gracefully
- Display error messages if loading fails
- Support keyboard navigation (arrow keys, j/k)

#### FR-3.2: Topic Selection
**Priority**: Critical
**Description**: Allow users to select a topic to filter subscriptions and set publish target.

**Requirements**:
- Highlight selected topic
- Press Enter to select
- On selection:
  - Filter subscriptions panel to show only subscriptions for this topic
  - Set publisher target to selected topic
  - Move focus to subscriptions panel
  - Log selection in activity panel
  - If a topic was already selected, it will disconnect the subscription. Notify the user and ask for confirmation

#### FR-3.3: Topic Filtering
**Priority**: High
**Description**: Filter topics using regex patterns.

**Requirements**:
- Press `/` to activate filter mode
- Display filter input field
- Apply regex filter in real-time
- Show filtered count
- Handle invalid regex gracefully
- Press Esc to clear filter and return to list navigation
- Filter is case-sensitive

#### FR-3.4: Topic Creation
**Priority**: Medium
**Description**: Create new topics from the UI.

**Requirements**:
- Press `n` (new) to enter creation mode
- Display input field for topic ID
- Validate topic ID format
- Call GCP API to create topic
- Refresh topic list on success
- Display error message on failure
- Log operation in activity panel

#### FR-3.5: Topic Deletion
**Priority**: Medium
**Description**: Delete existing topics.

**Requirements**:
- Press `d` (delete) to confirm deletion
- Show confirmation prompt
- Call GCP API to delete topic
- Refresh topic list on success
- Display error message on failure
- Log operation in activity panel

### 3.4 Subscriptions Panel

#### FR-4.1: Subscription Listing
**Priority**: Critical
**Description**: Display all Pub/Sub subscriptions in the current GCP project.

**Requirements**:
- Load subscriptions on application startup
- Display subscription name and associated topic
- Show loading state while fetching
- Handle empty subscription list gracefully
- Support keyboard navigation

**Data Display**:
- Title: Subscription name
- Description: Associated topic name

#### FR-4.2: Subscription Selection
**Priority**: Critical
**Description**: Select a subscription to start receiving messages.

**Requirements**:
- Highlight selected subscription
- Press Enter to start subscription
- On selection:
  - Stop any existing active subscription
  - Start receiving messages from selected subscription
  - Update subscriber panel with subscription info
  - Move focus to publisher panel
  - Log subscription start in activity panel

#### FR-4.3: Topic-Based Filtering
**Priority**: Critical
**Description**: Automatically filter subscriptions when a topic is selected.

**Requirements**:
- When topic is selected, show only subscriptions for that topic
- Display current filter topic in panel header
- Allow clearing topic filter to show all subscriptions
- Combine topic filter with regex filter

#### FR-4.4: Regex Filtering
**Priority**: High
**Description**: Filter subscriptions using regex patterns.

**Requirements**:
- Press `/` to activate filter mode
- Display filter input field
- Apply regex filter in real-time
- Filter on subscription names only (not topics)
- Handle invalid regex gracefully
- Press Esc to clear filter

#### FR-4.5: Subscription Creation
**Priority**: Medium
**Description**: Create new subscriptions from the UI.

**Requirements**:
- Press `n` to enter creation mode
- Require topic selection first
- Display input field for subscription ID
- Create subscription linked to selected topic
- Refresh subscription list on success
- Log operation in activity panel

#### FR-4.6: Subscription Deletion
**Priority**: Medium
**Description**: Delete existing subscriptions.

**Requirements**:
- Press `d` to confirm deletion
- Show confirmation prompt
- Call GCP API to delete subscription
- Refresh subscription list on success
- Log operation in activity panel

### 3.5 Publisher Panel

#### FR-5.1: JSON File Selection
**Priority**: Critical
**Description**: Select JSON files from the working directory for message publishing.

**Requirements**:
- Auto-load all `.json` files from working directory on startup
- Display files in scrollable list
- Show file preview on selection
- Support keyboard navigation
- Display "no files found" message if directory is empty

**File Preview**:
- Display file contents in viewport
- Support scrolling for large files
- Syntax highlighting for JSON (optional)
- Show first file by default

#### FR-5.2: Variable Substitution
**Priority**: Critical
**Description**: Support template variables in message files with runtime substitution.

**Template Syntax**: `${variableName}`

**Variable Input Format**: Space-separated key=value pairs
- Example: `user=john env=prod timestamp=2024-01-01`

**Requirements**:
- Parse variable definitions from input field
- Find all `${var}` patterns in message file
- Replace with corresponding values
- Keep original `${var}` if no value provided
- Display variables input field
- Press `v` to focus variables input
- Validate key=value format

**Example**:
```json
{
  "user": "${user}",
  "environment": "${env}",
  "data": "static value"
}
```
With variables: `user=alice env=production`

Becomes:
```json
{
  "user": "alice",
  "environment": "production",
  "data": "static value"
}
```

#### FR-5.3: Message Publishing
**Priority**: Critical
**Description**: Publish messages to the selected topic.

**Requirements**:
- Require topic selection before publishing
- Read selected JSON file
- Apply variable substitution
- Validate JSON format
- Publish to GCP Pub/Sub topic
- Display message ID on success
- Display error message on failure
- Log operation in activity panel

**Trigger Methods**:
- Press Enter when focused on variables input
- Press `Ctrl+P` anywhere in publisher panel
- Click publish button (if implemented)

**Status Display**:
- Show "Publishing..." during operation
- Show "Published successfully: [message-id]" on success
- Show "Publish failed: [error]" on failure
- Clear status after next operation

#### FR-5.4: File Content Preview
**Priority**: High
**Description**: Preview message file contents before publishing.

**Requirements**:
- Display file content in read-only viewport
- Update preview when file selection changes
- Support scrolling with arrow keys or mouse
- Format JSON with proper indentation
- Show substitution variables highlighted (optional)

#### FR-5.5: Target Topic Display
**Priority**: Medium
**Description**: Show which topic messages will be published to.

**Requirements**:
- Display target topic name in panel header
- Update when topic selection changes
- Show warning if no topic selected
- Prevent publishing without topic selection

### 3.6 Subscriber Panel

#### FR-6.1: Real-Time Message Reception
**Priority**: Critical
**Description**: Receive and display messages from the active subscription in real-time.

**Requirements**:
- Start streaming on subscription selection
- Display messages as they arrive
- Buffer up to 100 messages in memory
- Show message ID, publish time, and preview
- Maintain message order (newest at bottom)
- Continue receiving until subscription changes or app quits

**Message List Display**:
- Format: `[✓] <msg-id>... <time>`
- `✓` indicates acknowledged message
- Show first 8 characters of message ID
- Show time in HH:MM:SS format
- Show first 50 characters of data as description

#### FR-6.2: Message Detail View
**Priority**: Critical
**Description**: Display full message details in a side panel.

**Requirements**:
- Press Enter on selected message to view details
- Split panel: list on left (50%), detail on right (50%)
- Display full message data
- Format JSON with syntax highlighting
- Show message attributes
- Show publish timestamp
- Show acknowledgment status
- Press Esc to close detail view

**Detail View Content**:
```
Message ID: abc123def456...
Published: 2024-01-01 12:34:56
Status: [Acknowledged | Pending]

Attributes:
  key1: value1
  key2: value2

Data:
{
  "formatted": "json",
  "with": "colors"
}
```

#### FR-6.3: Message Acknowledgment
**Priority**: Critical
**Description**: Manually or automatically acknowledge received messages.

**Manual Acknowledgment**:
- Press `a` to acknowledge selected message
- Mark message with checkmark (✓)
- Call Pub/Sub ack function
- Log acknowledgment in activity panel
- Keep message in list after ack

**Auto-Acknowledgment**:
- Press `A` to toggle auto-ack mode
- Display toggle state in panel header: `Auto-ack: [x] enabled` or `[ ] disabled`
- When enabled, acknowledge messages immediately upon receipt
- When disabled, require manual acknowledgment
- Persist toggle state during session

#### FR-6.4: Message Filtering
**Priority**: High
**Description**: Filter messages using regex patterns.

**Requirements**:
- Press `/` to activate filter mode
- Display filter input field
- Apply regex filter in real-time
- Search both message ID and message data
- Update filtered list dynamically
- Press Esc to clear filter

#### FR-6.5: Message Counter
**Priority**: Low
**Description**: Display count of received messages.

**Requirements**:
- Show total messages received
- Show filtered message count when filter active
- Format: `Messages: 45` or `Messages: 12 / 45 (filtered)`

#### FR-6.6: Subscription Info Display
**Priority**: Medium
**Description**: Show active subscription details.

**Requirements**:
- Display subscription name in panel header
- Display associated topic name
- Show "Not subscribed" state when no active subscription
- Update on subscription change

### 3.7 Activity Log Panel

#### FR-7.1: Activity Logging
**Priority**: High
**Description**: Display timestamped log entries for all significant operations.

**Requirements**:
- Non-interactive display-only panel
- Auto-scroll to bottom on new entries
- Show timestamp in HH:MM:SS format
- Color-code by log level
- Store all logs in memory (no limit)

**Log Levels**:
- **Info** (default): General operations
- **Success** (green): Successful operations
- **Warning** (yellow): Warnings
- **Error** (red): Errors and failures
- **Network** (blue): Network operations (API calls)

**Log Format**: `[HH:MM:SS] Message text`

**Logged Events**:
- Application startup
- GCP client connection
- Topic/subscription loading
- Topic/subscription selection
- Message publishing (attempt, success, failure)
- Message receipt
- Message acknowledgment
- Resource creation/deletion
- Errors and warnings

#### FR-7.2: Log Scrolling
**Priority**: Low
**Description**: Support scrolling through log history.

**Requirements**:
- Auto-scroll to bottom by default
- Allow manual scrolling when activity panel is focused (future enhancement)
- Show scroll position indicator

### 3.8 Global Navigation and Controls

#### FR-8.1: Panel Navigation
**Priority**: Critical
**Description**: Navigate between panels using keyboard.

**Requirements**:
- Tab: Cycle focus forward (Topics → Subscriptions → Publisher → Subscriber)
- Shift+Tab: Cycle focus backward (future enhancement)
- Only one panel focused at a time
- Activity log is not focusable

#### FR-8.2: Application Exit
**Priority**: Critical
**Description**: Gracefully exit the application.

**Requirements**:
- Press `q` or `Ctrl+C` to quit
- Stop active subscriptions
- Close GCP client connections
- Cancel background operations
- Exit alt-screen mode cleanly
- Restore terminal state

#### FR-8.3: Help Text
**Priority**: Medium
**Description**: Display contextual help in the footer.

**Requirements**:
- Always visible at bottom of screen
- Show global shortcuts
- Show current project ID
- Format: `Tab: cycle panels | q/Ctrl+C: quit | Project: my-project`

### 3.9 Error Handling

#### FR-9.1: Startup Errors
**Priority**: Critical
**Description**: Handle initialization errors gracefully.

**Requirements**:
- Display authentication errors with setup instructions
- Display project configuration errors
- Display GCP API errors (permissions, network)
- Show error message in full screen with quit instructions
- Don't enter main UI if critical errors occur

#### FR-9.2: Runtime Errors
**Priority**: High
**Description**: Handle runtime errors without crashing.

**Requirements**:
- Display errors in relevant panel (status bar)
- Log errors in activity panel
- Continue operation after non-critical errors
- Provide actionable error messages
- Include error details for debugging

**Error Categories**:
- Network errors (retry guidance)
- Permission errors (IAM guidance)
- Invalid input (format requirements)
- Resource not found (refresh guidance)

#### FR-9.3: Resource Loading Errors
**Priority**: High
**Description**: Handle GCP resource loading failures.

**Requirements**:
- Show loading state during fetch
- Display error if loading fails
- Provide retry mechanism (future enhancement)
- Log error details in activity panel

## 4. Non-Functional Requirements

### 4.1 Performance

#### NFR-1.1: Startup Time
**Requirement**: Application should start and display UI within 2 seconds on standard hardware with network connectivity.

**Acceptance Criteria**:
- Initialize GCP client in background
- Display loading UI immediately
- Load topics and subscriptions asynchronously

#### NFR-1.2: Message Throughput
**Requirement**: Handle at least 100 messages per second without UI degradation.

**Acceptance Criteria**:
- Buffered channel (100 message capacity)
- Efficient list updates
- Throttled UI rendering (if needed)

#### NFR-1.3: Memory Usage
**Requirement**: Reasonable memory consumption for long-running sessions.

**Considerations**:
- In-memory message storage (unlimited)
- Consider message limit or circular buffer for production use
- Activity log unbounded (may need limit)

### 4.2 Usability

#### NFR-2.1: Keyboard Efficiency
**Requirement**: All operations accessible via keyboard without mouse.

**Acceptance Criteria**:
- No mouse required for any function
- Vim-style navigation (j/k) supported
- Common shortcuts (/, Esc, Enter)
- Consistent key bindings across panels

#### NFR-2.2: Visual Clarity
**Requirement**: Clear visual hierarchy and focus indication.

**Acceptance Criteria**:
- Focused panel clearly distinguished
- Color coding for message types
- Consistent styling across panels
- Readable at minimum terminal size (80x24)

#### NFR-2.3: Responsive Feedback
**Requirement**: Immediate visual feedback for user actions.

**Acceptance Criteria**:
- Input echoed immediately
- Loading states for async operations
- Success/error confirmations
- Status messages for all operations

### 4.3 Reliability

#### NFR-3.1: Error Recovery
**Requirement**: Graceful handling of network and API errors.

**Acceptance Criteria**:
- Continue operation after non-critical errors
- Clear error messages
- No data loss on transient failures
- Automatic reconnection handling (future)

#### NFR-3.2: State Consistency
**Requirement**: Maintain consistent application state.

**Acceptance Criteria**:
- Synchronized panel states
- Correct message passing between components
- No race conditions in concurrent operations
- Proper context cancellation

### 4.4 Compatibility

#### NFR-4.1: Terminal Compatibility
**Requirement**: Work on common terminal emulators.

**Supported Terminals**:
- iTerm2 (macOS)
- Terminal.app (macOS)
- GNOME Terminal (Linux)
- Windows Terminal
- tmux/screen

**Minimum Requirements**:
- 256 color support
- UTF-8 encoding
- Minimum size: 80x24

#### NFR-4.2: Platform Support
**Requirement**: Run on major operating systems.

**Supported Platforms**:
- macOS (Intel and ARM)
- Linux (x86_64, ARM64)
- Windows (WSL2 recommended)

#### NFR-4.3: Go Version
**Requirement**: Compatible with Go 1.20 and later.

### 4.5 Security

#### NFR-5.1: Credential Handling
**Requirement**: Never store or log credentials.

**Acceptance Criteria**:
- Use GCP Application Default Credentials only
- Don't log authentication tokens
- Don't display sensitive data in logs
- Follow Google Cloud security best practices

#### NFR-5.2: Message Privacy
**Requirement**: Handle message data securely.

**Acceptance Criteria**:
- No message data written to disk (except temp files)
- In-memory storage only
- Proper cleanup on exit

## 5. Data Requirements

### 5.1 Input Data

#### DR-1.1: Message Templates
**Format**: JSON files
**Location**: Working directory
**Pattern**: `*.json`

**Requirements**:
- Valid JSON syntax
- Support for variable placeholders `${varName}`
- UTF-8 encoding
- Reasonable file size (< 1MB recommended)

#### DR-1.2: Variable Definitions
**Format**: Space-separated key=value pairs
**Example**: `user=alice env=prod timestamp=2024-01-01`

**Requirements**:
- Key: alphanumeric and underscore
- Value: any string (spaces require quotes - future enhancement)
- Maximum 512 characters total

### 5.2 Output Data

#### DR-2.1: Published Messages
**Format**: Byte array sent to Pub/Sub
**Processing**:
- Read JSON file
- Apply variable substitution
- Send raw bytes to Pub/Sub (no additional encoding)

#### DR-2.2: Activity Logs
**Storage**: In-memory only
**Format**: Timestamped text entries
**Retention**: Session lifetime (cleared on exit)

### 5.3 GCP API Data

#### DR-3.1: Topics
**Source**: `pubsub.Client.Topics()`
**Format**: Array of topic names (strings)
**Caching**: Loaded once at startup, refreshed on create/delete

#### DR-3.2: Subscriptions
**Source**: `pubsub.Client.Subscriptions()`
**Format**: Array of `SubscriptionInfo` objects
**Fields**:
- Name (string): Subscription ID
- Topic (string): Associated topic ID

**Caching**: Loaded once at startup, refreshed on create/delete

#### DR-3.3: Messages
**Source**: Real-time subscription stream
**Format**: `ReceivedMessage` objects
**Fields**:
- ID (string): Message identifier
- Data ([]byte): Message payload
- Attributes (map[string]string): Message metadata
- PublishTime (time.Time): Publication timestamp
- AckID (string): Acknowledgment identifier
- AckFunc (function): Acknowledgment callback
- NackFunc (function): Negative acknowledgment callback

## 6. External Interface Requirements

### 6.1 GCP Authentication

**Method**: Application Default Credentials (ADC)

**Setup Methods** (in order of precedence):
1. `GOOGLE_APPLICATION_CREDENTIALS` environment variable
2. `gcloud auth application-default login` credentials
3. GCE/GKE instance metadata (when running on GCP)

**Required Scopes**:
- `https://www.googleapis.com/auth/pubsub` (full Pub/Sub access)

### 6.2 GCP Pub/Sub API

**API Endpoints Used**:

1. **List Topics**: `projects/{project}/topics`
2. **Create Topic**: `projects/{project}/topics/{topic}`
3. **Delete Topic**: `projects/{project}/topics/{topic}`
4. **List Subscriptions**: `projects/{project}/subscriptions`
5. **Create Subscription**: `projects/{project}/subscriptions/{subscription}`
6. **Delete Subscription**: `projects/{project}/subscriptions/{subscription}`
7. **Publish Message**: `projects/{project}/topics/{topic}:publish`
8. **Pull Messages**: `projects/{project}/subscriptions/{subscription}:streamingPull`
9. **Acknowledge Message**: `projects/{project}/subscriptions/{subscription}:acknowledge`

**Required IAM Permissions**:
- `pubsub.topics.list`
- `pubsub.topics.create` (for topic creation)
- `pubsub.topics.delete` (for topic deletion)
- `pubsub.topics.publish` (for publishing)
- `pubsub.subscriptions.list`
- `pubsub.subscriptions.create` (for subscription creation)
- `pubsub.subscriptions.delete` (for subscription deletion)
- `pubsub.subscriptions.consume` (for receiving messages)

**Recommended Role**: `roles/pubsub.editor`

### 6.3 Environment Variables

**GOOGLE_CLOUD_PROJECT** (optional)
- Description: GCP project ID to use
- Fallback: `gcloud config get-value project`
- Example: `export GOOGLE_CLOUD_PROJECT=my-project-id`

**GOOGLE_APPLICATION_CREDENTIALS** (optional)
- Description: Path to service account key JSON
- Used by: GCP SDK for authentication
- Example: `export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json`

### 6.4 File System

**Working Directory Access**:
- Read: List and read `.json` files for message templates
- Path: Current working directory where app is executed
- Pattern: `*.json` (recursive search not implemented)

**No Write Operations**:
- Application does not write files to disk
- All state is in-memory

## 7. Constraints and Assumptions

### 7.1 Constraints

**Technical Constraints**:
- Terminal-only interface (no GUI)
- Single GCP project per session
- In-memory message storage (no persistence)
- No message replay after acknowledgment
- Synchronous message publishing (one at a time)

**Environmental Constraints**:
- Requires active internet connection
- Requires configured GCP credentials
- Requires terminal with 256-color support
- Minimum terminal size: 80x24 (recommended: 120x30+)

### 7.2 Assumptions

**User Environment**:
- User has GCP project with Pub/Sub enabled
- User has appropriate IAM permissions
- User has configured gcloud CLI or ADC
- User is familiar with Pub/Sub concepts
- User has basic terminal/CLI proficiency

**Operational Assumptions**:
- Single user per app instance
- Short to medium session length (hours, not days)
- Moderate message volume (< 1000 messages/minute)
- Message sizes < 10MB (Pub/Sub limit)

**Development Assumptions**:
- Go 1.20+ is available
- Dependencies are managed via Go modules
- Standard Go build tools are used

## 8. Future Enhancements (Not Currently Implemented)

### 8.1 Potential Features

**High Priority**:
- Message search and history
- Export messages to file
- Configuration file support
- Message deadletter queue viewing
- Subscription configuration display (ack deadline, retention, etc.)
- Batch message publishing

**Medium Priority**:
- Message ordering display
- Filter presets/saved filters
- Multiple project support
- Custom key bindings
- Color theme customization
- Message replay

**Low Priority**:
- Mouse support for navigation
- Clipboard integration
- Message formatting options (XML, Protobuf, etc.)
- Statistics dashboard
- Performance monitoring
- Plugin system

### 8.2 Known Limitations

**Current Implementation Limitations**:
- No message persistence (lost on app exit)
- No pagination for large topic/subscription lists
- No concurrent subscriptions (one active at a time)
- No message size validation before publish
- No retry logic for failed operations
- Unbounded memory growth with unlimited message reception
- No message attributes support in publisher
- No filtering by message attributes in subscriber

## 9. Testing Requirements

### 9.1 Unit Testing

**Target Coverage**: Core business logic

**Test Areas**:
- Variable substitution logic
- Regex filtering utilities
- JSON formatting
- Message parsing
- State management

### 9.2 Integration Testing

**Test Scenarios**:
- GCP client initialization
- Topic/subscription listing
- Message publishing
- Message subscription
- Authentication failure handling

**Test Environment**:
- GCP test project
- Test topics and subscriptions
- Mock Pub/Sub server (optional)

### 9.3 Manual Testing

**Test Cases**:
- Fresh installation flow
- Authentication setup
- All keyboard shortcuts
- Panel navigation
- Window resizing
- Error scenarios
- Long-running sessions
- High message volume

## 10. Documentation Requirements

### 10.1 User Documentation

**README.md** ✅ (Completed)
- Installation instructions
- Setup guide (GCP, gcloud, credentials)
- Usage examples
- Keyboard shortcuts reference
- Troubleshooting guide

**Examples** 📋 (Partial)
- Sample message templates
- Common use cases
- Variable substitution examples

### 10.2 Developer Documentation

**Code Comments**: Inline documentation for public APIs

**Architecture Documentation**: This requirements document

**API Documentation**: Generate with `go doc`

## 11. Acceptance Criteria

### 11.1 Core Functionality

✅ **AC-1**: User can authenticate with GCP and connect to a project
✅ **AC-2**: User can list and select topics
✅ **AC-3**: User can list and select subscriptions
✅ **AC-4**: User can publish messages with variable substitution
✅ **AC-5**: User can receive messages in real-time
✅ **AC-6**: User can acknowledge messages (manually or auto)
✅ **AC-7**: User can filter topics, subscriptions, and messages by regex
✅ **AC-8**: User can navigate using only keyboard
✅ **AC-9**: Application handles errors gracefully
✅ **AC-10**: Application displays activity logs

### 11.2 User Experience

✅ **AC-11**: UI is responsive and provides immediate feedback
✅ **AC-12**: Focused panel is clearly indicated
✅ **AC-13**: All operations have status indicators
✅ **AC-14**: Help text is always visible
✅ **AC-15**: Application exits cleanly

### 11.3 Quality

✅ **AC-16**: Application starts successfully with valid credentials
✅ **AC-17**: Application handles window resize
✅ **AC-18**: No crashes during normal operation
✅ **AC-19**: Memory usage is reasonable for session length
✅ **AC-20**: Works on macOS, Linux, and Windows (WSL)

## 12. Glossary

**ADC**: Application Default Credentials - Google Cloud's credential discovery mechanism

**Ack/Acknowledgment**: Confirmation that a message has been successfully processed, allowing Pub/Sub to remove it from the subscription

**Bubbletea**: A Go framework for building terminal UIs based on The Elm Architecture

**Focus**: The active panel that receives keyboard input

**GCP**: Google Cloud Platform

**MVU**: Model-View-Update architecture pattern (also called The Elm Architecture)

**Nack**: Negative acknowledgment - indicates message processing failed and should be redelivered

**Panel**: A distinct section of the UI (Topics, Subscriptions, Publisher, Subscriber, Activity)

**Pub/Sub**: Google Cloud Pub/Sub - a messaging service for asynchronous communication

**Regex**: Regular expression - a pattern for matching text

**Subscription**: A named resource representing a stream of messages from a topic

**Topic**: A named resource to which messages are published

**TUI**: Terminal User Interface - a text-based user interface in a terminal emulator

**Variable Substitution**: Replacing placeholders (e.g., `${var}`) with actual values

**Viewport**: A scrollable view component for displaying content larger than the available space

---

**Document Revision History**:

| Date | Version | Author | Description |
|------|---------|--------|-------------|
| 2025-12-14 | 1.0 | System | Initial reverse-engineered requirements document |

---

**Approval**:

This document represents the reverse-engineered requirements based on the current implementation of the Google Cloud Pub/Sub TUI application. It serves as both historical documentation and a specification for future enhancements.
