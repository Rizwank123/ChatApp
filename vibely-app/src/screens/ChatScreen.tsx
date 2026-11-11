import { Ionicons } from "@expo/vector-icons";
import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  FlatList,
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import { api, apiConfig } from "../api/client";
import type { Conversation, Message, UUID } from "../api/types";
import { useApp } from "../state/AppContext";
import { colors } from "../theme/colors";

type Props = {
  initialConversationId?: UUID;
  initialParticipantId?: UUID;
  initialParticipantName?: string;
  onBack?: () => void;
};

// Instagram DM Notes mock
const DM_NOTES = [
  { id: "you", user: "Your note", note: "Share a thought...", isUser: true },
  { id: "1", user: "alexa", note: "Coding in coffee shop ☕" },
  { id: "2", user: "marcus", note: "Gym time 💪" },
  { id: "3", user: "elena", note: "New art dropping soon ✨" },
  { id: "4", user: "crypto", note: "Bull run loading 🚀" },
];

export function ChatScreen({
  initialConversationId,
  initialParticipantId,
  initialParticipantName,
  onBack,
}: Props = {}) {
  const { session, messages, appendMessage, showToast } = useApp();
  const socket = useRef<WebSocket | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [active, setActive] = useState<Conversation | undefined>();
  const [searchQuery, setSearchQuery] = useState("");
  const [text, setText] = useState("");
  const [showNewChatModal, setShowNewChatModal] = useState(false);
  const [targetUsername, setTargetUsername] = useState("");

  const activeMessages = active ? messages[active.id] ?? [] : [];
  const initializedParticipantRef = useRef<string | null>(null);

  const loadConversations = useCallback(async () => {
    if (!session) return;
    try {
      const result = await api.getConversations(session.token);
      const list = result.data ?? [];
      setConversations(list);

      setActive((prevActive) => {
        if (initialConversationId) {
          const found = list.find((c) => c.id === initialConversationId);
          if (found) return found;
        }
        return prevActive;
      });
    } catch {
      // Quiet fallback
    }
  }, [initialConversationId, session]);

  const loadMessages = useCallback(
    async (conversation: Conversation) => {
      if (!session) return;
      try {
        const result = await api.getMessages(session.token, conversation.id);
        (result.data ?? []).reverse().forEach(appendMessage);
      } catch {
        // Quiet fallback
      }
    },
    [appendMessage, session],
  );

  useEffect(() => {
    loadConversations().catch(() => undefined);
  }, [loadConversations]);

  // Handle direct navigation with initialParticipantId
  useEffect(() => {
    if (!session || !initialParticipantId) return;
    if (initializedParticipantRef.current === initialParticipantId) return;
    initializedParticipantRef.current = initialParticipantId;

    async function initDirectChat() {
      try {
        const result = await api.createConversation(
          session!.token,
          [initialParticipantId as UUID],
          "DIRECT",
          initialParticipantName || "",
        );
        setActive(result.data);
        const convsRes = await api.getConversations(session!.token);
        setConversations(convsRes.data ?? []);
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to open conversation";
        showToast({
          message: msg.includes("follow")
            ? "Direct message restricted: both users must follow each other."
            : msg,
          kind: "error",
        });
      }
    }

    initDirectChat();
  }, [initialParticipantId, initialParticipantName, session, showToast]);

  useEffect(() => {
    if (active) {
      loadMessages(active).catch(() => undefined);
    }
  }, [active?.id, loadMessages]);

  useEffect(() => {
    if (!session?.personnelId) return;
    const ws = new WebSocket(`${apiConfig.socketUrl}?id=${session.personnelId}`);
    socket.current = ws;
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === "error" || data.error_code === "FOLLOW_REQUIRED") {
          showToast({
            message: data.message || "Messaging restricted by follow relationship.",
            kind: "error",
          });
          return;
        }
        appendMessage(data as Message);
      } catch {
        // Ignore
      }
    };
    ws.onerror = () => {};
    return () => ws.close();
  }, [appendMessage, session?.personnelId, showToast]);

  async function startNewChatWithUsername() {
    if (!session || !targetUsername.trim()) return;
    try {
      // Find user by username
      const clean = targetUsername.trim().toLowerCase();
      const userRes = await api.findUserByUsername(session.token, clean);
      const targetUser = userRes.data;

      // Find personnel id or user id
      let targetPersonnelId = targetUser.id;
      try {
        const pRes = await api.findPersonnelByUserID(session.token, targetUser.id);
        if (pRes?.data?.id) {
          targetPersonnelId = pRes.data.id;
        }
      } catch {
        // Fallback
      }

      const convRes = await api.createConversation(
        session.token,
        [targetPersonnelId],
        "DIRECT",
        targetUser.user_name,
      );
      setTargetUsername("");
      setShowNewChatModal(false);
      await loadConversations();
      setActive(convRes.data);
      showToast({ message: `Chat opened with @${targetUser.user_name}`, kind: "success" });
    } catch (err) {
      showToast({
        message: err instanceof Error ? err.message : "User not found",
        kind: "error",
      });
    }
  }

  async function send(messageText?: string) {
    const toSend = messageText ?? text;
    if (!session || !active || !toSend.trim()) return;

    try {
      const result = await api.sendMessage(session.token, active.id, toSend.trim());
      appendMessage(result.data);
      if (!messageText) setText("");
    } catch (error) {
      const msg = error instanceof Error ? error.message : "Failed to send message";
      showToast({
        message: msg.includes("follow")
          ? "Direct message restricted: either user must follow the other."
          : msg,
        kind: "error",
      });
    }
  }

  const activeTitle =
    active?.name ||
    (initialParticipantName
      ? initialParticipantName
      : active
      ? `Chat with ${active.id.slice(0, 6)}`
      : "");

  // If inside an active conversation, show Instagram Message thread
  if (active) {
    return (
      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : undefined}
        style={styles.screen}
      >
        {/* Instagram DM Header */}
        <View style={styles.chatHeader}>
          <Pressable
            onPress={() => {
              setActive(undefined);
              onBack?.();
            }}
            style={styles.headerBackBtn}
            hitSlop={8}
          >
            <Ionicons name="arrow-back" size={24} color={colors.text} />
          </Pressable>

          <View style={styles.chatHeaderInfo}>
            <View style={styles.chatHeaderAvatar}>
              <Text style={styles.chatHeaderAvatarText}>
                {activeTitle.slice(0, 2).toUpperCase()}
              </Text>
            </View>
            <View>
              <Text style={styles.chatHeaderName} numberOfLines={1}>
                {activeTitle}
              </Text>
              <Text style={styles.chatHeaderStatus}>Active now</Text>
            </View>
          </View>

          <View style={styles.chatHeaderIcons}>
            <Pressable
              style={styles.headerIconBtn}
              onPress={() => showToast({ message: "Audio call feature coming soon", kind: "info" })}
              hitSlop={8}
            >
              <Ionicons name="call-outline" size={22} color={colors.text} />
            </Pressable>
            <Pressable
              style={styles.headerIconBtn}
              onPress={() => showToast({ message: "Video call feature coming soon", kind: "info" })}
              hitSlop={8}
            >
              <Ionicons name="videocam-outline" size={24} color={colors.text} />
            </Pressable>
          </View>
        </View>

        {/* Messages List */}
        <FlatList
          inverted={false}
          contentContainerStyle={styles.messagesList}
          data={activeMessages}
          keyExtractor={(item) => item.id}
          ListEmptyComponent={
            <View style={styles.emptyMessages}>
              <View style={styles.emptyAvatarBig}>
                <Text style={styles.emptyAvatarText}>
                  {activeTitle.slice(0, 2).toUpperCase()}
                </Text>
              </View>
              <Text style={styles.emptyTitle}>{activeTitle}</Text>
              <Text style={styles.emptySubtitle}>Instagram Direct</Text>
              <Text style={styles.emptyHint}>Say hi to start the conversation!</Text>
            </View>
          }
          renderItem={({ item }) => {
            const isMine =
              item.sender_id === session?.personnelId ||
              item.sender_id === session?.userName;
            return (
              <View
                style={[
                  styles.messageRow,
                  isMine ? styles.myMessageRow : styles.otherMessageRow,
                ]}
              >
                <View
                  style={[
                    styles.messageBubble,
                    isMine ? styles.myBubble : styles.otherBubble,
                  ]}
                >
                  <Text style={[styles.messageText, isMine ? styles.myText : styles.otherText]}>
                    {item.content}
                  </Text>
                </View>
              </View>
            );
          }}
        />

        {/* Instagram DM Input Bar */}
        <View style={styles.inputBar}>
          <Pressable
            style={styles.cameraIconBtn}
            onPress={() => showToast({ message: "Camera photo attachments coming soon", kind: "info" })}
            hitSlop={8}
          >
            <Ionicons name="camera" size={22} color="#ffffff" />
          </Pressable>

          <View style={styles.inputWrap}>
            <TextInput
              style={styles.messageInput}
              placeholder="Message..."
              placeholderTextColor={colors.textDim}
              value={text}
              onChangeText={setText}
              onSubmitEditing={() => send()}
            />
            {text.trim().length > 0 ? (
              <Pressable onPress={() => send()} style={styles.sendTextBtn} hitSlop={8}>
                <Text style={styles.sendText}>Send</Text>
              </Pressable>
            ) : (
              <View style={styles.inputTrailingIcons}>
                <Pressable
                  style={styles.trailingIcon}
                  onPress={() => showToast({ message: "Voice messaging coming soon", kind: "info" })}
                  hitSlop={6}
                >
                  <Ionicons name="mic-outline" size={20} color={colors.text} />
                </Pressable>
                <Pressable
                  style={styles.trailingIcon}
                  onPress={() => showToast({ message: "Select image from gallery", kind: "info" })}
                  hitSlop={6}
                >
                  <Ionicons name="image-outline" size={20} color={colors.text} />
                </Pressable>
                <Pressable
                  style={styles.trailingIcon}
                  onPress={() => send("❤️")}
                  hitSlop={6}
                >
                  <Ionicons name="heart" size={22} color={colors.heart} />
                </Pressable>
              </View>
            )}
          </View>
        </View>
      </KeyboardAvoidingView>
    );
  }

  // Inbox View (List of Direct Message threads)
  const filteredConversations = conversations.filter((c) =>
    (c.name || c.id).toLowerCase().includes(searchQuery.toLowerCase()),
  );

  return (
    <View style={styles.screen}>
      {/* Top Header */}
      <View style={styles.inboxHeader}>
        <View style={styles.inboxHeaderLeft}>
          <Text style={styles.inboxUsername}>{session?.userName || "Messages"}</Text>
          <Ionicons name="chevron-down" size={16} color={colors.text} />
        </View>

        <Pressable
          style={styles.newChatBtn}
          onPress={() => setShowNewChatModal((prev) => !prev)}
          hitSlop={8}
        >
          <Ionicons name="create-outline" size={24} color={colors.text} />
        </Pressable>
      </View>

      {/* New Chat Prompt Form */}
      {showNewChatModal && (
        <View style={styles.newChatBox}>
          <Text style={styles.newChatLabel}>New message to:</Text>
          <View style={styles.newChatInputRow}>
            <TextInput
              style={styles.newChatInput}
              placeholder="Enter username (e.g. alexa)..."
              placeholderTextColor={colors.textDim}
              value={targetUsername}
              onChangeText={setTargetUsername}
              autoCapitalize="none"
              autoFocus
            />
            <Pressable
              onPress={startNewChatWithUsername}
              disabled={!targetUsername.trim()}
              style={[
                styles.newChatSubmitBtn,
                !targetUsername.trim() && styles.disabledBtn,
              ]}
            >
              <Text style={styles.newChatSubmitText}>Chat</Text>
            </Pressable>
          </View>
        </View>
      )}

      {/* Search Bar */}
      <View style={styles.searchWrap}>
        <View style={styles.searchBar}>
          <Ionicons name="search" size={16} color={colors.textDim} />
          <TextInput
            placeholder="Search"
            placeholderTextColor={colors.textDim}
            style={styles.searchInput}
            value={searchQuery}
            onChangeText={setSearchQuery}
          />
        </View>
      </View>

      {/* Instagram Notes Tray at Top of DM */}
      <View style={styles.notesTray}>
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.notesScroll}>
          {DM_NOTES.map((item) => (
            <View key={item.id} style={styles.noteItem}>
              <View style={styles.noteBubble}>
                <Text style={styles.noteBubbleText} numberOfLines={2}>
                  {item.note}
                </Text>
              </View>
              <View style={styles.noteAvatarBorder}>
                <View style={styles.noteAvatar}>
                  <Text style={styles.noteAvatarText}>
                    {item.user.slice(0, 2).toUpperCase()}
                  </Text>
                </View>
                {item.isUser && (
                  <View style={styles.noteAddBadge}>
                    <Ionicons name="add" size={12} color="#ffffff" />
                  </View>
                )}
              </View>
              <Text style={styles.noteUserName} numberOfLines={1}>
                {item.user}
              </Text>
            </View>
          ))}
        </ScrollView>
      </View>

      {/* Section Header: Messages */}
      <View style={styles.messagesSectionHeader}>
        <Text style={styles.sectionTitle}>Messages</Text>
        <Text style={styles.requestsLink}>Requests (0)</Text>
      </View>

      {/* Conversation Thread List */}
      <FlatList
        data={filteredConversations}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.conversationsList}
        ListEmptyComponent={
          <View style={styles.emptyInbox}>
            <Ionicons name="chatbubbles-outline" size={54} color={colors.textDim} />
            <Text style={styles.emptyInboxTitle}>No messages yet</Text>
            <Text style={styles.emptyInboxSubtitle}>
              Tap the compose icon in the top right or search a friend to start chatting.
            </Text>
          </View>
        }
        renderItem={({ item }) => {
          const title = item.name || `Chat with @${item.id.slice(0, 6)}`;
          const lastMsg = messages[item.id]?.[messages[item.id]?.length - 1];
          const previewText = lastMsg ? lastMsg.content : "Active 2h ago";

          return (
            <Pressable
              style={styles.convRow}
              onPress={() => setActive(item)}
            >
              <View style={styles.convAvatarBorder}>
                <View style={styles.convAvatar}>
                  <Text style={styles.convAvatarText}>
                    {title.slice(0, 2).toUpperCase()}
                  </Text>
                </View>
                <View style={styles.onlineBadge} />
              </View>

              <View style={styles.convMeta}>
                <Text style={styles.convTitle}>{title}</Text>
                <Text style={styles.convPreview} numberOfLines={1}>
                  {previewText}
                </Text>
              </View>

              <Pressable
                style={styles.convCameraBtn}
                onPress={() => showToast({ message: "Camera feature", kind: "info" })}
                hitSlop={8}
              >
                <Ionicons name="camera-outline" size={22} color={colors.textDim} />
              </Pressable>
            </Pressable>
          );
        }}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  screen: {
    flex: 1,
    backgroundColor: colors.background,
  },
  inboxHeader: {
    height: 48,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    backgroundColor: colors.background,
  },
  inboxHeaderLeft: {
    flexDirection: "row",
    alignItems: "center",
    gap: 6,
  },
  inboxUsername: {
    color: colors.text,
    fontSize: 20,
    fontWeight: "700",
  },
  newChatBtn: {
    padding: 4,
  },
  newChatBox: {
    padding: 14,
    backgroundColor: colors.cardSubtle,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    gap: 8,
  },
  newChatLabel: {
    color: colors.textMuted,
    fontSize: 13,
    fontWeight: "600",
  },
  newChatInputRow: {
    flexDirection: "row",
    gap: 10,
  },
  newChatInput: {
    flex: 1,
    height: 40,
    backgroundColor: colors.background,
    borderRadius: 8,
    paddingHorizontal: 12,
    color: colors.text,
    fontSize: 14,
    borderWidth: 1,
    borderColor: colors.border,
  },
  newChatSubmitBtn: {
    paddingHorizontal: 18,
    height: 40,
    borderRadius: 8,
    backgroundColor: colors.primary,
    alignItems: "center",
    justifyContent: "center",
  },
  newChatSubmitText: {
    color: "#ffffff",
    fontSize: 14,
    fontWeight: "600",
  },
  disabledBtn: {
    opacity: 0.4,
  },
  searchWrap: {
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
  searchBar: {
    height: 38,
    borderRadius: 10,
    backgroundColor: colors.cardSubtle,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 12,
    gap: 8,
  },
  searchInput: {
    flex: 1,
    color: colors.text,
    fontSize: 14,
  },
  notesTray: {
    paddingVertical: 10,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
  },
  notesScroll: {
    paddingHorizontal: 16,
    gap: 16,
  },
  noteItem: {
    alignItems: "center",
    width: 74,
  },
  noteBubble: {
    backgroundColor: colors.cardElevated,
    borderRadius: 14,
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderWidth: 1,
    borderColor: colors.border,
    maxWidth: 74,
    marginBottom: -8,
    zIndex: 2,
    minHeight: 28,
    justifyContent: "center",
  },
  noteBubbleText: {
    color: colors.text,
    fontSize: 10,
    textAlign: "center",
  },
  noteAvatarBorder: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    position: "relative",
  },
  noteAvatar: {
    width: 52,
    height: 52,
    borderRadius: 26,
    backgroundColor: colors.cardElevated,
    alignItems: "center",
    justifyContent: "center",
  },
  noteAvatarText: {
    color: colors.text,
    fontSize: 13,
    fontWeight: "700",
  },
  noteAddBadge: {
    position: "absolute",
    bottom: 0,
    right: 0,
    width: 18,
    height: 18,
    borderRadius: 9,
    backgroundColor: colors.primary,
    borderWidth: 2,
    borderColor: colors.background,
    alignItems: "center",
    justifyContent: "center",
  },
  noteUserName: {
    color: colors.textMuted,
    fontSize: 11,
    marginTop: 4,
  },
  messagesSectionHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingTop: 12,
    paddingBottom: 6,
  },
  sectionTitle: {
    color: colors.text,
    fontSize: 15,
    fontWeight: "700",
  },
  requestsLink: {
    color: colors.primary,
    fontSize: 13,
    fontWeight: "600",
  },
  conversationsList: {
    paddingBottom: 24,
  },
  convRow: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 12,
  },
  convAvatarBorder: {
    position: "relative",
  },
  convAvatar: {
    width: 52,
    height: 52,
    borderRadius: 26,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
  },
  convAvatarText: {
    color: colors.text,
    fontSize: 16,
    fontWeight: "700",
  },
  onlineBadge: {
    position: "absolute",
    bottom: 2,
    right: 2,
    width: 12,
    height: 12,
    borderRadius: 6,
    backgroundColor: colors.emerald,
    borderWidth: 2,
    borderColor: colors.background,
  },
  convMeta: {
    flex: 1,
    gap: 2,
  },
  convTitle: {
    color: colors.text,
    fontSize: 14.5,
    fontWeight: "600",
  },
  convPreview: {
    color: colors.textDim,
    fontSize: 13,
  },
  convCameraBtn: {
    padding: 6,
  },
  emptyInbox: {
    padding: 40,
    alignItems: "center",
    justifyContent: "center",
    gap: 12,
    marginTop: 40,
  },
  emptyInboxTitle: {
    color: colors.text,
    fontSize: 18,
    fontWeight: "700",
  },
  emptyInboxSubtitle: {
    color: colors.textDim,
    fontSize: 13,
    textAlign: "center",
    maxWidth: 260,
    lineHeight: 18,
  },

  // Thread View Styles
  chatHeader: {
    height: 52,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 12,
    borderBottomWidth: 0.5,
    borderBottomColor: colors.border,
    backgroundColor: colors.background,
    gap: 8,
  },
  headerBackBtn: {
    padding: 4,
  },
  chatHeaderInfo: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
    flex: 1,
  },
  chatHeaderAvatar: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.cardSubtle,
    alignItems: "center",
    justifyContent: "center",
  },
  chatHeaderAvatarText: {
    color: colors.text,
    fontSize: 12,
    fontWeight: "700",
  },
  chatHeaderName: {
    color: colors.text,
    fontSize: 14,
    fontWeight: "700",
    maxWidth: 160,
  },
  chatHeaderStatus: {
    color: colors.emerald,
    fontSize: 11,
  },
  chatHeaderIcons: {
    flexDirection: "row",
    alignItems: "center",
    gap: 14,
  },
  headerIconBtn: {
    padding: 4,
  },
  messagesList: {
    paddingHorizontal: 14,
    paddingVertical: 16,
    gap: 8,
  },
  emptyMessages: {
    padding: 40,
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    marginTop: 40,
  },
  emptyAvatarBig: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: "center",
    justifyContent: "center",
    marginBottom: 6,
  },
  emptyAvatarText: {
    color: colors.text,
    fontSize: 22,
    fontWeight: "700",
  },
  emptyTitle: {
    color: colors.text,
    fontSize: 17,
    fontWeight: "700",
  },
  emptySubtitle: {
    color: colors.textDim,
    fontSize: 13,
  },
  emptyHint: {
    color: colors.textMuted,
    fontSize: 12,
    marginTop: 8,
  },
  messageRow: {
    flexDirection: "row",
    marginVertical: 2,
  },
  myMessageRow: {
    justifyContent: "flex-end",
  },
  otherMessageRow: {
    justifyContent: "flex-start",
  },
  messageBubble: {
    maxWidth: "75%",
    paddingHorizontal: 14,
    paddingVertical: 10,
    borderRadius: 18,
  },
  myBubble: {
    backgroundColor: colors.messageBubbleMine,
    borderBottomRightRadius: 4,
  },
  otherBubble: {
    backgroundColor: colors.messageBubbleOther,
    borderBottomLeftRadius: 4,
  },
  messageText: {
    fontSize: 14.5,
    lineHeight: 19,
  },
  myText: {
    color: "#ffffff",
  },
  otherText: {
    color: colors.text,
  },
  inputBar: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 12,
    paddingVertical: 8,
    gap: 10,
    borderTopWidth: 0.5,
    borderTopColor: colors.border,
    backgroundColor: colors.background,
  },
  cameraIconBtn: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: colors.primary,
    alignItems: "center",
    justifyContent: "center",
  },
  inputWrap: {
    flex: 1,
    minHeight: 40,
    borderRadius: 20,
    backgroundColor: colors.cardSubtle,
    borderWidth: 1,
    borderColor: colors.border,
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: 14,
  },
  messageInput: {
    flex: 1,
    color: colors.text,
    fontSize: 14,
    paddingVertical: 8,
  },
  sendTextBtn: {
    paddingLeft: 8,
  },
  sendText: {
    color: colors.primary,
    fontSize: 14,
    fontWeight: "700",
  },
  inputTrailingIcons: {
    flexDirection: "row",
    alignItems: "center",
    gap: 10,
  },
  trailingIcon: {
    padding: 2,
  },
});
