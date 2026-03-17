# API Reference

## Client

```go
client := sonzai.NewClient("sk-...",
    sonzai.WithBaseURL("https://api.sonz.ai"),  // or SONZAI_BASE_URL env var
    sonzai.WithTimeout(60 * time.Second),
)
```

| Field | Type | Description |
|-------|------|-------------|
| `client.Agents` | `*AgentsResource` | Chat, memory, personality, voice, and agent-scoped operations |
| `client.Eval` | `*eval.Client` | Evaluation, simulation, and benchmarking |
| `client.Voices` | `*VoicesResource` | Global voice catalog |

## Agents

### Chat

| Method | Returns | Description |
|--------|---------|-------------|
| `Chat(ctx, agentID, opts)` | `*ChatResponse, error` | Non-streaming chat response |
| `ChatStream(ctx, agentID, opts, callback)` | `error` | Streaming chat via SSE with callback |
| `ChatStreamChannel(ctx, agentID, opts)` | `<-chan ChatStreamEvent` | Streaming chat via channel |

### Agent Management

| Method | Returns | Description |
|--------|---------|-------------|
| `Create(ctx, opts)` | `*Agent, error` | Create a new agent |
| `Get(ctx, agentID)` | `*Agent, error` | Get agent by ID |
| `Update(ctx, agentID, opts)` | `*Agent, error` | Update agent profile |
| `Delete(ctx, agentID)` | `error` | Delete an agent |

### Context Engine Data

| Method | Returns | Description |
|--------|---------|-------------|
| `GetMood(ctx, agentID, userID, instanceID)` | `*MoodResponse, error` | Current mood state |
| `GetMoodHistory(ctx, agentID, userID, instanceID)` | `*MoodHistoryResponse, error` | Mood history |
| `GetMoodAggregate(ctx, agentID, userID, instanceID)` | `*MoodAggregateResponse, error` | Aggregated mood statistics |
| `GetRelationships(ctx, agentID, userID, instanceID)` | `*RelationshipsResponse, error` | Relationship data |
| `GetHabits(ctx, agentID, userID, instanceID)` | `*HabitsResponse, error` | Habit data |
| `GetGoals(ctx, agentID, userID, instanceID)` | `*GoalsResponse, error` | Goals data |
| `GetInterests(ctx, agentID, userID, instanceID)` | `*InterestsResponse, error` | Interest data |
| `GetDiary(ctx, agentID, userID, instanceID)` | `*DiaryResponse, error` | Diary entries |
| `GetUsers(ctx, agentID)` | `*UsersResponse, error` | Users interacting with this agent |

### Events & Dialogue

| Method | Returns | Description |
|--------|---------|-------------|
| `TriggerEvent(ctx, agentID, opts)` | `*TriggerEventResponse, error` | Trigger a game event or activity |
| `Dialogue(ctx, opts)` | `*DialogueResponse, error` | Multi-agent dialogue |

## Agents.Memory

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, agentID, opts)` | `*MemoryTreeResponse, error` | Full memory tree |
| `Search(ctx, agentID, opts)` | `*MemorySearchResponse, error` | Search memories by query |
| `Timeline(ctx, agentID, opts)` | `*MemoryTimelineResponse, error` | Memories in a date range |
| `ListFacts(ctx, agentID, opts)` | `*FactListResponse, error` | Atomic facts with filtering |
| `Reset(ctx, agentID, opts)` | `*MemoryResetResponse, error` | Delete all memory (optionally per-user) |

## Agents.Personality

| Method | Returns | Description |
|--------|---------|-------------|
| `Get(ctx, agentID, opts)` | `*PersonalityResponse, error` | Profile and evolution history |
| `Update(ctx, agentID, opts)` | `*UpdatePersonalityResponse, error` | Update Big5 scores |

## Agents.Sessions

| Method | Returns | Description |
|--------|---------|-------------|
| `Start(ctx, agentID, opts)` | `*SessionResponse, error` | Begin a session |
| `End(ctx, agentID, opts)` | `*SessionResponse, error` | End a session |
| `SetTools(ctx, agentID, sessionID, opts)` | `*SessionToolsResponse, error` | Configure session tools |

## Agents.Instances

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, agentID)` | `*InstanceListResponse, error` | List all instances |
| `Create(ctx, agentID, name, desc)` | `*Instance, error` | Create instance |
| `Get(ctx, agentID, instanceID)` | `*Instance, error` | Get instance |
| `Delete(ctx, agentID, instanceID)` | `error` | Delete instance |
| `Reset(ctx, agentID, instanceID)` | `*InstanceResetResponse, error` | Clear all instance context data |

## Agents.Notifications

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, agentID, opts)` | `*NotificationListResponse, error` | List notifications |
| `Consume(ctx, agentID, messageID)` | `*NotificationConsumeResponse, error` | Mark as consumed |
| `History(ctx, agentID, limit)` | `*NotificationHistoryResponse, error` | Full notification history |

## Agents.CustomState

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, agentID, opts)` | `*CustomStateListResponse, error` | List custom states |
| `Create(ctx, agentID, opts)` | `*CustomState, error` | Create custom state |
| `Update(ctx, agentID, stateID, opts)` | `*CustomState, error` | Update custom state |
| `Delete(ctx, agentID, stateID)` | `error` | Delete custom state |

## Agents.Image

| Method | Returns | Description |
|--------|---------|-------------|
| `Generate(ctx, agentID, opts)` | `*ImageResult, error` | Generate image from prompt |

## Agents.Voice

| Method | Returns | Description |
|--------|---------|-------------|
| `Match(ctx, agentID, opts)` | `*VoiceMatchResult, error` | Find best voice for agent personality |
| `Chat(ctx, agentID, opts)` | `*VoiceChatResponse, error` | Voice chat (audio in, audio + text out) |
| `TTS(ctx, agentID, opts)` | `*TTSResult, error` | Text-to-speech |
| `GetToken(ctx, agentID, opts)` | `*VoiceTokenResponse, error` | Get token for real-time voice stream |
| `Stream(ctx, agentID, opts)` | `*VoiceStream, error` | Real-time WebSocket voice stream |

## Agents.Wakeups

| Method | Returns | Description |
|--------|---------|-------------|
| `Schedule(ctx, agentID, opts)` | `*WakeupResponse, error` | Schedule a proactive check-in |

## Agents.Generation

| Method | Returns | Description |
|--------|---------|-------------|
| `GenerateBio(ctx, agentID)` | `*GenerateBioResponse, error` | Generate agent biography |
| `GenerateCharacter(ctx, opts)` | `*GenerateCharacterResponse, error` | Generate full character profile |
| `GenerateSeedMemories(ctx, agentID, opts)` | `*SeedMemoriesResponse, error` | Generate seed memories |
| `SeedMemories(ctx, agentID, opts)` | `*SeedMemoriesResponse, error` | Plant seed memories |

## Voices (Global)

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, opts)` | `*VoiceListResponse, error` | List available voices from catalog |

## Eval

```go
import "github.com/sonz-ai/sonzai-go/eval"
```

| Method | Returns | Description |
|--------|---------|-------------|
| `Evaluate(ctx, agentID, opts)` | `*EvaluationResult, error` | Evaluate agent responses |
| `Simulate(ctx, agentID, opts, callback)` | `error` | Run multi-turn simulation with SSE streaming |
| `Run(ctx, agentID, opts, callback)` | `error` | Simulation + evaluation combined |
| `ReEval(ctx, agentID, opts, callback)` | `error` | Re-evaluate using a different template |

## Eval.Templates

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, templateType)` | `*TemplateListResponse, error` | List templates |
| `Get(ctx, templateID)` | `*Template, error` | Get template |
| `Create(ctx, opts)` | `*Template, error` | Create template |
| `Update(ctx, templateID, opts)` | `*Template, error` | Update template |
| `Delete(ctx, templateID)` | `error` | Delete template |

## Eval.Runs

| Method | Returns | Description |
|--------|---------|-------------|
| `List(ctx, agentID, limit, offset)` | `*RunListResponse, error` | List runs |
| `Get(ctx, runID)` | `*Run, error` | Get run details |
| `Delete(ctx, runID)` | `error` | Delete run |

## Error Types

All API errors are returned as typed errors for precise handling:

```go
switch err.(type) {
case *sonzai.AuthenticationError:  // 401
case *sonzai.PermissionDeniedError: // 403
case *sonzai.NotFoundError:         // 404
case *sonzai.BadRequestError:       // 400
case *sonzai.RateLimitError:        // 429
case *sonzai.InternalServerError:   // 500+
}
```
