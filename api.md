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
| `client.Knowledge` | `*KnowledgeResource` | Project-scoped knowledge base operations |
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
| `GetToken(ctx, agentID, opts)` | `*VoiceStreamToken, error` | Get token for voice live WebSocket |
| `Stream(ctx, token)` | `*VoiceStream, error` | Real-time duplex voice via Gemini Live |

## Agents.Wakeups

| Method | Returns | Description |
|--------|---------|-------------|
| `Schedule(ctx, agentID, opts)` | `*WakeupResponse, error` | Schedule a proactive check-in |
| `List(ctx, agentID, opts)` | `*WakeupsResponse, error` | List scheduled wakeups (filter by user, status, limit) |

## Agents.Generation

| Method | Returns | Description |
|--------|---------|-------------|
| `GenerateBio(ctx, agentID)` | `*GenerateBioResponse, error` | Generate agent biography |
| `GenerateCharacter(ctx, opts)` | `*GenerateCharacterResponse, error` | Generate full character profile |
| `GenerateSeedMemories(ctx, agentID, opts)` | `*SeedMemoriesResponse, error` | Generate seed memories |
| `SeedMemories(ctx, agentID, opts)` | `*SeedMemoriesResponse, error` | Plant seed memories |
| `RegenerateAvatar(ctx, agentID, opts)` | `*RegenerateAvatarResponse, error` | Regenerate agent avatar image |

## Knowledge

Project-scoped knowledge base operations accessed via `client.Knowledge`.

### Documents

| Method | Returns | Description |
|--------|---------|-------------|
| `ListDocuments(ctx, projectID, limit)` | `*KBDocumentListResponse, error` | List documents |
| `GetDocument(ctx, projectID, docID)` | `*KBDocument, error` | Get a document |
| `UploadDocument(ctx, projectID, opts)` | `*KBDocument, error` | Upload a document (multipart file) |
| `DeleteDocument(ctx, projectID, docID)` | `error` | Delete a document |

### Facts / Graph

| Method | Returns | Description |
|--------|---------|-------------|
| `InsertFacts(ctx, projectID, opts)` | `*InsertFactsResponse, error` | Insert entities and relationships |
| `ListNodes(ctx, projectID, nodeType, limit)` | `*KBNodeListResponse, error` | List graph nodes |
| `GetNode(ctx, projectID, nodeID, includeHistory)` | `*KBNodeDetailResponse, error` | Get node with edges |
| `DeleteNode(ctx, projectID, nodeID)` | `error` | Soft-delete a node |
| `GetNodeHistory(ctx, projectID, nodeID, limit)` | `*KBNodeHistoryResponse, error` | Node version history |
| `BulkUpdate(ctx, projectID, opts)` | `*KBBulkUpdateResponse, error` | Batch-update node properties |

### Search

| Method | Returns | Description |
|--------|---------|-------------|
| `Search(ctx, projectID, opts)` | `*KBSearchResponse, error` | BM25 search with graph traversal |

### Schemas

| Method | Returns | Description |
|--------|---------|-------------|
| `CreateSchema(ctx, projectID, opts)` | `*KBEntitySchema, error` | Create entity schema |
| `ListSchemas(ctx, projectID)` | `*KBSchemaListResponse, error` | List schemas |
| `UpdateSchema(ctx, projectID, schemaID, opts)` | `*KBEntitySchema, error` | Update schema |
| `DeleteSchema(ctx, projectID, schemaID)` | `error` | Delete schema |

### Analytics

| Method | Returns | Description |
|--------|---------|-------------|
| `CreateAnalyticsRule(ctx, projectID, opts)` | `*KBAnalyticsRule, error` | Create analytics rule |
| `ListAnalyticsRules(ctx, projectID)` | `*KBAnalyticsRuleListResponse, error` | List rules |
| `GetAnalyticsRule(ctx, projectID, ruleID)` | `*KBAnalyticsRule, error` | Get rule |
| `UpdateAnalyticsRule(ctx, projectID, ruleID, opts)` | `*KBAnalyticsRule, error` | Update rule |
| `DeleteAnalyticsRule(ctx, projectID, ruleID)` | `error` | Delete rule |
| `RunAnalyticsRule(ctx, projectID, ruleID)` | `error` | Trigger manual rule run |
| `GetRecommendations(ctx, projectID, ruleID, sourceID, limit)` | `*KBRecommendationsResponse, error` | Get recommendations |
| `GetTrends(ctx, projectID, nodeID)` | `*KBTrendsResponse, error` | Get trend aggregations |
| `GetTrendRankings(ctx, projectID, ruleID, type, window, limit)` | `*KBTrendRankingsResponse, error` | Get trend rankings |
| `GetConversions(ctx, projectID, ruleID, segment)` | `*KBConversionsResponse, error` | Get conversion stats |
| `RecordFeedback(ctx, projectID, opts)` | `error` | Record recommendation feedback |
| `GetStats(ctx, projectID)` | `*KBStats, error` | Knowledge base statistics |

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
