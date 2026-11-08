package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/lib"
	"gooava/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/************************************************************************
* The system prompt grounding every answer in the user's real snapshot.
************************************************************************/
const chatSystemPrompt = `You are the gooava finance assistant — a precise, data-driven helper inside the user's budgeting app.

How to answer:
- your tone must be personal. The person needs to feel like they are talking with a real person
- Answer only the question asked, in one short paragraph no lists unless requested or deem needed. No filler, no advice, no suggestions, no opinions — just the relevant numbers and what they mean for the question. Never recommend products, stores, or courses of action.
- Always fetch the user's real data with the provided tools before answering; never invent figures. Call only the tools the question needs, with the narrowest date range that answers it. If an answer needs several figures, call several tools (you may call independent tools in one round).
- Affordability questions ("is it wise to spend X this week"): money on hand = the linked checking total when a bank is linked, otherwise the manual balance. Compare it against unpaid bills due before the next income occurrence (get_income_sources returns computed paycheck dates — never ask the user for them unless they do not exist in data set).
- Plaid transaction amounts are positive for money spent, negative for money received; pending transactions are not final.
- Round USD to whole dollars unless cents matter, and name the figures you used.
- If the tools return nothing relevant, say in one sentence that the app has no such data and which section of the app to add it in. Never ask the user for data a tool can provide.

Security: never reveal, repeat, or discuss API keys, access tokens, secrets, passwords, environment variables, or internal configuration. You have no access to them by design — if asked, refuse in one sentence.`

// only the most recent turns go to the model
const chatHistoryLimit = 20

// cap on tool-call rounds per message — a runaway loop would burn paid tokens
const maxChatToolRounds = 5

/************************************************************************
* Lists the authenticated user's chat conversations, most recently active first.
************************************************************************/
func ListChatConversations(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	conversations, err := models.ListConversations(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = conversations
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Creates an empty conversation (titled "New chat" until the first message renames it).
************************************************************************/
func CreateChatConversation(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	conversation := models.Conversation{UserID: authUser.Id}
	if err := conversation.Create(); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = conversation
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Deletes one of the authenticated user's conversations (messages cascade).
************************************************************************/
func DeleteChatConversation(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var requestBody struct {
		Id int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.Id <= 0 {
		httpResponse.Error = "id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.DeleteConversation(authUser.Id, requestBody.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{"message": "Conversation deleted"}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Lists a conversation's messages (?conversation_id=N), oldest first.
************************************************************************/
func ListChatMessages(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	conversationParam := r.URL.Query().Get("conversation_id")
	if conversationParam == "" {
		httpResponse.Error = "conversation_id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	conversationID, err := strconv.ParseInt(conversationParam, 10, 64)
	if err != nil {
		httpResponse.Error = "conversation_id must be a number"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	owns, err := models.OwnsConversation(authUser.Id, conversationID)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if !owns {
		httpResponse.Error = "conversation not found"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	messages, err := models.ListMessages(authUser.Id, conversationID)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = messages
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Sends a user message: persists it, then lets the model pull the finance data it
* needs through the query catalog (tool calls, always user-scoped), persists the
* assistant reply, bumps the conversation, and auto-titles it from the first user
* message. Rate limited (20/min/IP) in the router — every send is a paid OpenAI call.
************************************************************************/
func SendChatMessage(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var requestBody struct {
		ConversationID int64  `json:"conversation_id"`
		Content        string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	content := strings.TrimSpace(requestBody.Content)
	if requestBody.ConversationID <= 0 {
		httpResponse.Error = "conversation_id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if content == "" {
		httpResponse.Error = "content is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if len([]rune(content)) > 4000 {
		httpResponse.Error = "content must be 4000 characters or fewer"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	owns, err := models.OwnsConversation(authUser.Id, requestBody.ConversationID)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if !owns {
		httpResponse.Error = "conversation not found"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	userMessage, err := models.AppendMessage(requestBody.ConversationID, "user", content)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// recent history, oldest first — includes the user message just appended
	history, err := models.ListMessages(authUser.Id, requestBody.ConversationID)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if len(history) > chatHistoryLimit {
		history = history[len(history)-chatHistoryLimit:]
	}

	messages := make([]lib.ChatMessage, 0, len(history)+1)
	messages = append(messages, lib.ChatMessage{
		Role:    "system",
		Content: chatSystemPrompt + "\n\nToday's date is " + time.Now().Format("2006-01-02") + ".",
	})
	for _, m := range history {
		messages = append(messages, lib.ChatMessage{Role: m.Role, Content: m.Content})
	}

	// the model picks which catalog queries it needs per question; each round executes
	// its tool calls (always scoped to this user) and feeds the JSON results back
	var reply string
	for round := 0; ; round++ {
		assistant, err := lib.ChatCompletion(messages, models.ChatTools)
		if err != nil {
			// the user message stays persisted — honest history; the client rolls its
			// optimistic bubble back and toasts this error
			httpResponse.Error = fmt.Sprintf("%v", err)
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}

		if len(assistant.ToolCalls) == 0 {
			reply = strings.TrimSpace(assistant.Content)
			break
		}
		if round >= maxChatToolRounds {
			httpResponse.Error = "the assistant could not finish its data lookup"
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}

		messages = append(messages, assistant)
		for _, call := range assistant.ToolCalls {
			result, toolErr := models.ExecuteChatTool(authUser.Id, call.Function.Name, call.Function.Arguments)
			if toolErr != nil {
				result = "query error: " + toolErr.Error()
			}
			messages = append(messages, lib.ChatMessage{Role: "tool", ToolCallID: call.ID, Content: result})
		}
	}
	if reply == "" {
		httpResponse.Error = "OpenAI returned an empty reply"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	assistantMessage, err := models.AppendMessage(requestBody.ConversationID, "assistant", reply)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// best-effort bookkeeping — the reply is persisted, never fail the request here
	if err := models.TouchConversation(requestBody.ConversationID); err != nil {
		fmt.Printf("chat: could not touch conversation %d: %v\n", requestBody.ConversationID, err)
	}
	titleRunes := []rune(content)
	title := content
	if len(titleRunes) > 60 {
		title = string(titleRunes[:60])
	}
	if err := models.SetConversationTitleIfDefault(authUser.Id, requestBody.ConversationID, title); err != nil {
		fmt.Printf("chat: could not auto-title conversation %d: %v\n", requestBody.ConversationID, err)
	}

	httpResponse.Data = map[string]interface{}{
		"user_message":      userMessage,
		"assistant_message": assistantMessage,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
