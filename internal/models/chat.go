package models

import (
	"database/sql"
	"fmt"
)

// Conversation is one AI chat thread owned by a user.
type Conversation struct {
	Id        int64  `json:"id"`
	UserID    uint   `json:"user_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Message is one chat turn inside a conversation.
type Message struct {
	Id             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	Role           string `json:"role"` // "user" | "assistant"
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

/**************************************************************************************
* Create inserts the conversation (title defaults to "New chat") and reloads the row
* so Id/timestamps come back populated.
**************************************************************************************/
func (c *Conversation) Create() error {
	if c.Title == "" {
		c.Title = "New chat"
	}

	result, err := ModelsRepo.DB.Conn.Exec(
		"INSERT INTO chat_conversations(user_id, title) VALUES(?,?)",
		c.UserID, c.Title,
	)
	if err != nil {
		return fmt.Errorf("could not create conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last conversation id: %w", err)
	}
	c.Id = id

	row := ModelsRepo.DB.Conn.QueryRow(
		`SELECT title, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s')
		 FROM chat_conversations WHERE id = ?`, id)
	if err := row.Scan(&c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return fmt.Errorf("could not reload conversation: %w", err)
	}

	return nil
}

/**************************************************************************************
* ListConversations returns the user's conversations, most recently active first.
**************************************************************************************/
func ListConversations(userID uint) ([]Conversation, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT id, title, DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s')
		 FROM chat_conversations WHERE user_id = ? ORDER BY updated_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list conversations: %w", err)
	}
	defer rows.Close()

	conversations := make([]Conversation, 0)
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.Id, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("could not scan conversation: %w", err)
		}
		c.UserID = userID
		conversations = append(conversations, c)
	}

	return conversations, nil
}

/**************************************************************************************
* DeleteConversation removes one of the user's conversations (messages cascade).
**************************************************************************************/
func DeleteConversation(userID uint, id int64) error {
	result, err := ModelsRepo.DB.Conn.Exec(
		"DELETE FROM chat_conversations WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("could not delete conversation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify conversation deletion: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("conversation not found")
	}

	return nil
}

/**************************************************************************************
* OwnsConversation reports whether the conversation exists and belongs to the user.
* Required because ListMessages cannot tell "not yours" apart from "empty".
**************************************************************************************/
func OwnsConversation(userID uint, conversationID int64) (bool, error) {
	var exists int
	err := ModelsRepo.DB.Conn.QueryRow(
		"SELECT 1 FROM chat_conversations WHERE id = ? AND user_id = ?",
		conversationID, userID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not verify conversation ownership: %w", err)
	}
	return true, nil
}

/**************************************************************************************
* AppendMessage adds a message to the conversation and returns the stored row.
* Callers MUST verify ownership first (OwnsConversation) — the insert itself is
* intentionally not user-scoped so it stays cheap on the hot path.
**************************************************************************************/
func AppendMessage(conversationID int64, role, content string) (Message, error) {
	result, err := ModelsRepo.DB.Conn.Exec(
		"INSERT INTO chat_messages(conversation_id, role, content) VALUES(?,?,?)",
		conversationID, role, content,
	)
	if err != nil {
		return Message{}, fmt.Errorf("could not append message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Message{}, fmt.Errorf("could not get last message id: %w", err)
	}

	message := Message{Id: id, ConversationID: conversationID, Role: role, Content: content}
	row := ModelsRepo.DB.Conn.QueryRow(
		`SELECT DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') FROM chat_messages WHERE id = ?`, id)
	if err := row.Scan(&message.CreatedAt); err != nil {
		return Message{}, fmt.Errorf("could not reload message: %w", err)
	}

	return message, nil
}

/**************************************************************************************
* ListMessages returns the conversation's messages, oldest first. The JOIN keeps other
* users' conversations unreachable (an unowned id yields an empty list).
**************************************************************************************/
func ListMessages(userID uint, conversationID int64) ([]Message, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT m.id, m.conversation_id, m.role, m.content, DATE_FORMAT(m.created_at, '%Y-%m-%d %H:%i:%s')
		 FROM chat_messages m
		 JOIN chat_conversations c ON c.id = m.conversation_id
		 WHERE m.conversation_id = ? AND c.user_id = ?
		 ORDER BY m.id`, conversationID, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list messages: %w", err)
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Id, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("could not scan message: %w", err)
		}
		messages = append(messages, m)
	}

	return messages, nil
}

/**************************************************************************************
* TouchConversation bumps updated_at so the thread sorts to the top of the list.
**************************************************************************************/
func TouchConversation(id int64) error {
	if _, err := ModelsRepo.DB.Conn.Exec(
		"UPDATE chat_conversations SET updated_at = NOW() WHERE id = ?", id); err != nil {
		return fmt.Errorf("could not touch conversation: %w", err)
	}
	return nil
}

/**************************************************************************************
* SetConversationTitleIfDefault renames the conversation only while it still carries
* the default title — later messages never clobber an auto-set title.
**************************************************************************************/
func SetConversationTitleIfDefault(userID uint, id int64, title string) error {
	if _, err := ModelsRepo.DB.Conn.Exec(
		"UPDATE chat_conversations SET title = ? WHERE id = ? AND user_id = ? AND title = 'New chat'",
		title, id, userID); err != nil {
		return fmt.Errorf("could not set conversation title: %w", err)
	}
	return nil
}
