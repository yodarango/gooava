import { useAppContext } from "../../../views/context/appContextProvider";
import {
  API_GET_CHAT_CONVERSATIONS,
  API_POST_CHAT_CONVERSATIONS,
  API_POST_CHAT_CONVERSATIONS_DELETE,
  API_GET_CHAT_MESSAGES,
  API_POST_CHAT_MESSAGES,
} from "@constants";
import { ArrowLeft, Trash2, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { Button, Loading, TextArea } from "@ds";
import { useEffect, useRef, useState } from "react";
import { useGet, usePost } from "@utils";

// styles
import "./ChatDrawer.css";

/*********************************************************************************************************
 * Conversation list — "New chat" on top, then every previous conversation (most recently
 * active first) with a trash action. Clicking a row opens the chat view.
 * ******************************************************************************************************
 */
const ConversationList = (props) => {
  const { refreshKey, onChanged, onOpen } = props;
  const { showToast } = useAppContext();

  const conversations = useGet({
    url: API_GET_CHAT_CONVERSATIONS,
    dependencies: [refreshKey],
  });

  const create = usePost({
    url: API_POST_CHAT_CONVERSATIONS,
    callback: (data) => {
      if (!data) return;
      if (onChanged) onChanged();
      onOpen(data); // data is the new conversation row
    },
  });

  const remove = usePost({
    url: API_POST_CHAT_CONVERSATIONS_DELETE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Conversation deleted", type: "success" });
      if (onChanged) onChanged();
    },
  });

  // surface request errors as toasts
  useEffect(() => {
    const message = create.error || remove.error;
    if (message) showToast({ message: String(message), type: "danger" });
  }, [create.error, remove.error]);

  const handleDelete = (conversation) => {
    if (!window.confirm(`Delete "${conversation.title}"?`)) return;
    remove.post({ id: conversation.id });
  };

  const list = Array.isArray(conversations.data) ? conversations.data : [];

  return (
    <div className='chat-drawer-x7q2__list-view'>
      <div className='chat-drawer-x7q2__new'>
        <Button
          primary
          onClick={() => create.post({})}
          isLoading={create.loading}
        >
          New chat
        </Button>
      </div>

      {conversations.loading ? (
        <p className='chat-drawer-x7q2__empty'>Loading…</p>
      ) : list.length === 0 ? (
        <p className='chat-drawer-x7q2__empty'>
          No conversations yet — ask about your balance, bills, or spending.
        </p>
      ) : (
        <ul className='chat-drawer-x7q2__list'>
          {list.map((conversation) => (
            <li key={conversation.id} className='chat-drawer-x7q2__row'>
              <button
                className='chat-drawer-x7q2__row-open'
                onClick={() => onOpen(conversation)}
                type='button'
              >
                <span className='chat-drawer-x7q2__row-title'>
                  {conversation.title}
                </span>
                <span className='chat-drawer-x7q2__row-date'>
                  {(conversation.updated_at || "").slice(0, 10)}
                </span>
              </button>
              <button
                className='chat-drawer-x7q2__row-delete'
                onClick={() => handleDelete(conversation)}
                aria-label={`Delete ${conversation.title}`}
                disabled={remove.loading}
                title='Delete conversation'
                type='button'
              >
                <MorphIcon icon={Trash2} label='Delete' size={15} />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

/*********************************************************************************************************
 * One conversation — message bubbles (user right, assistant left), auto-scroll to the
 * newest message, and the composer. The user's message appears optimistically and is
 * rolled back (with a toast) if the send fails.
 * ******************************************************************************************************
 */
const ChatView = (props) => {
  const { conversation, onBack } = props;
  const { showToast } = useAppContext();

  const [messages, setMessages] = useState([]);
  const [draft, setDraft] = useState("");
  const scrollRef = useRef(null);
  const pendingIdRef = useRef(null);

  // this component mounts per conversation (key={id} in the parent), so the history
  // fetch fires exactly once per open
  const history = useGet({
    url: API_GET_CHAT_MESSAGES + "?conversation_id=" + conversation.id,
  });

  // hydrate the local list once the history lands
  useEffect(() => {
    if (Array.isArray(history.data)) setMessages(history.data);
  }, [history.data]);

  const send = usePost({
    url: API_POST_CHAT_MESSAGES,
    callback: (data) => {
      if (!data) return;
      setMessages((prev) => [
        ...prev.filter((m) => m.id !== pendingIdRef.current),
        data.user_message,
        data.assistant_message,
      ]);
    },
  });

  // surface send errors as toasts + roll back the optimistic bubble
  useEffect(() => {
    if (!send.error) return;
    setMessages((prev) => prev.filter((m) => m.id !== pendingIdRef.current));
    showToast({ message: String(send.error), type: "danger" });
  }, [send.error]);

  // keep the newest message in view
  useEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages, send.loading]);

  const handleSend = () => {
    const content = draft.trim();
    if (!content || send.loading) return;
    setDraft("");
    const tempId = "local-" + Date.now();
    pendingIdRef.current = tempId;
    setMessages((prev) => [
      ...prev,
      { id: tempId, role: "user", content, created_at: "" },
    ]);
    send.post({ conversation_id: conversation.id, content });
  };

  return (
    <div className='chat-drawer-x7q2__chat-view'>
      <div className='chat-drawer-x7q2__chat-header'>
        <button
          className='chat-drawer-x7q2__back'
          onClick={onBack}
          aria-label='All conversations'
          type='button'
        >
          <MorphIcon icon={ArrowLeft} label='Back' size={20} />
        </button>
        <h3 className='chat-drawer-x7q2__chat-title'>{conversation.title}</h3>
      </div>

      <div className='chat-drawer-x7q2__messages' ref={scrollRef}>
        {history.loading ? (
          <div className='chat-drawer-x7q2__loading'>
            <Loading size={40} />
          </div>
        ) : messages.length === 0 ? (
          <p className='chat-drawer-x7q2__empty'>
            Ask things like “if I spend $40 this week, is that wise?” or “what bills
            could I cut?”
          </p>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              className={`chat-drawer-x7q2__bubble-row ${message.role}`}
            >
              <div className={`chat-drawer-x7q2__bubble ${message.role}`}>
                <p className='chat-drawer-x7q2__bubble-text'>{message.content}</p>
                {message.created_at && (
                  <span className='chat-drawer-x7q2__bubble-time'>
                    {message.created_at.slice(5, 16)}
                  </span>
                )}
              </div>
            </div>
          ))
        )}
        {send.loading && (
          <div className='chat-drawer-x7q2__bubble-row assistant'>
            <div className='chat-drawer-x7q2__bubble assistant'>
              <Loading size={24} />
            </div>
          </div>
        )}
      </div>

      {/* the ds TextArea overrides onKeyDown and fires onPressEnter on Shift+Enter
          (inverted from chat norms), so Enter-to-send lives on this wrapper — events
          bubble up from the textarea and preventDefault still blocks the newline.
          Shift+Enter is a no-op in v1 (the TextArea preventDefaults it). */}
      <div
        className='chat-drawer-x7q2__composer'
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSend();
          }
        }}
      >
        <TextArea
          minRows={1}
          maxRows={4}
          maxCharacters={4000}
          placeholder='Ask about your money…'
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
        />
        <Button
          primary
          onClick={handleSend}
          isLoading={send.loading}
          disabled={!draft.trim()}
        >
          Send
        </Button>
      </div>
    </div>
  );
};

/*********************************************************************************************************
 * The drawer shell — portal-mounted from the bottom nav while open, so it works over
 * every page. Shows the conversation list, or the open conversation.
 * ******************************************************************************************************
 */
export const ChatDrawer = (props) => {
  const { onClose } = props;
  const [activeConvo, setActiveConvo] = useState(null);
  const [refreshKey, setRefreshKey] = useState(0);

  // close on Escape + lock body scroll while open
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = "";
    };
  }, [onClose]);

  return (
    <div
      className='chat-drawer-x7q2'
      role='dialog'
      aria-modal='true'
      aria-label='Finance assistant'
    >
      <div className='chat-drawer-x7q2__backdrop' onClick={onClose} />

      <aside className='chat-drawer-x7q2__drawer'>
        <div className='chat-drawer-x7q2__header'>
          <h2 className='chat-drawer-x7q2__title'>Finance assistant</h2>
          <button
            className='chat-drawer-x7q2__close'
            onClick={onClose}
            aria-label='Close'
            type='button'
          >
            <MorphIcon icon={X} label='Close' size={20} />
          </button>
        </div>

        <div className='chat-drawer-x7q2__body'>
          {activeConvo ? (
            <ChatView
              key={activeConvo.id}
              conversation={activeConvo}
              onBack={() => {
                setActiveConvo(null);
                setRefreshKey((key) => key + 1); // reflect new titles/activity
              }}
            />
          ) : (
            <ConversationList
              refreshKey={refreshKey}
              onChanged={() => setRefreshKey((key) => key + 1)}
              onOpen={(conversation) => setActiveConvo(conversation)}
            />
          )}
        </div>
      </aside>
    </div>
  );
};
