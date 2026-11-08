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
)

/************************************************************************
* The system prompt grounding every answer in the user's real snapshot.
************************************************************************/
const chatSystemPrompt = `You are the gooava finance assistant — a friendly, plain-spoken helper built into the user's budgeting app.

Ground every answer in the financial snapshot JSON provided after this prompt. It contains the user's real data: today's date, bank balance (manual and/or linked checking total), accounts, income sources, expense definitions, unpaid bills (past-due flagged), 90-day spending totals per category, and top merchants. Never invent numbers; if the snapshot lacks what a question needs, say so and suggest where in the app to add it.

Answering style:
- Be concise: short paragraphs or tight bullet lists, no filler.
- Use plain language and USD amounts rounded to whole dollars unless cents matter.
- When asked whether a purchase is affordable or financially wise, reason from the numbers: money on hand (checking total when a bank is linked, otherwise the manual balance) minus unpaid bills due before the next income — i.e. what's actually left over. State the leftover figure you used.
- When asked for a savings plan or which bills to cut, name concrete items from the expense definitions and the 90-day category totals (largest discretionary categories first), with estimated monthly impact.
- Past-due bills are urgent — call them out before anything discretionary.
- When you make a forward-looking plan, add one short "not financial advice" note at the end. Otherwise no disclaimers.`

// only the most recent turns go to the model
const chatHistoryLimit = 20

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
* Sends a user message: persists it, builds the finance snapshot, calls OpenAI with
* the system prompt + recent history, persists the assistant reply, bumps the
* conversation, and auto-titles it from the first user message.
* Rate limited (20/min/IP) in the router — every send is a paid OpenAI call.
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

	snapshot, err := models.BuildFinanceSnapshot(authUser.Id)
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
		Content: chatSystemPrompt + "\n\nHere is the user's current financial snapshot as JSON:\n" + snapshot,
	})
	for _, m := range history {
		messages = append(messages, lib.ChatMessage{Role: m.Role, Content: m.Content})
	}

	reply, err := lib.ChatCompletion(messages)
	if err != nil {
		// the user message stays persisted — honest history; the client rolls its
		// optimistic bubble back and toasts this error
		httpResponse.Error = fmt.Sprintf("%v", err)
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
